/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
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
	"errors"
	"regexp"
	"strings"
	"sync"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/go-sdk/pkg/client"
	sdkconfig "github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/event"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/rs/zerolog/log"
)

// OptlyCache implements the Cache interface backed by a concurrent map.
// The default OptlyClient lookup is based on supplied configuration via env variables.
type OptlyCache struct {
	loader   func(string) (*OptlyClient, error)
	optlyMap cmap.ConcurrentMap
	ctx      context.Context
	wg       sync.WaitGroup
}

// NewCache returns a new implementation of OptlyCache interface backed by a concurrent map.
func NewCache(ctx context.Context, conf config.ClientConfig, metricsRegistry *MetricsRegistry) *OptlyCache {

	// TODO is there a cleaner way to handle this translation???
	cmLoader := func(sdkkey string, options ...sdkconfig.OptionFunc) SyncedConfigManager {
		return sdkconfig.NewPollingProjectConfigManager(sdkkey, options...)
	}

	cache := &OptlyCache{
		ctx:      ctx,
		wg:       sync.WaitGroup{},
		loader:   defaultLoader(conf, metricsRegistry, cmLoader, event.NewBatchEventProcessor),
		optlyMap: cmap.New(),
	}

	return cache
}

// Init takes a slice of sdkKeys to warm the cache upon startup
func (c *OptlyCache) Init(sdkKeys []string) {
	for _, sdkKey := range sdkKeys {
		if _, err := c.GetClient(sdkKey); err != nil {
			log.Warn().Str("sdkKey", sdkKey).Msg("Failed to initialize Optimizely Client.")
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
	conf config.ClientConfig,
	metricsRegistry *MetricsRegistry,
	pcFactory func(sdkKey string, options ...sdkconfig.OptionFunc) SyncedConfigManager,
	bpFactory func(options ...event.BPOptionConfig) *event.BatchEventProcessor) func(clientKey string) (*OptlyClient, error) {
	validator := regexValidator(conf.SdkKeyRegex)

	return func(clientKey string) (*OptlyClient, error) {
		var sdkKey string
		var datafileAccessToken string
		var configManager SyncedConfigManager

		if !validator(clientKey) {
			log.Warn().Msgf("failed to validate sdk key: %q", sdkKey)
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

		log.Info().Str("sdkKey", sdkKey).Msg("Loading Optimizely instance")

		if datafileAccessToken != "" {
			configManager = pcFactory(
				sdkKey,
				sdkconfig.WithPollingInterval(conf.PollingInterval),
				sdkconfig.WithDatafileURLTemplate(conf.DatafileURLTemplate),
				sdkconfig.WithDatafileAccessToken(datafileAccessToken),
			)
		} else {
			configManager = pcFactory(
				sdkKey,
				sdkconfig.WithPollingInterval(conf.PollingInterval),
				sdkconfig.WithDatafileURLTemplate(conf.DatafileURLTemplate),
			)
		}

		if _, err := configManager.GetConfig(); err != nil {
			return &OptlyClient{}, err
		}

		q := event.NewInMemoryQueue(conf.QueueSize)
		ep := bpFactory(
			event.WithSDKKey(sdkKey),
			event.WithQueueSize(conf.QueueSize),
			event.WithBatchSize(conf.BatchSize),
			event.WithEventEndPoint(conf.EventURL),
			event.WithFlushInterval(conf.FlushInterval),
			event.WithQueue(q),
			event.WithEventDispatcherMetrics(metricsRegistry),
		)

		forcedVariations := decision.NewMapExperimentOverridesStore()
		optimizelyFactory := &client.OptimizelyFactory{SDKKey: sdkKey}
		optimizelyClient, err := optimizelyFactory.Client(
			client.WithConfigManager(configManager),
			client.WithExperimentOverrides(forcedVariations),
			client.WithEventProcessor(ep),
		)

		return &OptlyClient{optimizelyClient, configManager, forcedVariations}, err
	}
}
