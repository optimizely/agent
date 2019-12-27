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
	"github.com/optimizely/sidedoor/config"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/optimizely/sidedoor/pkg/optimizely"
)

// RouterOptions defines the configuration parameters for Router
type RouterOptions struct {
	cache          optimizely.Cache
	webhookConfigs config.WebhookConfig
}

// NewDefaultRouter creates a new router
func NewDefaultRouter(optlyCache optimizely.Cache) http.Handler {
	var conf config.WebhookConfig
	if err := viper.UnmarshalKey("webhook.configs", &conf); err != nil {
		log.Info().Msg("Unable to parse webhooks.")
	}

	spec := &RouterOptions{
		cache:          optlyCache,
		webhookConfigs: conf,
	}

	return NewRouter(spec)
}

// NewRouter returns HTTP API router
func NewRouter(opt *RouterOptions) *chi.Mux {
	r := chi.NewRouter()

	r.Use(render.SetContentType(render.ContentTypeJSON))
	webhookAPI := NewWebhookHandler(opt.cache, opt.webhookConfigs)

	r.Post("/webhooks/optimizely", webhookAPI.HandleWebhook)
	return r
}
