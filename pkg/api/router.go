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
	"net/http"

	"github.com/optimizely/sidedoor/pkg/api/handlers"
	"github.com/optimizely/sidedoor/pkg/api/middleware"
	"github.com/optimizely/sidedoor/pkg/optimizely"
	"github.com/spf13/viper"

	"github.com/go-chi/chi"
	chimw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// RouterOptions defines the configuration parameters for Router.
type RouterOptions struct {
	maxConns     int
	middleware   middleware.OptlyMiddleware
	featureAPI   handlers.FeatureAPI
	userEventAPI handlers.UserEventAPI
	userAPI      handlers.UserAPI
}

// NewDefaultRouter creates a new router with the default backing optimizely.Cache
func NewDefaultRouter(optlyCache optimizely.Cache) http.Handler {
	spec := &RouterOptions{
		maxConns:     viper.GetInt("api.maxconns"),
		middleware:   &middleware.CachedOptlyMiddleware{Cache: optlyCache},
		featureAPI:   new(handlers.FeatureHandler),
		userEventAPI: new(handlers.UserEventHandler),
		userAPI:      new(handlers.UserHandler),
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

	r.With(chimw.AllowContentType("application/json"), middleware.Metricize("user-event")).Post("/user-event", opt.userEventAPI.AddUserEvent)

	r.Route("/features", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx)
		r.With(middleware.Metricize("list-features")).Get("/", opt.featureAPI.ListFeatures)
		r.With(middleware.Metricize("get-feature")).Get("/{featureKey}", opt.featureAPI.GetFeature)
	})

	r.Route("/users/{userID}", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx, opt.middleware.UserCtx)

		r.With(middleware.Metricize("track-event")).Post("/events/{eventKey}", opt.userAPI.TrackEvent)

		// TODO: Fix metrics key name, potentially change api method name, for symmetry
		r.With(middleware.Metricize("list-user-features")).Get("/features", opt.userAPI.ListFeatures)
		r.With(middleware.Metricize("get-user-feature")).Get("/features/{featureKey}", opt.userAPI.GetFeature)
		r.With(middleware.Metricize("track-user-features")).Post("/features", opt.userAPI.TrackFeatures)
		r.With(middleware.Metricize("track-user-feature")).Post("/features/{featureKey}", opt.userAPI.TrackFeature)
		r.With(middleware.Metricize("get-variation")).Get("/experiments/{experimentKey}", opt.userAPI.GetVariation)
		r.With(middleware.Metricize("activate-experiment")).Post("/experiments/{experimentKey}", opt.userAPI.ActivateExperiment)
		r.With(middleware.Metricize("set-forced-variation")).Put("/experiments/{experimentKey}/variations/{variationKey}", opt.userAPI.SetForcedVariation)
		r.With(middleware.Metricize("remove-forced-variation")).Delete("/experiments/{experimentKey}/variations", opt.userAPI.RemoveForcedVariation)
	})

	return r
}
