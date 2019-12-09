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
	"errors"

	optimizelyclient "github.com/optimizely/go-sdk/pkg/client"
	optimizelyconfig "github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
)

// OptlyClient wraps an instance of the OptimizelyClient to provide higher level functionality
type OptlyClient struct {
	*optimizelyclient.OptimizelyClient
	ConfigManager    *optimizelyconfig.PollingProjectConfigManager
	ForcedVariations *decision.MapExperimentOverridesStore
}

// FeatureDecision is the return type of methods that provide feature enabled and variable value decisions for a given OptlyContext
type FeatureDecision struct {
	Key            string
	Enabled        bool
	VariableValues map[string]string
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

// DecideFeatures - Returns a slice of FeatureDecision pointers representing the decisions for all features for the argument context cotext
func (c *OptlyClient) DecideFeatures(ctx *OptlyContext) (map[string]*FeatureDecision, error) {
	featureDecisions := make(map[string]*FeatureDecision)

	featureEntities, err := c.ListFeatures()
	if err != nil {
		// TODO: wrap error?
		return featureDecisions, err
	}

	for _, feature := range featureEntities {
		enabled, variables, err := c.GetFeatureWithContext(feature.Key, ctx)
		if err != nil {
			// TODO: wrap error?
			return map[string]*FeatureDecision{}, err
		}

		featureDecisions[feature.Key] = &FeatureDecision{
			Enabled:        enabled,
			Key:            feature.Key,
			VariableValues: variables,
		}
	}

	return featureDecisions, nil
}

// GetFeature returns the feature definition
func (c *OptlyClient) GetFeature(featureKey string) (feature entities.Feature, err error) {
	projectConfig, err := c.GetProjectConfig()
	if err != nil {
		return feature, err
	}

	return projectConfig.GetFeatureByKey(featureKey)
}

// GetExperiment returns the experiment definition
func (c *OptlyClient) GetExperiment(experimentKey string) (experiment entities.Experiment, err error) {
	projectConfig, err := c.GetProjectConfig()
	if err != nil {
		return experiment, err
	}

	return projectConfig.GetExperimentByKey(experimentKey)
}

// UpdateConfig uses config manager to sync and set project config
func (c *OptlyClient) UpdateConfig() {
	if c.ConfigManager != nil {
		c.ConfigManager.SyncConfig([]byte{})
	}
}

// TrackEventWithContext calls the OptimizelyClient Track method with the current OptlyContext.
func (c *OptlyClient) TrackEventWithContext(eventKey string, ctx *OptlyContext, eventTags map[string]interface{}) error {
	return c.Track(eventKey, *ctx.UserContext, eventTags)
}

// GetFeatureWithContext calls the OptimizelyClient with the current OptlyContext
// TODO: Refactor to return FeatureDecision
func (c *OptlyClient) GetFeatureWithContext(featureKey string, ctx *OptlyContext) (enabled bool, variableMap map[string]string, err error) {
	return c.GetAllFeatureVariables(featureKey, *ctx.UserContext)
}

// GetExperimentVariation calls the OptimizelyClient with the current OptlyContext
func (c *OptlyClient) GetExperimentVariation(experimentKey string, shouldActivate bool, ctx *OptlyContext) (variation entities.Variation, err error) {
	var experiment entities.Experiment
	experiment, err = c.GetExperiment(experimentKey)
	if err != nil {
		return variation, nil
	}

	var variationKey string
	if shouldActivate {
		variationKey, err = c.Activate(experimentKey, *ctx.UserContext)
	} else {
		variationKey, err = c.GetVariation(experimentKey, *ctx.UserContext)
	}

	if err != nil {
		return variation, err
	}

	// @TODO: can expose a way to look up variation by key in the SDK
	for _, experimentVariation := range experiment.Variations {
		if experimentVariation.Key == variationKey {
			variation = experimentVariation
		}
	}

	return variation, nil
}

// ErrForcedVariationsUninitialized is returned from SetForcedVariation and GetForcedVariation when the forced variations store is not initialized
var ErrForcedVariationsUninitialized = errors.New("client forced variations store not initialized")

// SetForcedVariation sets a forced variation for the argument experiment key and user ID
// Returns false if the same forced variation was already set for the argument experiment and user, true otherwise
// Returns an error when forced variations are not available on this OptlyClient instance
func (c *OptlyClient) SetForcedVariation(experimentKey, userID, variationKey string) (bool, error) {
	if c.ForcedVariations == nil {
		return false, ErrForcedVariationsUninitialized
	}
	forcedVariationKey := decision.ExperimentOverrideKey{
		UserID:        userID,
		ExperimentKey: experimentKey,
	}
	previousVariationKey, ok := c.ForcedVariations.GetVariation(forcedVariationKey)
	c.ForcedVariations.SetVariation(forcedVariationKey, variationKey)
	wasSet := !ok || previousVariationKey != variationKey
	return wasSet, nil
}

// RemoveForcedVariation removes any forced variation that was previously set for the argument experiment key and user ID
func (c *OptlyClient) RemoveForcedVariation(experimentKey, userID string) error {
	if c.ForcedVariations == nil {
		return ErrForcedVariationsUninitialized
	}
	forcedVariationKey := decision.ExperimentOverrideKey{
		UserID:        userID,
		ExperimentKey: experimentKey,
	}
	c.ForcedVariations.RemoveVariation(forcedVariationKey)
	return nil
}
