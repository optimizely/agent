/****************************************************************************
 * Copyright 2019-202, Optimizely, Inc. and contributors                    *
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
	"strconv"

	optimizelyclient "github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
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
	UserID        string                 `json:"userId"`
	ExperimentKey string                 `json:"experimentKey"`
	FeatureKey    string                 `json:"featureKey"`
	VariationKey  string                 `json:"variationKey"`
	Type          string                 `json:"type"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	Enabled       bool                   `json:"enabled"`
	Error         string                 `json:"error,omitempty"`
}

// Override model
type Override struct {
	UserID           string   `json:"userId"`
	ExperimentKey    string   `json:"experimentKey"`
	VariationKey     string   `json:"variationKey"`
	PrevVariationKey string   `json:"prevVariationKey"`
	Messages         []string `json:"messages"`
}

// Track response model
type Track struct {
	UserID   string `json:"userId"`
	EventKey string `json:"eventKey"`
	Error    string `json:"error,omitempty"`
}

// UpdateConfig uses config manager to sync and set project config
func (c *OptlyClient) UpdateConfig() {
	if c.ConfigManager != nil {
		c.ConfigManager.SyncConfig()
	}
}

// TrackEvent checks for the existence of the event before calling the OptimizelyClient Track method
func (c *OptlyClient) TrackEvent(eventKey string, uc entities.UserContext, eventTags map[string]interface{}) (*Track, error) {
	tr := &Track{
		UserID:   uc.ID,
		EventKey: eventKey,
	}

	if pc, err := c.ConfigManager.GetConfig(); err != nil {
		return &Track{}, err
	} else if _, err := pc.GetEventByKey(eventKey); err != nil {
		tr.Error = err.Error()
		return tr, nil
	}

	if err := c.Track(eventKey, uc, eventTags); err != nil {
		return &Track{}, err
	}

	return tr, nil
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
	var enabled bool
	var featureDecision decision.FeatureDecision
	variables := make(map[string]interface{})
	var experimentKey, variationKey string

	projectConfig, err := c.OptimizelyClient.ConfigManager.GetConfig()
	if err != nil {
		return &Decision{}, err
	}

	if feature, err := projectConfig.GetFeatureByKey(key); err == nil {
		variable := entities.Variable{}
		decisionContext := decision.FeatureDecisionContext{
			Feature:       &feature,
			ProjectConfig: projectConfig,
			Variable:      variable,
		}
		if featureDecision, err = c.DecisionService.GetFeatureDecision(decisionContext, uc); err == nil && featureDecision.Variation != nil {
			enabled = featureDecision.Variation.FeatureEnabled
			variationKey = featureDecision.Variation.Key
			experimentKey = featureDecision.Experiment.Key

			if featureDecision.Source == decision.FeatureTest && !disableTracking {
				// send impression event for feature tests
				impressionEvent := event.CreateImpressionUserEvent(decisionContext.ProjectConfig, featureDecision.Experiment, *featureDecision.Variation, uc)
				c.EventProcessor.ProcessEvent(impressionEvent)
			}
		}

		for _, v := range feature.VariableMap {
			val := v.DefaultValue

			if enabled {
				if variable, ok := featureDecision.Variation.Variables[v.ID]; ok {
					val = variable.Value
				}
			}

			var out interface{}
			out = val
			switch varType := v.Type; varType {
			case entities.Boolean:
				out, err = strconv.ParseBool(val)
			case entities.Double:
				out, err = strconv.ParseFloat(val, 64)
			case entities.Integer:
				out, err = strconv.Atoi(val)
			case entities.String:
			default:
			}

			variables[v.Key] = out
		}
	}

	dec := &Decision{
		UserID:        uc.ID,
		FeatureKey:    key,
		Variables:     variables,
		Enabled:       enabled,
		Type:          "feature",
		ExperimentKey: experimentKey,
		VariationKey:  variationKey,
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
		UserID:        uc.ID,
		ExperimentKey: key,
		VariationKey:  variation,
		Enabled:       variation != "",
		Type:          "experiment",
		Variables:     map[string]interface{}{},
	}

	return dec, nil
}
