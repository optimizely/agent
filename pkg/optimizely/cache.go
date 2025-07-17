/****************************************************************************
 * Copyright 2019,2022-2024 Optimizely, Inc. and contributors               *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package optimizely wraps the Optimizely SDK
package optimizely

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/syncer"
	"github.com/optimizely/agent/plugins/odpcache"
	"github.com/optimizely/agent/plugins/userprofileservice"
	odpCachePkg "github.com/optimizely/go-sdk/v2/pkg/cache"
	"github.com/optimizely/go-sdk/v2/pkg/client"
	cmab "github.com/optimizely/go-sdk/v2/pkg/cmab"
	sdkconfig "github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/decision"
	"github.com/optimizely/go-sdk/v2/pkg/event"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
	"github.com/optimizely/go-sdk/v2/pkg/odp"
	odpEventPkg "github.com/optimizely/go-sdk/v2/pkg/odp/event"
	odpSegmentPkg "github.com/optimizely/go-sdk/v2/pkg/odp/segment"
	"github.com/optimizely/go-sdk/v2/pkg/tracing"
	"github.com/optimizely/go-sdk/v2/pkg/utils"
)

// User plugin strings required for internal usage
const (
	userProfileServicePlugin = "UserProfileService"
	odpCachePlugin           = "ODP Cache"
)

// OptlyCache implements the Cache interface backed by a concurrent map.
// The default OptlyClient lookup is based on supplied configuration via env variables.
type OptlyCache struct {
	loader                func(string) (*OptlyClient, error)
	optlyMap              cmap.ConcurrentMap
	userProfileServiceMap cmap.ConcurrentMap
	odpCacheMap           cmap.ConcurrentMap
	ctx                   context.Context
	wg                    sync.WaitGroup
}

// NewCache returns a new implementation of OptlyCache interface backed by a concurrent map.
func NewCache(ctx context.Context, conf config.AgentConfig, metricsRegistry *MetricsRegistry, tracer trace.Tracer) *OptlyCache {

	// TODO is there a cleaner way to handle this translation???
	cmLoader := func(sdkkey string, options ...sdkconfig.OptionFunc) SyncedConfigManager {
		return sdkconfig.NewPollingProjectConfigManager(sdkkey, options...)
	}

	userProfileServiceMap := cmap.New()
	odpCacheMap := cmap.New()
	cache := &OptlyCache{
		ctx:                   ctx,
		wg:                    sync.WaitGroup{},
		loader:                defaultLoader(conf, metricsRegistry, tracer, userProfileServiceMap, odpCacheMap, cmLoader, event.NewBatchEventProcessor),
		optlyMap:              cmap.New(),
		userProfileServiceMap: userProfileServiceMap,
		odpCacheMap:           odpCacheMap,
	}

	return cache
}

// Init takes a slice of sdkKeys to warm the cache upon startup
func (c *OptlyCache) Init(sdkKeys []string) {
	for _, sdkKey := range sdkKeys {
		if _, err := c.GetClient(sdkKey); err != nil {
			message := "Failed to initialize Optimizely Client."
			if ShouldIncludeSDKKey {
				log.Warn().Str("sdkKey", sdkKey).Msg(message)
				continue
			}
			log.Warn().Msg(message)
		}
	}
}

// GetClient is used to fetch an instance of the OptlyClient when the SDK Key is explicitly supplied.
func (c *OptlyCache) GetClient(sdkKey string) (*OptlyClient, error) {
	val, ok := c.optlyMap.Get(sdkKey)
	if ok {
		return val.(*OptlyClient), nil
	}

	oc, err := c.loader(sdkKey)
	if err != nil {
		return oc, err
	}

	set := c.optlyMap.SetIfAbsent(sdkKey, oc)
	if set {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			<-c.ctx.Done()
			oc.Close()
		}()
		return oc, err
	}

	// Clean-up to not leave any lingering un-unused goroutines
	go oc.Close()

	// If we didn't "set" the key in this method execution then it was set in another thread.
	// Recursively lookuping up the SDK key "should" only happen once.
	return c.GetClient(sdkKey)
}

// UpdateConfigs is used to update config for all clients corresponding to a particular SDK key.
func (c *OptlyCache) UpdateConfigs(sdkKey string) {
	for clientInfo := range c.optlyMap.IterBuffered() {
		if strings.HasPrefix(clientInfo.Key, sdkKey) {
			optlyClient, ok := clientInfo.Val.(*OptlyClient)
			if !ok {
				log.Error().Msgf("Value not instance of OptlyClient.")
			}
			optlyClient.UpdateConfig()
		}
	}
}

// SetUserProfileService sets userProfileService to be used for the given sdkKey
func (c *OptlyCache) SetUserProfileService(sdkKey, userProfileService string) {
	c.userProfileServiceMap.SetIfAbsent(sdkKey, userProfileService)
}

// SetODPCache sets odpCache to be used for the given sdkKey
func (c *OptlyCache) SetODPCache(sdkKey, odpCache string) {
	c.odpCacheMap.SetIfAbsent(sdkKey, odpCache)
}

// Wait for all optimizely clients to gracefully shutdown
func (c *OptlyCache) Wait() {
	c.wg.Wait()
}

// ErrValidationFailure is returned when the provided SDK key fails initial validation
var ErrValidationFailure = errors.New("sdkKey failed validation")

func regexValidator(sdkKeyRegex string) func(string) bool {
	r, err := regexp.Compile(sdkKeyRegex)
	if err != nil {
		log.Fatal().Err(err).Msgf("invalid sdkKeyRegex configuration")
	}

	return r.MatchString
}

func defaultLoader(
	agentConf config.AgentConfig,
	metricsRegistry *MetricsRegistry,
	tracer trace.Tracer,
	userProfileServiceMap cmap.ConcurrentMap,
	odpCacheMap cmap.ConcurrentMap,
	pcFactory func(sdkKey string, options ...sdkconfig.OptionFunc) SyncedConfigManager,
	bpFactory func(options ...event.BPOptionConfig) *event.BatchEventProcessor) func(clientKey string) (*OptlyClient, error) {
	clientConf := agentConf.Client
	validator := regexValidator(clientConf.SdkKeyRegex)

	return func(clientKey string) (*OptlyClient, error) {
		var sdkKey string
		var datafileAccessToken string
		var configManager SyncedConfigManager

		if !validator(clientKey) {
			message := "failed to validate sdk key"
			if ShouldIncludeSDKKey {
				log.Warn().Msgf("%v: %q", message, sdkKey)
			} else {
				log.Warn().Msg(message)
			}
			return &OptlyClient{}, ErrValidationFailure
		}

		clientKeySplit := strings.Split(clientKey, ":")

		// If there is a : then it is an authenticated datafile.
		// First part is the sdkKey.
		// Second part is the datafileAccessToken
		sdkKey = clientKeySplit[0]
		if len(clientKeySplit) == 2 {
			datafileAccessToken = clientKeySplit[1]
		}

		message := "Loading Optimizely instance"
		if ShouldIncludeSDKKey {
			log.Info().Str("sdkKey", sdkKey).Msg(message)
		} else {
			log.Info().Msg(message)
		}

		if datafileAccessToken != "" {
			configManager = pcFactory(
				sdkKey,
				sdkconfig.WithPollingInterval(clientConf.PollingInterval),
				sdkconfig.WithDatafileURLTemplate(clientConf.DatafileURLTemplate),
				sdkconfig.WithDatafileAccessToken(datafileAccessToken),
			)
		} else {
			configManager = pcFactory(
				sdkKey,
				sdkconfig.WithPollingInterval(clientConf.PollingInterval),
				sdkconfig.WithDatafileURLTemplate(clientConf.DatafileURLTemplate),
			)
		}

		if _, err := configManager.GetConfig(); err != nil {
			return &OptlyClient{}, err
		}

		q := event.NewInMemoryQueue(clientConf.QueueSize)
		ep := bpFactory(
			event.WithSDKKey(sdkKey),
			event.WithQueueSize(clientConf.QueueSize),
			event.WithBatchSize(clientConf.BatchSize),
			event.WithEventEndPoint(clientConf.EventURL),
			event.WithFlushInterval(clientConf.FlushInterval),
			event.WithQueue(q),
			event.WithEventDispatcherMetrics(metricsRegistry),
		)

		forcedVariations := decision.NewMapExperimentOverridesStore()
		optimizelyFactory := &client.OptimizelyFactory{SDKKey: sdkKey}

		clientOptions := []client.OptionFunc{
			client.WithConfigManager(configManager),
			client.WithExperimentOverrides(forcedVariations),
			client.WithEventProcessor(ep),
			client.WithOdpDisabled(clientConf.ODP.Disable),
			client.WithTracer(tracing.NewOtelTracer(tracer)),
		}

		if agentConf.Synchronization.Notification.Enable {
			syncedNC, err := syncer.NewSyncedNotificationCenter(context.Background(), sdkKey, agentConf.Synchronization)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to create SyncedNotificationCenter, reason: %s", err.Error())
			} else {
				clientOptions = append(clientOptions, client.WithNotificationCenter(syncedNC))
			}
		}

		var clientUserProfileService decision.UserProfileService
		var rawUPS = getServiceWithType(userProfileServicePlugin, sdkKey, userProfileServiceMap, clientConf.UserProfileService)
		// Check if ups was provided by user
		if rawUPS != nil {
			// convert ups to UserProfileService interface
			if convertedUPS, ok := rawUPS.(decision.UserProfileService); ok && convertedUPS != nil {
				clientUserProfileService = convertedUPS
				clientOptions = append(clientOptions, client.WithUserProfileService(clientUserProfileService))
			}
		}

		var clientODPCache odpCachePkg.Cache
		var rawODPCache = getServiceWithType(odpCachePlugin, sdkKey, odpCacheMap, clientConf.ODP.SegmentsCache)
		// Check if odp cache was provided by user
		if rawODPCache != nil {
			// convert odpCache to Cache interface
			if convertedODPCache, ok := rawODPCache.(odpCachePkg.Cache); ok && convertedODPCache != nil {
				clientODPCache = convertedODPCache
			}
		}

		// Create segment manager with odpConfig and custom cache
		segmentManager := odpSegmentPkg.NewSegmentManager(
			sdkKey,
			odpSegmentPkg.WithAPIManager(
				odpSegmentPkg.NewSegmentAPIManager(sdkKey, utils.NewHTTPRequester(logging.GetLogger(sdkKey, "SegmentAPIManager"), utils.Timeout(clientConf.ODP.SegmentsRequestTimeout))),
			),
			odpSegmentPkg.WithSegmentsCache(clientODPCache),
		)

		// Create event manager with odpConfig
		eventManager := odpEventPkg.NewBatchEventManager(
			odpEventPkg.WithAPIManager(
				odpEventPkg.NewEventAPIManager(
					sdkKey, utils.NewHTTPRequester(logging.GetLogger(sdkKey, "EventAPIManager"), utils.Timeout(clientConf.ODP.EventsRequestTimeout)),
				),
			),
			odpEventPkg.WithFlushInterval(clientConf.ODP.EventsFlushInterval),
		)

		// Create odp manager with custom segment and event manager
		odpManager := odp.NewOdpManager(
			sdkKey,
			clientConf.ODP.Disable,
			odp.WithSegmentManager(segmentManager),
			odp.WithEventManager(eventManager),
		)
		clientOptions = append(clientOptions, client.WithOdpManager(odpManager))

		// Parse CMAB cache configuration
		cacheSize := 1000            // default
		cacheTTL := 30 * time.Minute // default

		if cacheConfig, ok := clientConf.CMAB.Cache["size"].(int); ok {
			cacheSize = cacheConfig
		}

		if cacheTTLStr, ok := clientConf.CMAB.Cache["ttl"].(string); ok {
			if parsedTTL, err := time.ParseDuration(cacheTTLStr); err == nil {
				cacheTTL = parsedTTL
			} else {
				log.Warn().Err(err).Msgf("Failed to parse CMAB cache TTL: %s, using default", cacheTTLStr)
			}
		}

		// Parse retry configuration
		retryConfig := &cmab.RetryConfig{
			MaxRetries:        3,
			InitialBackoff:    100 * time.Millisecond,
			MaxBackoff:        10 * time.Second,
			BackoffMultiplier: 2.0,
		}

		if maxRetries, ok := clientConf.CMAB.RetryConfig["maxRetries"].(int); ok {
			retryConfig.MaxRetries = maxRetries
		}

		if initialBackoffStr, ok := clientConf.CMAB.RetryConfig["initialBackoff"].(string); ok {
			if parsedBackoff, err := time.ParseDuration(initialBackoffStr); err == nil {
				retryConfig.InitialBackoff = parsedBackoff
			}
		}

		if maxBackoffStr, ok := clientConf.CMAB.RetryConfig["maxBackoff"].(string); ok {
			if parsedBackoff, err := time.ParseDuration(maxBackoffStr); err == nil {
				retryConfig.MaxBackoff = parsedBackoff
			}
		}

		if multiplier, ok := clientConf.CMAB.RetryConfig["backoffMultiplier"].(float64); ok {
			retryConfig.BackoffMultiplier = multiplier
		}

		// Create CMAB client and service
		cmabClient := cmab.NewDefaultCmabClient(cmab.ClientOptions{
			HTTPClient: &http.Client{
				Timeout: clientConf.CMAB.RequestTimeout,
			},
			RetryConfig: retryConfig,
			Logger:      logging.GetLogger(sdkKey, "CmabClient"),
		})

		cmabService := cmab.NewDefaultCmabService(cmab.ServiceOptions{
			Logger:     logging.GetLogger(sdkKey, "CmabService"),
			CmabCache:  odpCachePkg.NewLRUCache(cacheSize, cacheTTL),
			CmabClient: cmabClient,
		})

		clientOptions = append(clientOptions, client.WithCmabService(cmabService))

		optimizelyClient, err := optimizelyFactory.Client(
			clientOptions...,
		)
		return &OptlyClient{optimizelyClient, configManager, forcedVariations, clientUserProfileService, clientODPCache}, err
	}
}

func getServiceWithType(serviceType, sdkKey string, serviceMap cmap.ConcurrentMap, serviceConf map[string]interface{}) interface{} {

	intializeServiceWithName := func(serviceName string) interface{} {
		if clientConfigMap, ok := serviceConf["services"].(map[string]interface{}); ok {
			if serviceConfig, ok := clientConfigMap[serviceName].(map[string]interface{}); ok {
				// Check if any such service was added using `Add` method
				var serviceInstance interface{}
				switch serviceType {
				case userProfileServicePlugin:
					if upsCreator, ok := userprofileservice.Creators[serviceName]; ok {
						serviceInstance = upsCreator()
					}
				case odpCachePlugin:
					if odpCreator, ok := odpcache.Creators[serviceName]; ok {
						serviceInstance = odpCreator()
					}
				default:
				}

				if serviceInstance != nil {
					// Trying to map service from client config to struct
					if serviceConfig, err := json.Marshal(serviceConfig); err != nil {
						log.Warn().Err(err).Msgf(`Error marshaling %s config: %q`, serviceType, serviceName)
					} else if err := json.Unmarshal(serviceConfig, serviceInstance); err != nil {
						log.Warn().Err(err).Msgf(`Error unmarshalling %s config: %q`, serviceType, serviceName)
					} else {
						log.Info().Msgf(`%s of type: %q created for sdkKey: %q`, serviceType, serviceName, sdkKey)
						return serviceInstance
					}
				}
				return nil
			}
		}
		return nil
	}

	// Check if service name was provided in the request headers
	if service, ok := serviceMap.Get(sdkKey); ok {
		if serviceNameStr, ok := service.(string); ok && serviceNameStr != "" {
			return intializeServiceWithName(serviceNameStr)
		}
	}

	// Check if any default service was provided and if it exists in client config
	if defaultServiceName, isAvailable := serviceConf["default"].(string); isAvailable && defaultServiceName != "" {
		return intializeServiceWithName(defaultServiceName)
	}
	return nil
}
