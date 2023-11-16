/****************************************************************************
 * Copyright 2019-2020,2023, Optimizely, Inc. and contributors              *
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
	"context"
	"fmt"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/handlers"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/optimizely/agent/pkg/optimizely"
)

// NewWebhookRouter returns HTTP API router
func NewWebhookRouter(ctx context.Context, optlyCache optimizely.Cache, conf config.AgentConfig) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.AllowContentType("application/json"))
	r.Use(render.SetContentType(render.ContentTypeJSON))
	webhookAPI := handlers.NewWebhookHandler(optlyCache, conf.Webhook.Projects, conf.Synchronization)
	if conf.Synchronization.Datafile.Enable {
		if err := webhookAPI.StartSyncer(ctx); err != nil {
			fmt.Errorf("failed to start datafile syncer: %s", err.Error())
		}
	}

	r.Post("/webhooks/optimizely", webhookAPI.HandleWebhook)
	return r
}
