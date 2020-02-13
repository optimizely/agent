/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                        *
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
	"net/http"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/handlers"
	"github.com/optimizely/agent/pkg/metrics"
	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"

	"github.com/go-chi/chi"
	chimw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// ClientOptions defines the configuration parameters for Router.
type ClientOptions struct {
	maxConns        int
	middleware      middleware.OptlyMiddleware
	handlers        apiHandlers
	metricsRegistry *metrics.Registry
}

// Define an interface to fasciliate testing
type apiHandlers interface {
	config(w http.ResponseWriter, r *http.Request)
	activate(w http.ResponseWriter, r *http.Request)
	trackEvent(w http.ResponseWriter, r *http.Request)
	override(w http.ResponseWriter, r *http.Request)
}

type defaultHandlers struct{}

func (d defaultHandlers) config(w http.ResponseWriter, r *http.Request) {
	handlers.OptimizelyConfig(w, r)
}

func (d defaultHandlers) activate(w http.ResponseWriter, r *http.Request) {
	handlers.Activate(w, r)
}

func (d defaultHandlers) trackEvent(w http.ResponseWriter, r *http.Request) {
	handlers.TrackEvent(w, r)
}

func (d defaultHandlers) override(w http.ResponseWriter, r *http.Request) {
	handlers.Override(w, r)
}

// NewDefaultClientRouter creates a new router with the default backing optimizely.Cache
func NewDefaultClientRouter(optlyCache optimizely.Cache, conf config.ClientConfig, metricsRegistry *metrics.Registry) http.Handler {
	spec := &ClientOptions{
		maxConns:        conf.MaxConns,
		middleware:      &middleware.CachedOptlyMiddleware{Cache: optlyCache},
		handlers:        new(defaultHandlers),
		metricsRegistry: metricsRegistry,
	}

	return NewClientRouter(spec)
}

// NewClientRouter returns HTTP Client router backed by an optimizely.Cache implementation
func NewClientRouter(opt *ClientOptions) *chi.Mux {
	r := chi.NewRouter()

	getConfigTimer := middleware.Metricize("get-config", opt.metricsRegistry)
	activateTimer := middleware.Metricize("activate", opt.metricsRegistry)
	overrideTimer := middleware.Metricize("override", opt.metricsRegistry)
	trackTimer := middleware.Metricize("track-event", opt.metricsRegistry)

	if opt.maxConns > 0 {
		// Note this is NOT a rate limiter, but a concurrency threshold
		r.Use(chimw.Throttle(opt.maxConns))
	}

	r.Use(middleware.SetTime, opt.middleware.ClientCtx)
	r.Use(render.SetContentType(render.ContentTypeJSON), middleware.SetRequestID)

	r.With(getConfigTimer).Get("/v1/config", opt.handlers.config)
	r.With(activateTimer).Post("/v1/activate", opt.handlers.activate)
	r.With(trackTimer).Post("/v1/track", opt.handlers.trackEvent)
	r.With(overrideTimer).Post("/v1/override", opt.handlers.override)

	return r
}
