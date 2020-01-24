/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                        *
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

// Package handlers //
package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
)

type eventTags map[string]interface{}

// UserHandler implements the UserAPI interface
type UserHandler struct{}

// TrackEvent - track a given event for the current user
func (h *UserHandler) TrackEvent(w http.ResponseWriter, r *http.Request) {
	optlyClient, optlyContext, err := parseContext(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	var tags eventTags
	err = ParseRequestBody(r, &tags)
	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	eventKey := chi.URLParam(r, "eventKey")
	if eventKey == "" {
		err = fmt.Errorf("missing required path parameter: eventKey")
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	err = optlyClient.TrackEventWithContext(eventKey, optlyContext, tags)
	if err != nil {
		middleware.GetLogger(r).Error().Err(err).Str("eventKey", eventKey).Msg("error tracking event")
		RenderError(err, http.StatusNotFound, w, r)
		return
	}
	middleware.GetLogger(r).Debug().Str("eventKey", eventKey).Msg("tracking event")
	render.NoContent(w, r)
}

// GetFeature - Return the feature. Note: no impressions recorded for associated feature tests.
func (h *UserHandler) GetFeature(w http.ResponseWriter, r *http.Request) {
	optlyClient, optlyContext, err := parseContext(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}
	feature, err := middleware.GetFeature(r)
	if err != nil {
		middleware.GetLogger(r).Error().Err(err).Msg("Calling GetFeature")
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}
	renderFeature(w, r, feature.Key, optlyClient, optlyContext)
}

// TrackFeature - Return the feature and record impression if applicable.
// Tracking impressions is only supported for "Feature Tests" as part of the SDK contract.
func (h *UserHandler) TrackFeature(w http.ResponseWriter, r *http.Request) {
	optlyClient, optlyContext, err := parseContext(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	feature, err := middleware.GetFeature(r)
	if err != nil {
		middleware.GetLogger(r).Error().Err(err).Msg("Calling GetFeature")
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	// HACK - Triggers an impression event when applicable. This is not
	// ideal since we're making TWO decisions now. OASIS-5549
	enabled, softErr := optlyClient.IsFeatureEnabled(feature.Key, *optlyContext.UserContext)
	middleware.GetLogger(r).Info().Str("featureKey", feature.Key).Bool("enabled", enabled).Msg("Calling IsFeatureEnabled")

	if softErr != nil {
		// Swallowing the error to allow the response to be made and not break downstream consumers.
		middleware.GetLogger(r).Error().Err(softErr).Str("featureKey", feature.Key).Msg("Calling IsFeatureEnabled")
	}

	renderFeature(w, r, feature.Key, optlyClient, optlyContext)
}

// GetVariation - Return the variation that a user is bucketed into
func (h *UserHandler) GetVariation(w http.ResponseWriter, r *http.Request) {
	optlyClient, optlyContext, err := parseContext(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	experiment, err := middleware.GetExperiment(r)
	if err != nil {
		middleware.GetLogger(r).Error().Err(err).Msg("Calling middleware GetExperiment")
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	renderVariation(w, r, experiment.Key, false, optlyClient, optlyContext)
}

// ActivateExperiment - Return the variatoin that a user is bucketed into and track an impression event
func (h *UserHandler) ActivateExperiment(w http.ResponseWriter, r *http.Request) {
	optlyClient, optlyContext, err := parseContext(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	experiment, err := middleware.GetExperiment(r)
	if err != nil {
		middleware.GetLogger(r).Error().Err(err).Msg("Calling middleware GetExperiment")
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	renderVariation(w, r, experiment.Key, true, optlyClient, optlyContext) // true to send impression
}

// ListFeatures - List all feature decisions for a user
// Note: no impressions recorded for associated feature tests.
func (h *UserHandler) ListFeatures(w http.ResponseWriter, r *http.Request) {
	optlyClient, optlyContext, err := parseContext(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	renderFeatures(w, r, optlyClient, optlyContext)
}

// TrackFeatures - List all feature decisions for a user. Impression events are recorded for all applicable feature tests.
func (h *UserHandler) TrackFeatures(w http.ResponseWriter, r *http.Request) {
	optlyClient, optlyContext, err := parseContext(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	// HACK - Triggers impression events when applicable. This is not
	// ideal since we're making TWO decisions for each feature now. OASIS-5549
	enabledFeatures, softErr := optlyClient.GetEnabledFeatures(*optlyContext.UserContext)
	middleware.GetLogger(r).Info().Strs("enabledFeatures", enabledFeatures).Msg("Calling GetEnabledFeatures")
	if softErr != nil {
		// Swallowing the error to allow the response to be made and not break downstream consumers.
		middleware.GetLogger(r).Error().Err(softErr).Msg("Calling GetEnabledFeatures")
	}

	renderFeatures(w, r, optlyClient, optlyContext)
}

// parseContext extract the common references from the request context
func parseContext(r *http.Request) (*optimizely.OptlyClient, *optimizely.OptlyContext, error) {
	optlyClient, err := middleware.GetOptlyClient(r)
	if err != nil {
		return nil, nil, err
	}

	optlyContext, err := middleware.GetOptlyContext(r)
	if err != nil {
		return nil, nil, err
	}

	return optlyClient, optlyContext, nil
}

// getModelOfFeatureDecision - Returns a Feature representing the feature decision from the provided client and context
func getModelOfFeatureDecision(featureKey string, optlyClient *optimizely.OptlyClient, optlyContext *optimizely.OptlyContext) (*Feature, error) {
	enabled, variables, err := optlyClient.GetFeatureWithContext(featureKey, optlyContext)
	if err != nil {
		return nil, err
	}
	return &Feature{
		Key:       featureKey,
		Enabled:   enabled,
		Variables: variables,
	}, nil
}

// renderFeature excapsulates extracting a Feature from the Optimizely SDK and rendering a feature response.
func renderFeature(w http.ResponseWriter, r *http.Request, featureKey string, optlyClient *optimizely.OptlyClient, optlyContext *optimizely.OptlyContext) {
	featureModel, err := getModelOfFeatureDecision(featureKey, optlyClient, optlyContext)
	if err != nil {
		middleware.GetLogger(r).Error().Err(err).Str("featureKey", featureKey).Msg("Calling GetFeatureWithContext")
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}
	middleware.GetLogger(r).Debug().Str("featureKey", featureKey).Msg("rendering feature")
	render.JSON(w, r, featureModel)
}

// renderFeatures encapsulates extracting decisions for all available features from the Optimizely SDK and rendering a response with all those decisions
func renderFeatures(w http.ResponseWriter, r *http.Request, optlyClient *optimizely.OptlyClient, optlyContext *optimizely.OptlyContext) {
	features, err := optlyClient.ListFeatures()
	if err != nil {
		middleware.GetLogger(r).Error().Msg("Calling ListFeatures")
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	featuresCount := len(features)
	featureModels := make([]*Feature, 0, featuresCount)
	for _, feature := range features {
		featureModel, err := getModelOfFeatureDecision(feature.Key, optlyClient, optlyContext)
		if err != nil {
			middleware.GetLogger(r).Error().Err(err).Str("featureKey", feature.Key).Msg("Calling GetFeatureWithContext")
			RenderError(err, http.StatusInternalServerError, w, r)
			return
		}
		featureModels = append(featureModels, featureModel)
		middleware.GetLogger(r).Debug().Str("featureKey", feature.Key).Msg("rendering feature")
	}

	render.JSON(w, r, featureModels)
}

// renderVariation encapsulates extracting Variation from the Optimizely SDK and rendering a response
func renderVariation(w http.ResponseWriter, r *http.Request, experimentKey string, shouldActivate bool, optlyClient *optimizely.OptlyClient, optlyContext *optimizely.OptlyContext) {
	variation, err := optlyClient.GetExperimentVariation(experimentKey, shouldActivate, optlyContext)
	if err != nil {
		middleware.GetLogger(r).Error().Err(err).Str("experimentKey", experimentKey).Msg("Calling GetVariation")
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	variationModel := &Variation{
		Key: variation.Key,
		ID:  variation.ID,
	}
	middleware.GetLogger(r).Debug().Str("experimentKey", experimentKey).Msg("rendering variation")
	render.JSON(w, r, variationModel)
}
