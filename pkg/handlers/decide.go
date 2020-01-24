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
	"net/http"

	"github.com/optimizely/agent/pkg/middleware"
)

// Decision Model
type Decision struct {
	Key       string            `json:"key"`
	Variables map[string]string `json:"variables,omitempty"`
	ID        int32             `json:"id,omitempty"`
	Enabled   bool              `json:"enabled"`
}

// GetFeature - Return the feature. Note: no impressions recorded for associated feature tests.
func GetFeature(w http.ResponseWriter, r *http.Request) {
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
func TrackFeature(w http.ResponseWriter, r *http.Request) {
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
func GetVariation(w http.ResponseWriter, r *http.Request) {
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

func Decide(w http.ResponseWriter, r *http.Request) {
	optlyClient, optlyContext, err := parseContext(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	if feature, err := middleware.GetFeature(r); err == nil {
		renderFeature(w, r, feature.Key, optlyClient, optlyContext)
		return
	}

	if experiment, err := middleware.GetExperiment(r); err == nil {
		renderVariation(w, r, experiment.Key, true, optlyClient, optlyContext) // true to send impression
		return
	}
}

// ActivateExperiment - Return the variatoin that a user is bucketed into and track an impression event
func ActivateExperiment(w http.ResponseWriter, r *http.Request) {
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
func ListFeatures(w http.ResponseWriter, r *http.Request) {
	optlyClient, optlyContext, err := parseContext(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	renderFeatures(w, r, optlyClient, optlyContext)
}

// TrackFeatures - List all feature decisions for a user. Impression events are recorded for all applicable feature tests.
func TrackFeatures(w http.ResponseWriter, r *http.Request) {
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
