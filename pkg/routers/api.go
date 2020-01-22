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
type APIOptions struct {
	maxConns        int
	middleware      middleware.OptlyMiddleware
	experimentAPI   handlers.ExperimentAPI
	featureAPI      handlers.FeatureAPI
	userAPI         handlers.UserAPI
	userOverrideAPI handlers.UserOverrideAPI
	metricsRegistry *metrics.Registry
}

// NewDefaultAPIRouter creates a new router with the default backing optimizely.Cache
func NewDefaultAPIRouter(optlyCache optimizely.Cache, conf config.APIConfig, metricsRegistry *metrics.Registry) http.Handler {
	spec := &APIOptions{
		maxConns:        conf.MaxConns,
		middleware:      &middleware.CachedOptlyMiddleware{Cache: optlyCache},
		experimentAPI:   new(handlers.ExperimentHandler),
		featureAPI:      new(handlers.FeatureHandler),
		userAPI:         new(handlers.UserHandler),
		userOverrideAPI: new(handlers.UserOverrideHandler),
		metricsRegistry: metricsRegistry,
	}

	return NewAPIRouter(spec)
}

// NewAPIRouter returns HTTP API router backed by an optimizely.Cache implementation
func NewAPIRouter(opt *APIOptions) *chi.Mux {
	r := chi.NewRouter()

	listFeaturesTimer := middleware.Metricize("list-features", opt.metricsRegistry)
	getFeatureTimer := middleware.Metricize("get-feature", opt.metricsRegistry)
	listExperimentsTimer := middleware.Metricize("list-experiments", opt.metricsRegistry)
	getExperimentTimer := middleware.Metricize("get-experiment", opt.metricsRegistry)
	trackEventTimer := middleware.Metricize("track-event", opt.metricsRegistry)
	listUserFeaturesTimer := middleware.Metricize("list-user-features", opt.metricsRegistry)
	trackUserFeaturesTimer := middleware.Metricize("track-user-features", opt.metricsRegistry)
	getUserFeatureTimer := middleware.Metricize("get-user-feature", opt.metricsRegistry)
	trackUserFeatureTimer := middleware.Metricize("track-user-feature", opt.metricsRegistry)
	getVariationTimer := middleware.Metricize("get-variation", opt.metricsRegistry)
	activateExperimentTimer := middleware.Metricize("activate-experiment", opt.metricsRegistry)
	setForcedVariationTimer := middleware.Metricize("set-forced-variation", opt.metricsRegistry)
	removeForcedVariationTimer := middleware.Metricize("remove-forced-variation", opt.metricsRegistry)

	if opt.maxConns > 0 {
		// Note this is NOT a rate limiter, but a concurrency threshold
		r.Use(chimw.Throttle(opt.maxConns))
	}

	r.Use(middleware.SetTime)
	r.Use(render.SetContentType(render.ContentTypeJSON), middleware.SetRequestID)

	r.Route("/features", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx)
		r.With(listFeaturesTimer).Get("/", opt.featureAPI.ListFeatures)
		r.With(getFeatureTimer, opt.middleware.FeatureCtx).Get("/{featureKey}", opt.featureAPI.GetFeature)
	})

	r.Route("/experiments", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx)
		r.With(listExperimentsTimer).Get("/", opt.experimentAPI.ListExperiments)
		r.With(getExperimentTimer, opt.middleware.ExperimentCtx).Get("/{experimentKey}", opt.experimentAPI.GetExperiment)
	})

	r.Route("/users/{userID}", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx, opt.middleware.UserCtx)

		r.With(trackEventTimer).Post("/events/{eventKey}", opt.userAPI.TrackEvent)

		r.With(listUserFeaturesTimer).Get("/features", opt.userAPI.ListFeatures)
		r.With(trackUserFeaturesTimer).Post("/features", opt.userAPI.TrackFeatures)
		r.With(getUserFeatureTimer, opt.middleware.FeatureCtx).Get("/features/{featureKey}", opt.userAPI.GetFeature)
		r.With(trackUserFeatureTimer, opt.middleware.FeatureCtx).Post("/features/{featureKey}", opt.userAPI.TrackFeature)
		r.With(getVariationTimer, opt.middleware.ExperimentCtx).Get("/experiments/{experimentKey}", opt.userAPI.GetVariation)
		r.With(activateExperimentTimer, opt.middleware.ExperimentCtx).Post("/experiments/{experimentKey}", opt.userAPI.ActivateExperiment)
	})

	r.Route("/overrides/users/{userID}", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx, opt.middleware.UserCtx)

		r.With(setForcedVariationTimer).Put("/experiments/{experimentKey}/variations/{variationKey}", opt.userOverrideAPI.SetForcedVariation)
		r.With(removeForcedVariationTimer).Delete("/experiments/{experimentKey}/variations", opt.userOverrideAPI.RemoveForcedVariation)
	})

	return r
}
