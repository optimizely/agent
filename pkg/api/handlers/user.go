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

// Package handlers //
package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/optimizely/sidedoor/pkg/api/middleware"
	"github.com/optimizely/sidedoor/pkg/api/models"
)

type eventTags map[string]interface{}

// UserHandler implements the UserAPI interface
type UserHandler struct{}

// TrackEvent - track a given event for the current user
func (h *UserHandler) TrackEvent(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	if err != nil {
		RenderError(err, http.StatusUnprocessableEntity, w, r)
		return
	}

	optlyContext, err := middleware.GetOptlyContext(r)
	if err != nil {
		RenderError(err, http.StatusUnprocessableEntity, w, r)
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
		middleware.GetLogger(r).Err(err).Str("eventKey", eventKey).Msg("error tracking event")
		RenderError(err, http.StatusNotFound, w, r)
		return
	}

	render.NoContent(w, r)
}

// GetFeature - Return the feature and record impression if applicable.
// Tracking impressions is only supported for "Feature Tests" as part of the SDK contract,
// and only applicable during a POST request.
func (h *UserHandler) GetFeature(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	if err != nil {
		RenderError(err, http.StatusUnprocessableEntity, w, r)
		return
	}

	optlyContext, err := middleware.GetOptlyContext(r)
	if err != nil {
		RenderError(err, http.StatusUnprocessableEntity, w, r)
		return
	}

	featureKey := chi.URLParam(r, "featureKey")

	// Track events during a POST.
	// Note: Tracking is only done in the context of a "Feature Test"
	if r.Method == "POST" {
		// HACK - Triggers an impression event when applicable. This is not
		// ideal since we're making TWO deicisions now. OASIS-5549
		_, softErr := optlyClient.IsFeatureEnabled(featureKey, *optlyContext.UserContext)
		if softErr != nil {
			// Swallowing the error to allow the response to be made and not break downstream consumers.
			middleware.GetLogger(r).Error().Err(softErr).Str("featureKey", featureKey).Msg("Calling IsFeatureEnabled")
		}
	}

	enabled, variables, err := optlyClient.GetAndTrackFeatureWithContext(featureKey, optlyContext)

	if err != nil {
		middleware.GetLogger(r).Error().Str("featureKey", featureKey).Msg("Calling GetFeature")
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	feature := &models.Feature{
		Enabled:   enabled,
		Key:       featureKey,
		Variables: variables,
	}

	render.JSON(w, r, feature)
}
