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
	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/handlers"
	"github.com/optimizely/agent/pkg/middleware"

	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// NewAdminRouter returns HTTP admin router
func NewAdminRouter(conf config.AgentConfig) http.Handler {
	r := chi.NewRouter()

	var authProvider middleware.Auth
	checkClaims := map[string]struct{}{"exp": {}, "admin": {}}
	if conf.Admin.Auth.HMACSecret == "" {
		authProvider = middleware.NewAuth(middleware.NoAuth{}, checkClaims)
	} else {
		authProvider = middleware.NewAuth(middleware.NewJWTVerifier(conf.Admin.Auth.HMACSecret), checkClaims)
	}

	optlyAdmin := handlers.NewAdmin(conf)
	tokenHandler := handlers.NewOAuthHandler(&conf.Admin.Auth)
	r.Use(optlyAdmin.AppInfoHeader)

	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.With(authProvider.Authorize).Get("/config", optlyAdmin.AppConfig)
	r.With(authProvider.Authorize).Get("/health", optlyAdmin.Health)
	r.With(authProvider.Authorize).Get("/info", optlyAdmin.AppInfo)
	r.With(authProvider.Authorize).Get("/metrics", optlyAdmin.Metrics)

	r.Post("/oauth/token", tokenHandler.GetAdminAccessToken)
	return r
}
