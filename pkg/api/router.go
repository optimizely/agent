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

// Package api //
package api

import (
	"github.com/optimizely/sidedoor/config"
	"net/http"

	"github.com/go-chi/chi"
	chimw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/optimizely/sidedoor/pkg/api/handlers"
	"github.com/optimizely/sidedoor/pkg/api/middleware"
	"github.com/optimizely/sidedoor/pkg/optimizely"
)

var listFeaturesTimer func(http.Handler) http.Handler
var getFeatureTimer func(http.Handler) http.Handler
var listExperimentsTimer func(http.Handler) http.Handler
var getExperimentTimer func(http.Handler) http.Handler
var trackEventTimer func(http.Handler) http.Handler
var listUserFeaturesTimer func(http.Handler) http.Handler
var trackUserFeaturesTimer func(http.Handler) http.Handler
var getUserFeatureTimer func(http.Handler) http.Handler
var trackUserFeatureTimer func(http.Handler) http.Handler
var getVariationTimer func(http.Handler) http.Handler
var activateExperimentTimer func(http.Handler) http.Handler
var setForcedVariationTimer func(http.Handler) http.Handler
var removeForcedVariationTimer func(http.Handler) http.Handler

func init() {
	listFeaturesTimer = middleware.Metricize("list-features")
	getFeatureTimer = middleware.Metricize("get-feature")
	listExperimentsTimer = middleware.Metricize("list-experiments")
	getExperimentTimer = middleware.Metricize("get-experiment")
	trackEventTimer = middleware.Metricize("track-event")
	listUserFeaturesTimer = middleware.Metricize("list-user-features")
	trackUserFeaturesTimer = middleware.Metricize("track-user-features")
	getUserFeatureTimer = middleware.Metricize("get-user-feature")
	trackUserFeatureTimer = middleware.Metricize("track-user-feature")
	getVariationTimer = middleware.Metricize("get-variation")
	activateExperimentTimer = middleware.Metricize("activate-experiment")
	setForcedVariationTimer = middleware.Metricize("set-forced-variation")
	removeForcedVariationTimer = middleware.Metricize("remove-forced-variation")
}

// RouterOptions defines the configuration parameters for Router.
type RouterOptions struct {
	maxConns      int
	middleware    middleware.OptlyMiddleware
	experimentAPI handlers.ExperimentAPI
	featureAPI    handlers.FeatureAPI
	userAPI       handlers.UserAPI
}

// NewDefaultRouter creates a new router with the default backing optimizely.Cache
func NewDefaultRouter(optlyCache optimizely.Cache, conf config.APIConfig) http.Handler {
	spec := &RouterOptions{
		maxConns:      conf.MaxConns,
		middleware:    &middleware.CachedOptlyMiddleware{Cache: optlyCache},
		experimentAPI: new(handlers.ExperimentHandler),
		featureAPI:    new(handlers.FeatureHandler),
		userAPI:       new(handlers.UserHandler),
	}

	return NewRouter(spec)
}

// NewRouter returns HTTP API router backed by an optimizely.Cache implementation
func NewRouter(opt *RouterOptions) *chi.Mux {
	r := chi.NewRouter()

	if opt.maxConns > 0 {
		// Note this is NOT a rate limiter, but a concurrency threshold
		r.Use(chimw.Throttle(opt.maxConns))
	}

	r.Use(middleware.SetTime)
	r.Use(render.SetContentType(render.ContentTypeJSON), middleware.SetRequestID)

	r.Route("/features", func(r chi.Router) {
		r.With(listFeaturesTimer, opt.middleware.ClientCtx).Get("/", opt.featureAPI.ListFeatures)
		r.With(getFeatureTimer, opt.middleware.FeatureCtx).Get("/{featureKey}", opt.featureAPI.GetFeature)
	})

	r.Route("/experiments", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx)
		r.With(listExperimentsTimer).Get("/", opt.experimentAPI.ListExperiments)
		r.With(getExperimentTimer).Get("/{experimentKey}", opt.experimentAPI.GetExperiment)
	})

	r.Route("/users/{userID}", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx, opt.middleware.UserCtx)

		r.With(trackEventTimer).Post("/events/{eventKey}", opt.userAPI.TrackEvent)

		r.With(listUserFeaturesTimer).Get("/features", opt.userAPI.ListFeatures)
		r.With(trackUserFeaturesTimer).Post("/features", opt.userAPI.TrackFeatures)
		r.With(getUserFeatureTimer).Get("/features/{featureKey}", opt.userAPI.GetFeature)
		r.With(trackUserFeatureTimer).Post("/features/{featureKey}", opt.userAPI.TrackFeature)
		r.With(getVariationTimer).Get("/experiments/{experimentKey}", opt.userAPI.GetVariation)
		r.With(activateExperimentTimer).Post("/experiments/{experimentKey}", opt.userAPI.ActivateExperiment)
		r.With(setForcedVariationTimer).Put("/experiments/{experimentKey}/variations/{variationKey}", opt.userAPI.SetForcedVariation)
		r.With(removeForcedVariationTimer).Delete("/experiments/{experimentKey}/variations", opt.userAPI.RemoveForcedVariation)
	})

	return r
}
