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

// Package routers //
package routers

import (
	"github.com/go-chi/chi"
	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/handlers"
)

// NewAPIRouter returns HTTP API router backed by an optimizely.Cache implementation
func NewOAuthRouter(conf *config.OAuthConfig, authConfigs []*config.ServiceAuthConfig) *chi.Mux {
	r := chi.NewRouter()
	handler := handlers.NewOAuthHandler(authConfigs, conf.HMACSecret)
	r.Get("/oauth/v2/api/accessToken", handler.GetAPIAccessToken)
	return r
}
