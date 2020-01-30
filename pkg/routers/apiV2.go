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

// APIOptions defines the configuration parameters for Router.
type API2Options struct {
	maxConns        int
	middleware      middleware.OptlyMiddleware
	handlers        APIHandlers
	metricsRegistry *metrics.Registry
}

type APIHandlers interface {
	ListExperiments(w http.ResponseWriter, r *http.Request)
	GetExperiment(w http.ResponseWriter, r *http.Request)

	ListFeatures(w http.ResponseWriter, r *http.Request)
	GetFeature(w http.ResponseWriter, r *http.Request)

	Decide(w http.ResponseWriter, r *http.Request)
	DecideAll(w http.ResponseWriter, r *http.Request)

	TrackEvent(w http.ResponseWriter, r *http.Request)

	Override(w http.ResponseWriter, r *http.Request)
}

type DefaultHandlers struct{}

func (d DefaultHandlers) ListExperiments(w http.ResponseWriter, r *http.Request) {
	handlers.ListExperiments(w, r)
}

func (d DefaultHandlers) GetExperiment(w http.ResponseWriter, r *http.Request) {
	handlers.GetExperiment(w, r)
}

func (d DefaultHandlers) ListFeatures(w http.ResponseWriter, r *http.Request) {
	handlers.ListFeatures(w, r)
}

func (d DefaultHandlers) GetFeature(w http.ResponseWriter, r *http.Request) {
	handlers.GetFeature(w, r)
}

func (d DefaultHandlers) Decide(w http.ResponseWriter, r *http.Request) {
	handlers.Decide(w, r)
}

func (d DefaultHandlers) DecideAll(w http.ResponseWriter, r *http.Request) {
	handlers.Decide(w, r)
}

func (d DefaultHandlers) TrackEvent(w http.ResponseWriter, r *http.Request) {
	handlers.TrackEvent(w, r)
}

func (d DefaultHandlers) Override(w http.ResponseWriter, r *http.Request) {
	handlers.Override(w, r)
}

// NewDefaultAPIRouter creates a new router with the default backing optimizely.Cache
func NewDefaultAPIV2Router(optlyCache optimizely.Cache, conf config.APIConfig, metricsRegistry *metrics.Registry) http.Handler {
	spec := &API2Options{
		maxConns:        conf.MaxConns,
		middleware:      &middleware.CachedOptlyMiddleware{Cache: optlyCache},
		handlers:        new(DefaultHandlers),
		metricsRegistry: metricsRegistry,
	}

	return NewAPIV2Router(spec)
}

// NewAPIRouter returns HTTP API router backed by an optimizely.Cache implementation
func NewAPIV2Router(opt *API2Options) *chi.Mux {
	r := chi.NewRouter()

	listFeaturesTimer := middleware.Metricize("list-features", opt.metricsRegistry)
	getFeatureTimer := middleware.Metricize("get-feature", opt.metricsRegistry)
	listExperimentsTimer := middleware.Metricize("list-experiments", opt.metricsRegistry)
	getExperimentTimer := middleware.Metricize("get-experiment", opt.metricsRegistry)

	decideTimer := middleware.Metricize("decide", opt.metricsRegistry)
	decideAllTimer := middleware.Metricize("decide-all", opt.metricsRegistry)
	overrideTimer := middleware.Metricize("set-override", opt.metricsRegistry)
	trackTimer := middleware.Metricize("track-event", opt.metricsRegistry)

	if opt.maxConns > 0 {
		// Note this is NOT a rate limiter, but a concurrency threshold
		r.Use(chimw.Throttle(opt.maxConns))
	}

	r.Use(middleware.SetTime)
	r.Use(render.SetContentType(render.ContentTypeJSON), middleware.SetRequestID)

	r.Route("/features", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx)
		r.With(listFeaturesTimer).Get("/", opt.handlers.ListFeatures)
		r.With(getFeatureTimer, opt.middleware.FeatureCtx).Get("/{featureKey}", opt.handlers.GetFeature)
	})

	r.Route("/experiments", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx)
		r.With(listExperimentsTimer).Get("/", opt.handlers.ListExperiments)
		r.With(getExperimentTimer, opt.middleware.ExperimentCtx).Get("/{experimentKey}", opt.handlers.GetExperiment)
	})

	r.Route("/decide", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx)
		r.With(decideTimer).Get("/", opt.handlers.DecideAll)
		r.With(decideAllTimer, opt.middleware.FeatureCtx, opt.middleware.ExperimentCtx).Post("/{decisionKey}", opt.handlers.Decide)
	})

	r.With(trackTimer, opt.middleware.ClientCtx).Post("/track/{eventKey}", handlers.TrackEvent)
	r.With(overrideTimer, opt.middleware.ClientCtx).Post("/override/{decisionKey}", handlers.Override)

	return r
}
