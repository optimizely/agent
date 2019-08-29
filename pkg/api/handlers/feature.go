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

	"github.com/optimizely/sidedoor/pkg/api/models"
	"github.com/optimizely/sidedoor/pkg/optimizely"
)

// ActivateFeature - Return the feature and record impression
func ActivateFeature(w http.ResponseWriter, r *http.Request) {
	featureKey := chi.URLParam(r, "featureKey")
	userID := r.URL.Query().Get("userId")

	if userID == "" {

		render.JSON(w, r, render.M{
			"error": "missing userID",
		})
		return
	}

	context := optimizely.NewContext(userID, map[string]interface{}{})

	enabled, err := context.IsFeatureEnabled(featureKey)

	if err != nil {
		render.JSON(w, r, render.M{
			"error": err,
		})
		return
	}

	feature := &models.Feature{
		Enabled: enabled,
		Key:     featureKey,
	}

	render.JSON(w, r, feature)
}
