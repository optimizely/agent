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
	"os"

	"github.com/optimizely/go-sdk/pkg/config"

	"github.com/optimizely/go-sdk/pkg/client"
	cmap "github.com/orcaman/concurrent-map"
)

// OptlyCache implements the Cache interface backed by a concurrent map.
// The default OptlyClient lookup is based on supplied configuration via env variables.
type OptlyCache struct {
	defaultKey string
	loader     func(string) (*OptlyClient, error)
	optlyMap   cmap.ConcurrentMap
}

// NewCache returns a new implementation of OptlyCache interface backed by a concurrent map.
func NewCache() *OptlyCache {
	// TODO replace with actual configuration component
	// If Map lookup are expensive we can store a direct reference to the "default" client.
	sdkKey := os.Getenv("SDK_KEY")
	return &OptlyCache{
		defaultKey: sdkKey,
		optlyMap:   cmap.New(),
		loader:     initOptlyClient,
	}
}

// GetDefaultClient returns a default OptlyClient where the SDK Key is sourced via server configuration.
func (c *OptlyCache) GetDefaultClient() (*OptlyClient, error) {
	return c.GetClient(c.defaultKey)
}

// GetClient is used to fetch an instance of the OptlyClient when the SDK Key is explicitly supplied.
func (c *OptlyCache) GetClient(sdkKey string) (*OptlyClient, error) {
	val, ok := c.optlyMap.Get(sdkKey)
	if ok {
		return val.(*OptlyClient), nil
	}

	oc, err := c.loader(sdkKey)
	if err != nil {
		return &OptlyClient{}, err
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

	optimizelyFactory := &client.OptimizelyFactory{}
	configManager := config.NewPollingProjectConfigManager(sdkKey)
	forcedVariations := NewCMapExpOverridesStore()
	// TODO: Provide forcedVariations to client via option function or factory argument
	// For now, forced variations can be set, but will have no effect
	optimizelyClient, err := optimizelyFactory.Client(client.WithConfigManager(configManager))
	return &OptlyClient{optimizelyClient, configManager, forcedVariations}, err
}
