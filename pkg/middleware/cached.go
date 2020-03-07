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
	"net/http"
	"strings"

	"github.com/optimizely/agent/pkg/optimizely"
)

type contextKey string

// OptlyClientKey is the context key for the OptlyClient
const OptlyClientKey = contextKey("optlyClient")

// OptlyContextKey is the context key for the OptlyContext
const OptlyContextKey = contextKey("optlyContext")

// OptlyFeatureKey is the context key used by FeatureCtx for setting a Feature
const OptlyFeatureKey = contextKey("featureKey")

// OptlyExperimentKey is the context key used by ExperimentCtx for setting an Experiment
const OptlyExperimentKey = contextKey("experimentKey")

// OptlySDKHeader is the header key for an ad-hoc SDK key
const OptlySDKHeader = "X-Optimizely-SDK-Key"

// CachedOptlyMiddleware implements OptlyMiddleware backed by a cache
type CachedOptlyMiddleware struct {
	Cache optimizely.Cache
}

// ClientCtx adds a pointer to an OptlyClient to the request context.
// Precedence is given for any SDK key provided within the request header
// else the default OptlyClient will be used.
func (mw *CachedOptlyMiddleware) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sdkKey := r.Header.Get(OptlySDKHeader)
		if sdkKey == "" {
			RenderError(fmt.Errorf("missing required %s header", OptlySDKHeader), http.StatusBadRequest, w, r)
			return
		}

		optlyClient, err := mw.Cache.GetClient(sdkKey)
		if err != nil {
			GetLogger(r).Error().Err(err).Msg("Initializing OptimizelyClient")

			// Check if error indicates a 403 from the CDN. Ideally we'd use errors.Is(), but the go-sdk isn't 1.13
			if strings.Contains(err.Error(), "403") {
				RenderError(err, http.StatusForbidden, w, r)
			} else {
				RenderError(fmt.Errorf("failed to instantiate Optimizely for SDK Key: %s", sdkKey), http.StatusInternalServerError, w, r)
			}

			return
		}

		ctx := context.WithValue(r.Context(), OptlyClientKey, optlyClient)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
