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

	"github.com/optimizely/sidedoor/pkg/api/middleware"
)

// FeatureHandler implements the FeatureAPI interface
type FeatureHandler struct{}

// ListFeatures - List all features
func (h *FeatureHandler) ListFeatures(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	if err != nil {
		RenderError(err, http.StatusUnprocessableEntity, w, r)
		return
	}

	features, err := optlyClient.ListFeatures()
	if err != nil {
		middleware.GetLogger(r).Error().Msg("Calling ListFeature")
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	render.JSON(w, r, features)
}

// GetFeature - Get requested feature
func (h *FeatureHandler) GetFeature(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	if err != nil {
		RenderError(err, http.StatusUnprocessableEntity, w, r)
		return
	}

	featureKey := chi.URLParam(r, "featureKey")
	feature, err := optlyClient.GetFeature(featureKey)
	if err != nil {
		middleware.GetLogger(r).Error().Str("featureKey", featureKey).Msg("Calling GetFeature")
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	render.JSON(w, r, feature)
}
