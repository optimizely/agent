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
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"

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

// UserCtx extracts the userId and any associated attributes from the request
// to create an optimizely.UserContext which will be used by downstream handlers.
// Future iterations of this middleware would capture pulling additional
// detail from a UPS or attribute store.
func (mw *CachedOptlyMiddleware) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			RenderError(fmt.Errorf("invalid request, missing userId"), http.StatusBadRequest, w, r)
			return
		}

		// Remove userId and copy values into the attributes map
		logger := GetLogger(r)
		values := r.URL.Query()
		attributes := make(map[string]interface{})

		for k, v := range values {
			// Assuming a single KV pair exists in the query parameters
			val := v[0]
			attributes[k] = CoerceType(val)
			logger.Debug().Str("attrKey", k).Str("attrVal", val).Msg("User attribute.")
		}

		optlyContext := optimizely.NewContext(userID, attributes)
		ctx := context.WithValue(r.Context(), OptlyContextKey, optlyContext)

		logger.Debug().Str("userId", userID).Msg("Adding user context to request.")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// FeatureCtx extracts the featureKey URL param and adds an optimizelyconfig.OptimizelyFeature to the request context.
// If no such feature exists in the current config, returns 404
// Note: featureKey must be available as a URL param, and ClientCtx middleware must run prior to this middleware
func (mw *CachedOptlyMiddleware) FeatureCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		optlyClient, err := GetOptlyClient(r)
		logger := GetLogger(r)
		if err != nil {
			RenderError(fmt.Errorf("optlyClient not available in FeatureCtx"), http.StatusInternalServerError, w, r)
			return
		}

		featureKey := getDecisionKey(r, "featureKey")
		if featureKey == "" {
			log.Debug().Msg("no featureKey provided")
			next.ServeHTTP(w, r)
			return
		}

		feature, err := optlyClient.GetFeature(featureKey)
		var statusCode int
		switch {
		case err == nil:
			logger.Debug().Str("featureKey", featureKey).Msg("Added feature to request context in FeatureCtx")
			ctx := context.WithValue(r.Context(), OptlyFeatureKey, &feature)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		case errors.Is(err, optimizely.ErrEntityNotFound):
			statusCode = http.StatusNotFound
			return
		default:
			statusCode = http.StatusInternalServerError
		}
		logger.Debug().Err(err).Str("featureKey", featureKey).Msg("Calling GetFeature in FeatureCtx")
		RenderError(err, statusCode, w, r)
	})
}

// ExperimentCtx extracts the experimentKey URL param and adds a optimizelyconfig.OptimizelyExperiment to the request context.
// If no such experiment exists in the current config, returns 404
// Note: experimentKey must be available as a URL param, and ClientCtx middleware must run prior to this middleware
func (mw *CachedOptlyMiddleware) ExperimentCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		optlyClient, err := GetOptlyClient(r)
		logger := GetLogger(r)
		if err != nil {
			RenderError(fmt.Errorf("optlyClient not available in ExperimentCtx"), http.StatusInternalServerError, w, r)
			return
		}

		experimentKey := getDecisionKey(r, "experimentKey")
		if experimentKey == "" {
			log.Debug().Msg("no experimentKey provided")
			RenderError(fmt.Errorf("invalid request, missing experimentKey in ExperimentCtx"), http.StatusBadRequest, w, r)
			return
		}

		logger.Debug().Msg("looking for experiment")
		experiment, err := optlyClient.GetExperiment(experimentKey)
		var statusCode int
		switch {
		case err == nil:
			logger.Debug().Str("experimentKey", experimentKey).Msg("Added experiment to request context in ExperimentCtx")
			ctx := context.WithValue(r.Context(), OptlyExperimentKey, &experiment)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		case errors.Is(err, optimizely.ErrEntityNotFound):
			//statusCode = http.StatusNotFound
			next.ServeHTTP(w, r)
			return
		default:
			statusCode = http.StatusInternalServerError
		}
		logger.Debug().Err(err).Str("experimentKey", experimentKey).Msg("Calling GetExperiment in ExperimentCtx")
		RenderError(err, statusCode, w, r)
	})
}

func getDecisionKey(r *http.Request, key string) string {
	decisionKey := chi.URLParam(r, "decisionKey")
	if decisionKey != "" {
		return decisionKey
	}

	return chi.URLParam(r, key)
}
