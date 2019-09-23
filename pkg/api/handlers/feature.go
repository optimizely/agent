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
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"

	"github.com/optimizely/sidedoor/pkg/api/middleware"

	"github.com/optimizely/sidedoor/pkg/api/models"
	"github.com/optimizely/sidedoor/pkg/optimizely"
)

// FeatureHandler implements the FeatureApi interface
type FeatureHandler struct{}

// ListFeatures - List all features
func (h *FeatureHandler) ListFeatures(w http.ResponseWriter, r *http.Request) {
	optlyClient, ok := r.Context().Value(middleware.OptlyClientKey).(*optimizely.OptlyClient)
	if !ok {
		http.Error(w, "OptlyClient not available", http.StatusUnprocessableEntity)
		return
	}

	features, err := optlyClient.ListFeatures()
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, render.M{
			"error": err.Error(),
		})
		return
	}

	render.JSON(w, r, features)
}

// GetFeature - Get requested feature
func (h *FeatureHandler) GetFeature(w http.ResponseWriter, r *http.Request) {
	optlyClient, ok := r.Context().Value(middleware.OptlyClientKey).(*optimizely.OptlyClient)
	if !ok {
		http.Error(w, "OptlyClient not available", http.StatusUnprocessableEntity)
		return
	}

	featureKey := chi.URLParam(r, "featureKey")
	feature, err := optlyClient.GetFeature(featureKey)
	if err != nil {
		// TODO need to disinguish between an error and DNE
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, render.M{
			"error": err.Error(),
		})
		return
	}

	render.JSON(w, r, feature)
}

// ActivateFeature - Return the feature and record impression
func (h *FeatureHandler) ActivateFeature(w http.ResponseWriter, r *http.Request) {
	optlyClient, ok := r.Context().Value(middleware.OptlyClientKey).(*optimizely.OptlyClient)
	if !ok {
		http.Error(w, "OptlyClient not available", http.StatusUnprocessableEntity)
		return
	}

	featureKey := chi.URLParam(r, "featureKey")
	userID := r.URL.Query().Get("userId")

	if userID == "" {
		log.Error().Msg("Invalid request, missing userId")
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, render.M{
			"error": "missing userId",
		})
		return
	}

	// TODO replace with middleware for testability
	context := optimizely.NewContext(userID, map[string]interface{}{})
	enabled, variables, err := optlyClient.GetAndTrackFeatureWithContext(featureKey, context)

	if err != nil {
		log.Error().Str("featureKey", featureKey).Str("userID", userID).Msg("Calling ActivateFeature")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, render.M{
			"error": err,
		})
		return
	}

	feature := &models.Feature{
		Enabled:   enabled,
		Key:       featureKey,
		Variables: variables,
	}

	render.JSON(w, r, feature)
}
