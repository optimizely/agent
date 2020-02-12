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
	"fmt"

	optimizelyclient "github.com/optimizely/go-sdk/pkg/client"
	optimizelyconfig "github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
)

var errNullOptimizelyConfig = errors.New("optimizely config is null")

// OptlyClient wraps an instance of the OptimizelyClient to provide higher level functionality
type OptlyClient struct {
	*optimizelyclient.OptimizelyClient
	ConfigManager    *optimizelyconfig.PollingProjectConfigManager
	ForcedVariations *decision.MapExperimentOverridesStore
}

// Decision Model
type Decision struct {
	ExperimentKey string                 `json:"experimentKey"`
	FeatureKey    string                 `json:"featureKey"`
	VariationKey  string                 `json:"variationKey"`
	Type          string                 `json:"type"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	Enabled       bool                   `json:"enabled"`
}

// ListFeatures returns all available features
func (c *OptlyClient) ListFeatures() (features []optimizelyconfig.OptimizelyFeature, err error) {
	optimizelyConfig := c.GetOptimizelyConfig()
	if optimizelyConfig == nil {
		return features, errNullOptimizelyConfig
	}
	features = []optimizelyconfig.OptimizelyFeature{}
	for _, feature := range optimizelyConfig.FeaturesMap {
		features = append(features, feature)
	}

	return features, err
}

// ErrEntityNotFound is returned when no entity exists with a given key
var ErrEntityNotFound = errors.New("not found")

// GetFeature returns the feature definition
func (c *OptlyClient) GetFeature(featureKey string) (optimizelyconfig.OptimizelyFeature, error) {

	optimizelyConfig := c.GetOptimizelyConfig()
	if optimizelyConfig == nil {
		return optimizelyconfig.OptimizelyFeature{}, errNullOptimizelyConfig
	}

	if feature, ok := optimizelyConfig.FeaturesMap[featureKey]; ok {
		return feature, nil
	}

	return optimizelyconfig.OptimizelyFeature{}, fmt.Errorf("feature %s %w", featureKey, ErrEntityNotFound)
}

// ListExperiments returns all available experiments
func (c *OptlyClient) ListExperiments() (experiments []optimizelyconfig.OptimizelyExperiment, err error) {
	optimizelyConfig := c.GetOptimizelyConfig()
	if optimizelyConfig == nil {
		return experiments, errNullOptimizelyConfig
	}
	experiments = []optimizelyconfig.OptimizelyExperiment{}
	for _, experiment := range optimizelyConfig.ExperimentsMap {
		experiments = append(experiments, experiment)
	}

	return experiments, err
}

// GetExperiment returns the experiment definition
func (c *OptlyClient) GetExperiment(experimentKey string) (optimizelyconfig.OptimizelyExperiment, error) {
	optimizelyConfig := c.GetOptimizelyConfig()
	if optimizelyConfig == nil {
		return optimizelyconfig.OptimizelyExperiment{}, errNullOptimizelyConfig
	}

	if experiment, ok := optimizelyConfig.ExperimentsMap[experimentKey]; ok {
		return experiment, nil
	}

	return optimizelyconfig.OptimizelyExperiment{}, fmt.Errorf("experiment %s %w", experimentKey, ErrEntityNotFound)
}

// UpdateConfig uses config manager to sync and set project config
func (c *OptlyClient) UpdateConfig() {
	if c.ConfigManager != nil {
		c.ConfigManager.SyncConfig()
	}
}

// TrackEventWithContext calls the OptimizelyClient Track method with the current OptlyContext.
func (c *OptlyClient) TrackEventWithContext(eventKey string, ctx *OptlyContext, eventTags map[string]interface{}) error {
	return c.Track(eventKey, *ctx.UserContext, eventTags)
}

// GetFeatureWithContext calls the OptimizelyClient with the current OptlyContext
func (c *OptlyClient) GetFeatureWithContext(featureKey string, ctx *OptlyContext) (enabled bool, variableMap map[string]interface{}, err error) {
	return c.GetAllFeatureVariables(featureKey, *ctx.UserContext)
}

// GetExperimentVariation calls the OptimizelyClient with the current OptlyContext
func (c *OptlyClient) GetExperimentVariation(experimentKey string, shouldActivate bool, ctx *OptlyContext) (variation optimizelyconfig.OptimizelyVariation, err error) {

	optimizelyConfig := c.GetOptimizelyConfig()
	if optimizelyConfig == nil {
		return variation, errors.New("optimizely config is null")
	}

	var experiment optimizelyconfig.OptimizelyExperiment
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

	if experimentVariation, ok := experiment.VariationsMap[variationKey]; ok {
		variation = experimentVariation
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

// ActivateFeature activates a feature for a given user by getting the feature enabled status and all
// associated variables
func (c *OptlyClient) ActivateFeature(feature *optimizelyconfig.OptimizelyFeature, uc entities.UserContext, disableTracking bool) (*Decision, error) {
	enabled, variables, err := c.GetAllFeatureVariables(feature.Key, uc)
	if err != nil {
		return &Decision{}, err
	}

	// HACK - Triggers impression events when applicable. This is not
	// ideal since we're making TWO decisions for each feature now. TODO OASIS-5549
	if !disableTracking {
		_, tErr := c.IsFeatureEnabled(feature.Key, uc)
		if tErr != nil {
			return &Decision{}, tErr
		}
	}

	// TODO add experiment and variation keys where applicable
	dec := &Decision{
		FeatureKey: feature.Key,
		Variables:  variables,
		Enabled:    enabled,
		Type:       "feature",
	}

	return dec, nil
}

// ActivateExperiment activates an experiment
func (c *OptlyClient) ActivateExperiment(experiment *optimizelyconfig.OptimizelyExperiment, uc entities.UserContext, disableTracking bool) (*Decision, error) {
	var variation string
	var err error

	if disableTracking {
		variation, err = c.GetVariation(experiment.Key, uc)
	} else {
		variation, err = c.Activate(experiment.Key, uc)
	}
	if err != nil {
		return &Decision{}, err
	}

	dec := &Decision{
		ExperimentKey: experiment.Key,
		VariationKey:  variation,
		Enabled:       variation != "",
		Type:          "experiment",
		Variables:     map[string]interface{}{},
	}

	return dec, nil
}
