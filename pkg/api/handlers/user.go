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
	"github.com/optimizely/sidedoor/pkg/api/models"
)

// UserHandler implements the UserAPI interface
type UserHandler struct{}

// ActivateFeature - Return the feature and record impression
func (h *UserHandler) ActivateFeature(w http.ResponseWriter, r *http.Request) {
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
	enabled, variables, err := optlyClient.GetAndTrackFeatureWithContext(featureKey, optlyContext)

	if err != nil {
		middleware.GetLogger(r).Error().Str("featureKey", featureKey).Str("userID", optlyContext.GetUserID()).Msg("Calling ActivateFeature")
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
