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

// Package admin //
package admin

import (
	"github.com/optimizely/sidedoor/pkg/admin/handlers"
	"github.com/spf13/viper"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// Version holds the admin version
var Version string // can be set at compile time

// NewRouter returns HTTP admin router
func NewRouter(srvcs []handlers.HealthChecker) *chi.Mux {
	r := chi.NewRouter()

	version := viper.GetString("app.version")
	author := viper.GetString("app.author")
	appName := viper.GetString("app.name")

	optlyAdmin := handlers.NewAdmin(version, author, appName, srvcs)
	r.Use(optlyAdmin.AppInfoHeader)

	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/health", optlyAdmin.Health)
	r.Get("/info", optlyAdmin.AppInfo)
	r.Get("/metrics", optlyAdmin.Metrics)

	return r
}
