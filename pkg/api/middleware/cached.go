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

	"github.com/go-chi/chi"
	"github.com/optimizely/sidedoor/pkg/optimizely"
)

type contextKey string

// OptlyClientKey is the context key for the OptlyClient
const OptlyClientKey = contextKey("optlyClient")

// OptlyContextKey is the context key for the OptlyContext
const OptlyContextKey = contextKey("optlyContext")

// OptlySDKHeader is the header key for an ad-hoc SDK key
const OptlySDKHeader = "X-Optimizely-SDK-Key"

// OptlyRequestHeader is the header key for the request ID
const OptlyRequestHeader = "X-Request-Id"

// CachedOptlyMiddleware implements OptlyMiddleware backed by a cache
type CachedOptlyMiddleware struct {
	Cache optimizely.Cache
}

// ClientCtx adds a pointer to an OptlyClient to the request context.
// Precedence is given for any SDK key provided within the request header
// else the default OptlyClient will be used.
func (ctx *CachedOptlyMiddleware) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sdkKey := r.Header.Get(OptlySDKHeader)
		if sdkKey == "" {
			RenderError(fmt.Errorf("missing required %s header", OptlySDKHeader), http.StatusBadRequest, w, r)
			return
		}

		optlyClient, err := ctx.Cache.GetClient(sdkKey)
		if err != nil {
			GetLogger(r).Error().Err(err).Msg("Initializing OptimizelyClient")
			RenderError(fmt.Errorf("failed to instantiate Optimizely for SDK Key: %s", sdkKey), http.StatusInternalServerError, w, r)
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
func (ctx *CachedOptlyMiddleware) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userID := chi.URLParam(r, "userID")
		if userID == "" {
			RenderError(fmt.Errorf("invalid request, missing userId"), http.StatusBadRequest, w, r)
			return
		}

		// Remove userId and copy values into the attributes map
		values := r.URL.Query()
		attributes := make(map[string]interface{})
		for k, v := range values {
			// Assuming a single KV pair exists in the query parameters
			attributes[k] = v[0]
			GetLogger(r).Debug().Str("attrKey", k).Str("attrVal", v[0]).Msg("User attribute.")
		}

		optlyContext := optimizely.NewContext(userID, attributes)
		ctx := context.WithValue(r.Context(), OptlyContextKey, optlyContext)

		GetLogger(r).Debug().Str("userId", userID).Msg("Adding user context to request.")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
