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

	"github.com/go-chi/chi"
	chimw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"

	"github.com/optimizely/sidedoor/pkg/api/handlers"
	"github.com/optimizely/sidedoor/pkg/api/middleware"
	"github.com/optimizely/sidedoor/pkg/optimizely"
)

// Router defines the configuration parameters for Router.
type RouterOptions struct {
	middleware   middleware.OptlyMiddleware
	featureAPI   handlers.FeatureAPI
	userEventAPI handlers.UserEventAPI
}

// NewDefaultRouter creates a new router with the default backing optimizely.Cache
func NewDefaultRouter() *chi.Mux {
	spec := &RouterOptions{
		middleware:   &middleware.OptlyContext{optimizely.NewCache()},
		featureAPI:   new(handlers.FeatureHandler),
		userEventAPI: new(handlers.UserEventHandler),
	}

	return NewRouter(spec)
}

// NewRouter returns HTTP API router backed by an optimizely.Cache implementation
func NewRouter(opt *RouterOptions) *chi.Mux {
	r := chi.NewRouter()
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("pong")); err != nil {
			log.Fatal().Msg("unable to write response")
		}
	})

	r.With(chimw.AllowContentType("application/json")).Post("/user-event", opt.userEventAPI.AddUserEvent)

	r.Route("/features", func(r chi.Router) {
		r.Use(opt.middleware.ClientCtx)
		r.Get("/", opt.featureAPI.ListFeatures)

		r.Route("/{featureKey}", func(r chi.Router) {
			// TODO r.Use(FeatureCtx)
			r.Get("/", opt.featureAPI.GetFeature)
			r.Post("/activate", opt.featureAPI.ActivateFeature)
		})
	})

	return r
}
