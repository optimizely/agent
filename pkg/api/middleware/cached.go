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
	"net/http"

	"github.com/optimizely/sidedoor/pkg/optimizely"
)

type contextKey string

// OptlyClientKey is the context key for the OptlyClient
const OptlyClientKey = contextKey("optlyClient")

// OptlySDKHeader is the header key for an ad-hoc SDK key
const OptlySDKHeader = "X-Optimizely-SDK-Key"

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

		var err error
		var optlyClient *optimizely.OptlyClient
		if sdkKey == "" {
			optlyClient, err = ctx.Cache.GetDefaultClient()
		} else {
			optlyClient, err = ctx.Cache.GetClient(sdkKey)
		}

		if err != nil {
			http.Error(w, "Failed to instantiate Optimizely", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), OptlyClientKey, optlyClient)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
