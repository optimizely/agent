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

// APIV1Options defines the configuration parameters for Router.
type APIV1Options struct {
	maxConns        int
	middleware      middleware.OptlyMiddleware
	handlers        apiHandlers
	metricsRegistry *metrics.Registry
}

// Define an interface to fasciliate testing
type apiHandlers interface {
	listExperiments(w http.ResponseWriter, r *http.Request)
	getExperiment(w http.ResponseWriter, r *http.Request)

	listFeatures(w http.ResponseWriter, r *http.Request)
	getFeature(w http.ResponseWriter, r *http.Request)

	decide(w http.ResponseWriter, r *http.Request)
	decideAll(w http.ResponseWriter, r *http.Request)

	trackEvent(w http.ResponseWriter, r *http.Request)

	override(w http.ResponseWriter, r *http.Request)
}

type defaultHandlers struct{}

func (d defaultHandlers) listExperiments(w http.ResponseWriter, r *http.Request) {
	handlers.ListExperiments(w, r)
}

func (d defaultHandlers) getExperiment(w http.ResponseWriter, r *http.Request) {
	handlers.GetExperiment(w, r)
}

func (d defaultHandlers) listFeatures(w http.ResponseWriter, r *http.Request) {
	handlers.ListFeatures(w, r)
}

func (d defaultHandlers) getFeature(w http.ResponseWriter, r *http.Request) {
	handlers.GetFeature(w, r)
}

func (d defaultHandlers) decide(w http.ResponseWriter, r *http.Request) {
	handlers.Activate(w, r)
}

func (d defaultHandlers) decideAll(w http.ResponseWriter, r *http.Request) {
	handlers.ActivateAll(w, r)
}

func (d defaultHandlers) trackEvent(w http.ResponseWriter, r *http.Request) {
	handlers.TrackEvent(w, r)
}

func (d defaultHandlers) override(w http.ResponseWriter, r *http.Request) {
	handlers.Override(w, r)
}

// NewDefaultAPIV1Router creates a new router with the default backing optimizely.Cache
func NewDefaultAPIV1Router(optlyCache optimizely.Cache, conf config.APIConfig, metricsRegistry *metrics.Registry) http.Handler {
	spec := &APIV1Options{
		maxConns:        conf.MaxConns,
		middleware:      &middleware.CachedOptlyMiddleware{Cache: optlyCache},
		handlers:        new(defaultHandlers),
		metricsRegistry: metricsRegistry,
	}

	return NewAPIV1Router(spec)
}

// NewAPIV1Router returns HTTP API router backed by an optimizely.Cache implementation
func NewAPIV1Router(opt *APIV1Options) *chi.Mux {
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

	r.Route("/v1/features", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx)
		r.With(listFeaturesTimer).Get("/", opt.handlers.listFeatures)
		r.With(getFeatureTimer, opt.middleware.FeatureCtx).Get("/{featureKey}", opt.handlers.getFeature)
	})

	r.Route("/v1/experiments", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx)
		r.With(listExperimentsTimer).Get("/", opt.handlers.listExperiments)
		r.With(getExperimentTimer, opt.middleware.ExperimentCtx).Get("/{experimentKey}", opt.handlers.getExperiment)
	})

	r.Route("/v1/activate", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx)
		r.With(decideTimer).Post("/", opt.handlers.decideAll)
		r.With(decideAllTimer, middleware.ActivationCtx).Post("/{decisionKey}", opt.handlers.decide)
	})

	r.With(trackTimer, opt.middleware.ClientCtx).Post("/v1/track/{eventKey}", handlers.TrackEvent)
	r.With(overrideTimer, opt.middleware.ClientCtx).Post("/v1/override/{decisionKey}", handlers.Override)

	return r
}
