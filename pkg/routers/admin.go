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
	"github.com/optimizely/sidedoor/config"
	"github.com/optimizely/sidedoor/pkg/handlers"

	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// NewAdminRouter returns HTTP admin router
func NewAdminRouter(conf config.AdminConfig) http.Handler {
	r := chi.NewRouter()

	optlyAdmin := handlers.NewAdmin(conf.Version, conf.Author, conf.Name)
	r.Use(optlyAdmin.AppInfoHeader)

	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/health", optlyAdmin.Health)
	r.Get("/info", optlyAdmin.AppInfo)
	r.Get("/metrics", optlyAdmin.Metrics)

	return r
}
