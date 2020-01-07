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
	"sync"

	"github.com/optimizely/sidedoor/config"

	"github.com/optimizely/go-sdk/pkg/client"
	sdkconfig "github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decision"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/rs/zerolog/log"
)

// OptlyCache implements the Cache interface backed by a concurrent map.
// The default OptlyClient lookup is based on supplied configuration via env variables.
type OptlyCache struct {
	loader   func(string, config.ProcessorConfig) (*OptlyClient, error)
	optlyMap cmap.ConcurrentMap
	ctx      context.Context
	wg       sync.WaitGroup
	conf     config.OptlyConfig
}

// NewCache returns a new implementation of OptlyCache interface backed by a concurrent map.
func NewCache(ctx context.Context, conf config.OptlyConfig) *OptlyCache {
	cache := &OptlyCache{
		ctx:      ctx,
		wg:       sync.WaitGroup{},
		loader:   initOptlyClient,
		optlyMap: cmap.New(),
		conf:     conf,
	}

	cache.init()
	return cache
}

func (c *OptlyCache) init() {
	for _, sdkKey := range c.conf.SDKKeys {
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

	oc, err := c.loader(sdkKey, c.conf.Processor)
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

// Wait for all optimizely clients to gracefully shutdown
func (c *OptlyCache) Wait() {
	c.wg.Wait()
}

func initOptlyClient(sdkKey string, conf config.ProcessorConfig) (*OptlyClient, error) {
	log.Info().Str("sdkKey", sdkKey).Msg("Loading Optimizely instance")
	configManager := sdkconfig.NewPollingProjectConfigManager(sdkKey)
	if _, err := configManager.GetConfig(); err != nil {
		return &OptlyClient{}, err
	}

	forcedVariations := decision.NewMapExperimentOverridesStore()
	optimizelyFactory := &client.OptimizelyFactory{}
	optimizelyClient, err := optimizelyFactory.Client(
		client.WithConfigManager(configManager),
		client.WithExperimentOverrides(forcedVariations),
		client.WithBatchEventProcessor(conf.BatchSize, conf.QueueSize, conf.FlushInterval),
	)

	return &OptlyClient{optimizelyClient, configManager, forcedVariations}, err
}
