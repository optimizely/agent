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

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/optimizely/sidedoor/pkg/api/models"
	"github.com/optimizely/sidedoor/pkg/optimizely"
)

// ListFeatures - List all features
func ListFeatures(w http.ResponseWriter, r *http.Request) {
	features, err := optimizely.Client().ListFeatures()
	if err != nil {
		render.JSON(w, r, render.M{
			"error": err.Error(),
		})
		return
	}

	render.JSON(w, r, features)
}

// GetFeature - Get requested feature
func GetFeature(w http.ResponseWriter, r *http.Request) {
	featureKey := chi.URLParam(r, "featureKey")

	feature, err := optimizely.Client().GetFeature(featureKey)
	if err != nil {
		render.JSON(w, r, render.M{
			"error": err.Error(),
		})
		return
	}

	render.JSON(w, r, feature)
}

// ActivateFeature - Return the feature and record impression
func ActivateFeature(w http.ResponseWriter, r *http.Request) {
	featureKey := chi.URLParam(r, "featureKey")
	userID := r.URL.Query().Get("userId")

	if userID == "" {
		log.Error().Msg("Invalid request, missing userId")
		render.JSON(w, r, render.M{
			"error": "missing userId",
		})
		return
	}

	context := optimizely.NewContext(userID, map[string]interface{}{})
	enabled, variables, err := context.GetFeature(featureKey)

	if err != nil {
		log.Error().Str("featureKey", featureKey).Str("userID", userID).Msg("Calling ActivateFeature")
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
