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

// Package middleware //
package middleware

import (
	"context"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/optimizely/sidedoor/pkg/optimizely"
	"net/http"
)

// FeatureCtx extracts the featureKey URL param and adds a Feature to the request context.
// If no such feature exists in the current config, returns 404
// If no OptlyClient client is available, returns 500
// Note: This middleware has two dependencies:
//	- ClientCtx middleware should be running prior to this one
//	- featureKey must be available as a URL param
func FeatureCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		optlyClient, err := GetOptlyClient(r)
		if err != nil {
			RenderError(fmt.Errorf("optlyClient not available in FeatureCtx"), http.StatusInternalServerError, w, r)
			return
		}

		featureKey := chi.URLParam(r, "featureKey")
		if featureKey == "" {
			RenderError(fmt.Errorf("invalid request, missing featureKey"), http.StatusBadRequest, w, r)
			return
		}

		feature, err := optlyClient.GetFeature(featureKey)
		var statusCode int
		switch err {
		case nil:
			GetLogger(r).Debug().Str("featureKey", featureKey).Msg("Added feature to request context")
			ctx := context.WithValue(r.Context(), OptlyFeatureKey, &feature)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		case optimizely.ErrFeatureNotFound:
			statusCode = http.StatusNotFound
		default:
			statusCode = http.StatusInternalServerError
		}
		GetLogger(r).Error().Err(err).Str("featureKey", featureKey).Msg("Calling GetFeature")
		RenderError(err, statusCode, w, r)
	})
}
