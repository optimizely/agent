/****************************************************************************
 * Copyright 2019-2020,2022-2023, Optimizely, Inc. and contributors         *
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
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	optimizelyclient "github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/odp/cache"
)

// ErrEntityNotFound is returned when no entity exists with a given key
var ErrEntityNotFound = errors.New("not found")

// ErrForcedVariationsUninitialized is returned from SetForcedVariation and GetForcedVariation when the forced variations store is not initialized
var ErrForcedVariationsUninitialized = errors.New("client forced variations store not initialized")

// OptlyClient wraps an instance of the OptimizelyClient to provide higher level functionality
type OptlyClient struct {
	*optimizelyclient.OptimizelyClient
	ConfigManager      SyncedConfigManager
	ForcedVariations   *decision.MapExperimentOverridesStore
	UserProfileService decision.UserProfileService
	odpCache           cache.Cache
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

// SendOdpEventResponseModel response model
type SendOdpEventResponseModel struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// UpdateConfig uses config manager to sync and set project config
func (c *OptlyClient) UpdateConfig() {
	if c.ConfigManager != nil {
		c.ConfigManager.SyncConfig()
	}
}

// TrackEvent checks for the existence of the event before calling the OptimizelyClient Track method
func (c *OptlyClient) TrackEvent(ctx context.Context, eventKey string, uc entities.UserContext, eventTags map[string]interface{}) (*Track, error) {
	_, span := otel.Tracer("trackHandler").Start(ctx, "TrackEvent")
	defer span.End()
	span.SetAttributes(attribute.String("trackEventKey", eventKey))

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

	if err := c.Track(ctx, eventKey, uc, eventTags); err != nil {
		return &Track{}, err
	}

	return tr, nil
}

// SetForcedVariation sets a forced variation for the argument experiment key and user ID
// Returns false if the same forced variation was already set for the argument experiment and user, true otherwise
// Returns an error when forced variations are not available on this OptlyClient instance
func (c *OptlyClient) SetForcedVariation(ctx context.Context, experimentKey, userID, variationKey string) (*Override, error) {
	_, span := otel.Tracer("overrideHandler").Start(ctx, "SetForcedVariation")
	defer span.End()

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
	if optimizelyConfig := c.GetOptimizelyConfig(ctx); optimizelyConfig == nil {
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

	span.SetAttributes(attribute.String("variationKey", override.VariationKey))
	span.SetAttributes(attribute.String("experimentKey", override.ExperimentKey))

	c.ForcedVariations.SetVariation(forcedVariationKey, variationKey)
	return &override, nil
}

// RemoveForcedVariation removes any forced variation that was previously set for the argument experiment key and user ID
func (c *OptlyClient) RemoveForcedVariation(ctx context.Context, experimentKey, userID string) (*Override, error) {
	_, span := otel.Tracer("overrideHandler").Start(ctx, "RemoveForcedVariation")
	defer span.End()

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

	span.SetAttributes(attribute.String("variationKey", override.VariationKey))
	span.SetAttributes(attribute.String("experimentKey", override.ExperimentKey))

	return &override, nil
}

// ActivateFeature activates a feature for a given user by getting the feature enabled status and all
// associated variables
func (c *OptlyClient) ActivateFeature(ctx context.Context, key string, uc entities.UserContext, disableTracking bool) (*Decision, error) {
	_, span := otel.Tracer("activateHandler").Start(ctx, "ActivateFeature")
	defer span.End()

	unsafeDecisionInfo, err := c.GetDetailedFeatureDecisionUnsafe(ctx, key, uc, disableTracking)
	if err != nil {
		return &Decision{}, err
	}

	dec := &Decision{
		UserID:        uc.ID,
		FeatureKey:    key,
		Variables:     unsafeDecisionInfo.VariableMap,
		Enabled:       unsafeDecisionInfo.Enabled,
		Type:          "feature",
		ExperimentKey: unsafeDecisionInfo.ExperimentKey,
		VariationKey:  unsafeDecisionInfo.VariationKey,
	}

	span.SetAttributes(attribute.String("variationKey", dec.VariationKey))
	span.SetAttributes(attribute.String("experimentKey", dec.ExperimentKey))
	span.SetAttributes(attribute.String("featureKey", dec.FeatureKey))
	span.SetAttributes(attribute.Bool("enabled", dec.Enabled))
	span.SetAttributes(attribute.String("type", dec.Type))

	return dec, nil
}

// ActivateExperiment activates an experiment
func (c *OptlyClient) ActivateExperiment(ctx context.Context, key string, uc entities.UserContext, disableTracking bool) (*Decision, error) {
	_, span := otel.Tracer("activateHandler").Start(ctx, "ActivateExperiment")
	defer span.End()

	var variation string
	var err error

	if disableTracking {
		variation, err = c.GetVariation(ctx, key, uc)
	} else {
		variation, err = c.Activate(ctx, key, uc)
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

	span.SetAttributes(attribute.String("variationKey", dec.VariationKey))
	span.SetAttributes(attribute.String("experimentKey", dec.ExperimentKey))
	span.SetAttributes(attribute.Bool("enabled", dec.Enabled))
	span.SetAttributes(attribute.String("type", dec.Type))

	return dec, nil
}
