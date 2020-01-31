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
package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
)

// ActivationCtx attaches either a feature or a experiment to the context for activation
// TODO extract the "track" query parameter
func ActivationCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		optlyClient, err := GetOptlyClient(r)
		logger := GetLogger(r)
		if err != nil {
			RenderError(fmt.Errorf("optlyClient not available in ActivationCtx"), http.StatusInternalServerError, w, r)
			return
		}

		activationKey := chi.URLParam(r, "activationKey")
		if activationKey == "" {
			log.Debug().Msg("no activationKey provided")
			RenderError(fmt.Errorf("invalid request, missing activationKey in ActivationCtx"), http.StatusBadRequest, w, r)
			return
		}

		oConf := optlyClient.GetOptimizelyConfig()

		if f, ok := oConf.FeaturesMap[activationKey]; ok {
			logger.Debug().Str("featureKey", activationKey).Msg("Added feature to request context in ActivationCtx")
			ctx := context.WithValue(r.Context(), OptlyFeatureKey, &f)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if e, ok := oConf.ExperimentsMap[activationKey]; ok {
			logger.Debug().Str("experimentKey", activationKey).Msg("Added experiment to request context in ActivationCtx")
			ctx := context.WithValue(r.Context(), OptlyExperimentKey, &e)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		RenderError(fmt.Errorf("unable to find entity for key %s", activationKey), http.StatusNotFound, w, r)
	})
}
