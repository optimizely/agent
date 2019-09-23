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

	optimizelyclient "github.com/optimizely/go-sdk/optimizely/client"
	optimizelyconfig "github.com/optimizely/go-sdk/optimizely/config"
	optimizelyutils "github.com/optimizely/go-sdk/optimizely/utils"
	"github.com/optimizely/go-sdk/optimizely/entities"
)

var once sync.Once
var optlyClient *optimizelyclient.OptimizelyClient
var configManager *optimizelyconfig.PollingProjectConfigManager

// OptlyClient wraps an instance of the OptimizelyClient to provide higher level functionality
type OptlyClient struct {
	*optimizelyclient.OptimizelyClient
}

// NewClient returns an OptlyClient reference providing higher level access to the OptimizelyClient
func NewClient() (client *OptlyClient) {
	return &OptlyClient{
		GetOptimizely(),
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

// SetConfig uses config manager to sync and set project config
func SetConfig() {
	configManager.SyncConfig([]byte{})
}

// GetOptimizely returns an instance of OptimizelyClient
// TODO Support multiple SDK keys
func GetOptimizely() *optimizelyclient.OptimizelyClient {

	// Short circuit for testing
	if optlyClient != nil {
		return optlyClient
	}

	// TODO handle failure to prevent deadlocks.
	once.Do(func() { // <-- atomic, does not allow repeating
		sdkKey := os.Getenv("SDK_KEY")
		sublogger := log.With().Str("sdkKey", sdkKey).Logger()
		sublogger.Info().Msg("Fetching new OptimizelyClient")

		optimizelyFactory := &optimizelyclient.OptimizelyFactory{}

		var err error

		// TODO introduce another initialization option on PollingProjectConfigManager to take care of creating execution context
		execCtx := optimizelyutils.NewCancelableExecutionCtx()
		configManager := optimizelyconfig.NewPollingProjectConfigManager(execCtx, sdkKey)
		optlyClient, err = optimizelyFactory.ClientWithOptions(optimizelyclient.Options{
			ProjectConfigManager: configManager,
		})

		if err != nil {
			sublogger.Error().Err(err).Msg("Initializing OptimizelyClient")
			return
		}
	})

	return optlyClient
}
