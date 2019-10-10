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

// Package webhook //
package webhook

import (
	"github.com/optimizely/sidedoor/pkg/admin"
	"github.com/optimizely/sidedoor/pkg/webhook/handlers"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// Version holds the webhook version
var Version string // can be set at compile time

// NewRouter returns HTTP API router
func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	optlyAdmin := admin.NewAdmin(Version, "DevX", "webhook")
	r.Use(optlyAdmin.AppInfoHeader)

	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Route("/admin", func(r chi.Router) {
		r.Get("/health", optlyAdmin.Health)
		r.Get("/info", optlyAdmin.AppInfo)
	})

	webhookAPI := new(handlers.OptlyWebhookHandler)

	r.Post("/webhooks/optimizely", webhookAPI.HandleWebhook)
	return r
}
