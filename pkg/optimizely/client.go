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
	"github.com/rs/zerolog/log"

	optimizelyclient "github.com/optimizely/go-sdk/pkg/client"
	optimizelyconfig "github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/entities"
)

// OptlyClient wraps an instance of the OptimizelyClient to provide higher level functionality
type OptlyClient struct {
	*optimizelyclient.OptimizelyClient
	ConfigManager *optimizelyconfig.PollingProjectConfigManager
}

// ListFeatures returns all available features
func (c *OptlyClient) ListFeatures() (features []entities.Feature, err error) {
	projectConfig, err := c.GetProjectConfig()
	if err != nil {
		log.Error().Err(err).Msg("Attempting to ListFeatures")
		return features, err
	}

	features = projectConfig.GetFeatureList()
	return features, err
}

// GetFeature returns the feature definition
func (c *OptlyClient) GetFeature(featureKey string) (feature entities.Feature, err error) {
	projectConfig, err := c.GetProjectConfig()
	if err != nil {
		log.Error().Err(err).Str("featureKey", featureKey).Msg("Attempting to GetFeature")
		return feature, err
	}

	return projectConfig.GetFeatureByKey(featureKey)
}

// UpdateConfig uses config manager to sync and set project config
func (c *OptlyClient) UpdateConfig() {
	if c.ConfigManager != nil {
		c.ConfigManager.SyncConfig([]byte{})
	}
}

// GetAndTrackFeatureWithContext calls the OptimizelyClient with the current OptlyContext this does NOT track experiment conversions
func (c *OptlyClient) GetAndTrackFeatureWithContext(featureKey string, ctx *OptlyContext) (enabled bool, variableMap map[string]string, err error) {
	// TODO add tracking
	return c.GetFeatureWithContext(featureKey, ctx)
}

// GetFeatureWithContext calls the OptimizelyClient with the current OptlyContext
func (c *OptlyClient) GetFeatureWithContext(featureKey string, ctx *OptlyContext) (enabled bool, variableMap map[string]string, err error) {
	return c.GetAllFeatureVariables(featureKey, *ctx.UserContext)
}
