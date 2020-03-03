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
	"github.com/go-chi/render"
	"net/http"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/handlers"
	"github.com/optimizely/agent/pkg/metrics"
	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"

	"github.com/go-chi/chi"
	chimw "github.com/go-chi/chi/middleware"
)

// APIOptions defines the configuration parameters for Router.
type APIOptions struct {
	maxConns         int
	enableOverrides  bool
	middleware       middleware.OptlyMiddleware
	experimentAPI    handlers.ExperimentAPI
	featureAPI       handlers.FeatureAPI
	userAPI          handlers.UserAPI
	notificationsAPI handlers.NotificationAPI
	userOverrideAPI  handlers.UserOverrideAPI
	metricsRegistry  *metrics.Registry
	oAuthHandler     *handlers.OAuthHandler
	oAuthMiddleware  middleware.Auth
}

// NewDefaultAPIRouter creates a new router with the default backing optimizely.Cache
func NewDefaultAPIRouter(optlyCache optimizely.Cache, conf config.APIConfig, metricsRegistry *metrics.Registry) http.Handler {

	authProvider := middleware.NewAuth(&conf.Auth)
	if _, ok := authProvider.Verifier.(middleware.BadAuth); ok {
		return nil
	}

	var notificationsAPI handlers.NotificationAPI
	notificationsAPI = handlers.NewDisabledNotificationHandler()
	if conf.EnableNotifications {
		notificationsAPI = handlers.NewNotificationHandler()
	}

	var userOverrideAPI handlers.UserOverrideAPI
	userOverrideAPI = new(handlers.DisabledUserOverrideHandler)
	if conf.EnableOverrides {
		userOverrideAPI = new(handlers.UserOverrideHandler)
	}

	spec := &APIOptions{
		maxConns:         conf.MaxConns,
		middleware:       &middleware.CachedOptlyMiddleware{Cache: optlyCache},
		experimentAPI:    new(handlers.ExperimentHandler),
		featureAPI:       new(handlers.FeatureHandler),
		userAPI:          new(handlers.UserHandler),
		notificationsAPI: notificationsAPI,
		userOverrideAPI:  userOverrideAPI,
		metricsRegistry:  metricsRegistry,
		oAuthHandler:     handlers.NewOAuthHandler(&conf.Auth),
		oAuthMiddleware:  authProvider,
		enableOverrides:  conf.EnableOverrides,
	}

	return NewAPIRouter(spec)
}

func setMiddleWareTime(r chi.Router) {
	r.Use(middleware.SetTime)
	r.Use(render.SetContentType(render.ContentTypeJSON), middleware.SetRequestID)
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

	r.Route("/notifications/event-stream", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx, opt.oAuthMiddleware.AuthorizeAPI)
		r.Get("/", opt.notificationsAPI.HandleEventSteam)
	})

	r.Route("/features", func(r chi.Router) {
		setMiddleWareTime(r)
		r.Use(opt.middleware.ClientCtx, opt.oAuthMiddleware.AuthorizeAPI)
		r.With(listFeaturesTimer).Get("/", opt.featureAPI.ListFeatures)
		r.With(getFeatureTimer, opt.middleware.FeatureCtx).Get("/{featureKey}", opt.featureAPI.GetFeature)
	})

	r.Route("/experiments", func(r chi.Router) {
		setMiddleWareTime(r)
		r.Use(opt.middleware.ClientCtx, opt.oAuthMiddleware.AuthorizeAPI)
		r.With(listExperimentsTimer).Get("/", opt.experimentAPI.ListExperiments)
		r.With(getExperimentTimer, opt.middleware.ExperimentCtx).Get("/{experimentKey}", opt.experimentAPI.GetExperiment)
	})

	r.Route("/users/{userID}", func(r chi.Router) {
		setMiddleWareTime(r)

		r.Use(opt.middleware.ClientCtx, opt.middleware.UserCtx, opt.oAuthMiddleware.AuthorizeAPI)

		r.With(trackEventTimer).Post("/events/{eventKey}", opt.userAPI.TrackEvent)

		r.With(listUserFeaturesTimer).Get("/features", opt.userAPI.ListFeatures)
		r.With(trackUserFeaturesTimer).Post("/features", opt.userAPI.TrackFeatures)
		r.With(getUserFeatureTimer, opt.middleware.FeatureCtx).Get("/features/{featureKey}", opt.userAPI.GetFeature)
		r.With(trackUserFeatureTimer, opt.middleware.FeatureCtx).Post("/features/{featureKey}", opt.userAPI.TrackFeature)
		r.With(getVariationTimer, opt.middleware.ExperimentCtx).Get("/experiments/{experimentKey}", opt.userAPI.GetVariation)
		r.With(activateExperimentTimer, opt.middleware.ExperimentCtx).Post("/experiments/{experimentKey}", opt.userAPI.ActivateExperiment)
	})

	r.Route("/overrides/users/{userID}", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx, opt.middleware.UserCtx, opt.oAuthMiddleware.AuthorizeAPI)

		r.With(setForcedVariationTimer).Put("/experiments/{experimentKey}", opt.userOverrideAPI.SetForcedVariation)
		r.With(removeForcedVariationTimer).Delete("/experiments/{experimentKey}", opt.userOverrideAPI.RemoveForcedVariation)
	})

	r.Post("/oauth/token", opt.oAuthHandler.CreateAPIAccessToken)

	// Kind of a hack to support concurrent APIs without a larger refactor.
	spec := &APIV1Options{
		maxConns:        opt.maxConns,
		enableOverrides: opt.enableOverrides,
		middleware:      opt.middleware,
		handlers:        new(defaultHandlers),
		metricsRegistry: opt.metricsRegistry,
		oAuthHandler:    opt.oAuthHandler,
		oAuthMiddleware: opt.oAuthMiddleware,
	}

	r.Group(func(r chi.Router) { WithAPIV1Router(spec, r) })

	return r
}
