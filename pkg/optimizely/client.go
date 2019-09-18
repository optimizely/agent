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
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/entities"
)

var once sync.Once
var optlyClient *client.OptimizelyClient

// OptlyClient wraps an instance of the OptimizelyClient to provide higher level functionality
type OptlyClient struct {
	*client.OptimizelyClient
}

// Client returns a ClientHolder reference providing higher level access to the OptimizelyClient
func Client() (client *OptlyClient) {
	return ClientWithOptimizelyClient(GetOptimizely())
}

// ClientWithOptimizelyClient returns a ClientHolder reference providing higher level access to the OptimizelyClient
func ClientWithOptimizelyClient(optimizelyClient *client.OptimizelyClient) (client *OptlyClient) {
	return &OptlyClient{
		optimizelyClient,
	}
}

// ListFeatures returns all available features
func (c *OptlyClient) ListFeatures() (features []entities.Feature, err error) {
	projectConfig, err := c.GetProjectConfig()
	if err != nil {
		return features, err
	}

	features = projectConfig.GetFeatureList()
	return features, err
}

// GetFeature returns the feature definition
func (c *OptlyClient) GetFeature(featureKey string) (feature entities.Feature, err error) {
	projectConfig, err := c.GetProjectConfig()
	if err != nil {
		return feature, err
	}

	return projectConfig.GetFeatureByKey(featureKey)
}

// GetOptimizely returns an instance of OptimizelyClient
// TODO Support multiple SDK keys
func GetOptimizely() *client.OptimizelyClient {

	// Short circuit for testing
	if optlyClient != nil {
		return optlyClient
	}

	// TODO handle failure to prevent deadlocks.
	once.Do(func() { // <-- atomic, does not allow repeating
		sdkKey := os.Getenv("SDK_KEY")
		sublogger := log.With().Str("sdkKey", sdkKey).Logger()
		sublogger.Info().Msg("Fetching new OptimizelyClient")

		optimizelyFactory := &client.OptimizelyFactory{
			// TODO parameterize
			SDKKey: sdkKey,
		}

		var err error
		optlyClient, err = optimizelyFactory.StaticClient()

		if err != nil {
			sublogger.Error().Err(err).Msg("Initializing OptimizelyClient")
			return
		}
	})

	return optlyClient
}
