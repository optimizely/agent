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
	"expvar"

	"github.com/optimizely/sidedoor/pkg/api/handlers"
	"github.com/optimizely/sidedoor/pkg/api/middleware"
	"github.com/optimizely/sidedoor/pkg/optimizely"

	"github.com/go-chi/chi"
	chimw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// RouterOptions defines the configuration parameters for Router.
type RouterOptions struct {
	middleware   middleware.OptlyMiddleware
	featureAPI   handlers.FeatureAPI
	userEventAPI handlers.UserEventAPI
	userAPI      handlers.UserAPI
}

// NewDefaultRouter creates a new router with the default backing optimizely.Cache
func NewDefaultRouter(optlyCache optimizely.Cache) *chi.Mux {
	spec := &RouterOptions{
		middleware:   &middleware.CachedOptlyMiddleware{Cache: optlyCache},
		featureAPI:   new(handlers.FeatureHandler),
		userEventAPI: new(handlers.UserEventHandler),
		userAPI:      new(handlers.UserHandler),
	}

	return NewRouter(spec)
}

const metricsPrefix = "route_counters"

var routeCounts = expvar.NewMap(metricsPrefix)

// NewRouter returns HTTP API router backed by an optimizely.Cache implementation
func NewRouter(opt *RouterOptions) *chi.Mux {
	r := chi.NewRouter()

	r.Use(render.SetContentType(render.ContentTypeJSON), middleware.SetRequestID)

	r.With(chimw.AllowContentType("application/json"), middleware.HitCount(routeCounts)).Post("/user-event", opt.userEventAPI.AddUserEvent)

	r.Route("/features", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx)
		r.With(middleware.HitCount(routeCounts)).Get("/", opt.featureAPI.ListFeatures)
		r.With(middleware.HitCount(routeCounts)).Get("/{featureKey}", opt.featureAPI.GetFeature)
	})

	r.Route("/users/{userID}", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx, opt.middleware.UserCtx)

		r.With(middleware.HitCount(routeCounts)).Post("/events/{eventKey}", opt.userAPI.TrackEvent)

		r.With(middleware.HitCount(routeCounts)).Get("/features/{featureKey}", opt.userAPI.GetFeature)
		r.With(middleware.HitCount(routeCounts)).Post("/features/{featureKey}", opt.userAPI.TrackFeature)
	})

	return r
}
