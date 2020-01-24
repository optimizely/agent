/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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
	"errors"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/optimizely/agent/pkg/middleware"
)

// UserOverrideBody defines the overrides to be applied.
type UserOverrideBody struct {
	VariationKey string `json:"variationKey"`
}

// UserOverrideHandler implements the UserAPI interface
type UserOverrideHandler struct{}

// SetForcedVariation - set a forced variation
func (h *UserOverrideHandler) SetForcedVariation(w http.ResponseWriter, r *http.Request) {
	optlyClient, optlyContext, err := parseContext(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}
	experimentKey := chi.URLParam(r, "experimentKey")
	if experimentKey == "" {
		RenderError(errors.New("empty experimentKey"), http.StatusBadRequest, w, r)
		return
	}

	override := &UserOverrideBody{}
	if err = ParseRequestBody(r, override); err != nil {
		RenderError(errors.New("empty variationKey"), http.StatusBadRequest, w, r)
		return
	}

	wasSet, err := optlyClient.SetForcedVariation(experimentKey, optlyContext.UserContext.ID, override.VariationKey)
	switch {
	case err != nil:
		middleware.GetLogger(r).Error().Err(err).Msg("error setting forced variation")
		RenderError(err, http.StatusInternalServerError, w, r)

	case wasSet:
		w.WriteHeader(http.StatusCreated)

	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

// RemoveForcedVariation - Remove a forced variation
func (h *UserOverrideHandler) RemoveForcedVariation(w http.ResponseWriter, r *http.Request) {
	optlyClient, optlyContext, err := parseContext(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}
	experimentKey := chi.URLParam(r, "experimentKey")
	if experimentKey == "" {
		RenderError(errors.New("empty experimentKey"), http.StatusBadRequest, w, r)
		return
	}

	err = optlyClient.RemoveForcedVariation(experimentKey, optlyContext.UserContext.ID)
	if err != nil {
		middleware.GetLogger(r).Error().Err(err).Msg("error removing forced variation")
		RenderError(err, http.StatusInternalServerError, w, r)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}
