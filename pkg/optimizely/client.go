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
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
)

// ErrEntityNotFound is returned when no entity exists with a given key
var ErrEntityNotFound = errors.New("not found")

// ErrForcedVariationsUninitialized is returned from SetForcedVariation and GetForcedVariation when the forced variations store is not initialized
var ErrForcedVariationsUninitialized = errors.New("client forced variations store not initialized")

// OptlyClient wraps an instance of the OptimizelyClient to provide higher level functionality
type OptlyClient struct {
	*optimizelyclient.OptimizelyClient
	ConfigManager    SyncedConfigManager
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
	Error         string                 `json:"error"`
}

// Override model
type Override struct {
	UserID           string   `json:"userId"`
	ExperimentKey    string   `json:"experimentKey"`
	VariationKey     string   `json:"variationKey"`
	PrevVariationKey string   `json:"prevVariationKey"`
	Messages         []string `json:"messages"`
}

// UpdateConfig uses config manager to sync and set project config
func (c *OptlyClient) UpdateConfig() {
	if c.ConfigManager != nil {
		c.ConfigManager.SyncConfig()
	}
}

// TrackEvent checks for the existence of the event before calling the OptimizelyClient Track method
func (c *OptlyClient) TrackEvent(eventKey string, uc entities.UserContext, eventTags map[string]interface{}) error {
	pc, err := c.ConfigManager.GetConfig()
	if err != nil {
		return err
	}

	if _, err = pc.GetEventByKey(eventKey); err != nil {
		return fmt.Errorf("eventKey: %q %w", eventKey, ErrEntityNotFound)
	}

	return c.Track(eventKey, uc, eventTags)
}

// SetForcedVariation sets a forced variation for the argument experiment key and user ID
// Returns false if the same forced variation was already set for the argument experiment and user, true otherwise
// Returns an error when forced variations are not available on this OptlyClient instance
func (c *OptlyClient) SetForcedVariation(experimentKey, userID, variationKey string) (*Override, error) {
	if c.ForcedVariations == nil {
		return &Override{}, ErrForcedVariationsUninitialized
	}

	override := Override{
		UserID:        userID,
		ExperimentKey: experimentKey,
		VariationKey:  variationKey,
	}

	messages := make([]string, 0, 2)
	// Check the entities exist as part of the Optimizely configuration
	if optimizelyConfig := c.GetOptimizelyConfig(); optimizelyConfig == nil {
		messages = append(messages, "override cannot be validated via configuration")
	} else if experiment, ok := optimizelyConfig.ExperimentsMap[experimentKey]; !ok {
		messages = append(messages, "experimentKey not found in configuration")
	} else if _, ok := experiment.VariationsMap[variationKey]; !ok {
		messages = append(messages, "variationKey not found in configuration")
	}

	forcedVariationKey := decision.ExperimentOverrideKey{
		UserID:        userID,
		ExperimentKey: experimentKey,
	}

	if prevVariationKey, ok := c.ForcedVariations.GetVariation(forcedVariationKey); ok {
		override.PrevVariationKey = prevVariationKey
		messages = append(messages, "updating previous override")
	}

	if len(messages) > 0 {
		override.Messages = messages
	}

	c.ForcedVariations.SetVariation(forcedVariationKey, variationKey)
	return &override, nil
}

// RemoveForcedVariation removes any forced variation that was previously set for the argument experiment key and user ID
func (c *OptlyClient) RemoveForcedVariation(experimentKey, userID string) (*Override, error) {
	if c.ForcedVariations == nil {
		return &Override{}, ErrForcedVariationsUninitialized
	}

	override := Override{
		UserID:        userID,
		ExperimentKey: experimentKey,
		VariationKey:  "",
	}

	forcedVariationKey := decision.ExperimentOverrideKey{
		UserID:        userID,
		ExperimentKey: experimentKey,
	}

	messages := make([]string, 0, 1)
	if prevVariationKey, ok := c.ForcedVariations.GetVariation(forcedVariationKey); ok {
		override.PrevVariationKey = prevVariationKey
		messages = append(messages, "removing previous override")
	} else {
		messages = append(messages, "no pre-existing override")
	}

	override.Messages = messages
	c.ForcedVariations.RemoveVariation(forcedVariationKey)

	return &override, nil
}

// ActivateFeature activates a feature for a given user by getting the feature enabled status and all
// associated variables
func (c *OptlyClient) ActivateFeature(key string, uc entities.UserContext, disableTracking bool) (*Decision, error) {
	enabled, variables, err := c.GetAllFeatureVariables(key, uc)
	if err != nil {
		return &Decision{}, err
	}

	// HACK - Triggers impression events when applicable. This is not
	// ideal since we're making TWO decisions for each feature now. TODO OASIS-5549
	if !disableTracking {
		_, tErr := c.IsFeatureEnabled(key, uc)
		if tErr != nil {
			return &Decision{}, tErr
		}
	}

	// TODO add experiment and variation keys where applicable
	dec := &Decision{
		FeatureKey: key,
		Variables:  variables,
		Enabled:    enabled,
		Type:       "feature",
	}

	return dec, nil
}

// ActivateExperiment activates an experiment
func (c *OptlyClient) ActivateExperiment(key string, uc entities.UserContext, disableTracking bool) (*Decision, error) {
	var variation string
	var err error

	if disableTracking {
		variation, err = c.GetVariation(key, uc)
	} else {
		variation, err = c.Activate(key, uc)
	}
	if err != nil {
		return &Decision{}, err
	}

	dec := &Decision{
		ExperimentKey: key,
		VariationKey:  variation,
		Enabled:       variation != "",
		Type:          "experiment",
		Variables:     map[string]interface{}{},
	}

	return dec, nil
}
