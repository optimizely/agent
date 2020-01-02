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
	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decision"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const dispatcherMetricsPrefix = "dispatcher."

var metrics *Metrics

// OptlyCache implements the Cache interface backed by a concurrent map.
// The default OptlyClient lookup is based on supplied configuration via env variables.
type OptlyCache struct {
	loader   func(string) (*OptlyClient, error)
	optlyMap cmap.ConcurrentMap
}

// NewCache returns a new implementation of OptlyCache interface backed by a concurrent map.
func NewCache() *OptlyCache {
	cache := &OptlyCache{
		optlyMap: cmap.New(),
		loader:   initOptlyClient,
	}

	cache.init()
	return cache
}

func (c *OptlyCache) init() {
	sdkKeys := viper.GetStringSlice("optimizely.sdkKeys")
	metrics = NewMetrics(dispatcherMetricsPrefix)
	for _, sdkKey := range sdkKeys {
		if _, err := c.GetClient(sdkKey); err != nil {
			log.Warn().Str("sdkKey", sdkKey).Msg("Failed to initialize Opimizely Client.")
		}
	}

	//pollingFrequency := viper.GetDuration("metrics.pollingfreqency")
	//go func() {
	//
	//	t := time.NewTicker(pollingFrequency)
	//	for {
	//		select {
	//		case <-t.C:
	//			metrics.SetMetrics(c.GetEventDispatcherMetrics())
	//		}
	//	}
	//}()
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
		return oc, err
	}

	// If we didn't "set" the key in this method execution then it was set in another thread.
	// Recursively lookuping up the SDK key "should" only happen once.
	return c.GetClient(sdkKey)
}

func initOptlyClient(sdkKey string) (*OptlyClient, error) {
	log.Info().Str("sdkKey", sdkKey).Msg("Loading Optimizely instance")
	configManager := config.NewPollingProjectConfigManager(sdkKey)
	if _, err := configManager.GetConfig(); err != nil {
		return &OptlyClient{}, err
	}

	forcedVariations := decision.NewMapExperimentOverridesStore()
	optimizelyFactory := &client.OptimizelyFactory{}
	optimizelyClient, err := optimizelyFactory.Client(
		client.WithConfigManager(configManager),
		client.WithExperimentOverrides(forcedVariations),
		client.WithMetrics(metrics),
	)
	return &OptlyClient{optimizelyClient, configManager, forcedVariations}, err
}

// GetEventDispatcherMetrics aggregates metrics from all clients
//func (c *OptlyCache) GetEventDispatcherMetrics() *event.DefaultMetrics {
//
//	clientsMetrics := &event.DefaultMetrics{}
//	for _, client := range c.optlyMap.Items() {
//
//		if client, ok := client.(*OptlyClient); ok {
//			if eventProcessor, goodEP := client.EventProcessor.(*event.BatchEventProcessor); goodEP {
//				if metric, goodMetric := eventProcessor.EventDispatcher.GetMetrics().(*event.DefaultMetrics); goodMetric {
//					clientsMetrics.Add(metric)
//				}
//
//			}
//		}
//	}
//	return clientsMetrics
//}
