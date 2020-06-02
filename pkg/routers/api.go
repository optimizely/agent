/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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

	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog/log"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/handlers"
	"github.com/optimizely/agent/pkg/metrics"
	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	_ "github.com/optimizely/agent/statik" // Required to serve openapi.yaml

	"github.com/go-chi/chi"
	chimw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
)

// APIOptions defines the configuration parameters for Router.
type APIOptions struct {
	maxConns        int
	sdkMiddleware   func(next http.Handler) http.Handler
	metricsRegistry *metrics.Registry
	configHandler   http.HandlerFunc
	activateHandler http.HandlerFunc
	healthHandler   http.HandlerFunc
	trackHandler    http.HandlerFunc
	overrideHandler http.HandlerFunc
	nStreamHandler  http.HandlerFunc
	oAuthHandler    http.HandlerFunc
	oAuthMiddleware func(next http.Handler) http.Handler
	corsHandler     func(next http.Handler) http.Handler
}

func forbiddenHandler(message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, message, http.StatusForbidden)
	}
}

// NewDefaultAPIRouter creates a new router with the default backing optimizely.Cache
func NewDefaultAPIRouter(optlyCache optimizely.Cache, conf config.APIConfig, metricsRegistry *metrics.Registry) http.Handler {

	authProvider := middleware.NewAuth(&conf.Auth)
	if authProvider == nil {
		log.Error().Msg("unable to initialize api auth middleware.")
		return nil
	}

	authHandler := handlers.NewOAuthHandler(&conf.Auth)
	if authHandler == nil {
		log.Error().Msg("unable to initialize api auth handler.")
		return nil
	}

	overrideHandler := handlers.Override
	if !conf.EnableOverrides {
		overrideHandler = forbiddenHandler("Overrides not enabled")
	}

	nStreamHandler := handlers.NotificationEventSteamHandler
	if !conf.EnableNotifications {
		nStreamHandler = forbiddenHandler("Notification stream not enabled")
	}

	mw := middleware.CachedOptlyMiddleware{Cache: optlyCache}
	corsHandler := createCorsHandler(conf.CORS)

	spec := &APIOptions{
		maxConns:        conf.MaxConns,
		metricsRegistry: metricsRegistry,
		configHandler:   handlers.OptimizelyConfig,
		activateHandler: handlers.Activate,
		healthHandler:   handlers.Health,
		overrideHandler: overrideHandler,
		trackHandler:    handlers.TrackEvent,
		sdkMiddleware:   mw.ClientCtx,
		nStreamHandler:  nStreamHandler,
		oAuthHandler:    authHandler.CreateAPIAccessToken,
		oAuthMiddleware: authProvider.AuthorizeAPI,
		corsHandler:     corsHandler,
	}

	return NewAPIRouter(spec)
}

// NewAPIRouter returns HTTP API router backed by an optimizely.Cache implementation
func NewAPIRouter(opt *APIOptions) *chi.Mux {
	r := chi.NewRouter()
	WithAPIRouter(opt, r)
	return r
}

// WithAPIRouter appends routes and middleware to the given router.
// See https://godoc.org/github.com/go-chi/chi#Mux.Group for usage
func WithAPIRouter(opt *APIOptions, r chi.Router) {
	getConfigTimer := middleware.Metricize("get-config", opt.metricsRegistry)
	activateTimer := middleware.Metricize("activate", opt.metricsRegistry)
	overrideTimer := middleware.Metricize("override", opt.metricsRegistry)
	trackTimer := middleware.Metricize("track-event", opt.metricsRegistry)
	createAccesstokenTimer := middleware.Metricize("create-api-access-token", opt.metricsRegistry)
	healthTimer := middleware.Metricize("health", opt.metricsRegistry)

	if opt.maxConns > 0 {
		// Note this is NOT a rate limiter, but a concurrency threshold
		r.Use(chimw.Throttle(opt.maxConns))
	}

	r.Use(middleware.SetTime)
	r.Use(render.SetContentType(render.ContentTypeJSON), middleware.SetRequestID)

	r.Route("/v1", func(r chi.Router) {
		r.Use(opt.corsHandler, opt.sdkMiddleware)
		r.With(getConfigTimer, opt.oAuthMiddleware).Get("/config", opt.configHandler)
		r.With(activateTimer, opt.oAuthMiddleware).Post("/activate", opt.activateHandler)
		r.With(trackTimer, opt.oAuthMiddleware).Post("/track", opt.trackHandler)
		r.With(overrideTimer, opt.oAuthMiddleware).Post("/override", opt.overrideHandler)
		r.With(opt.oAuthMiddleware).Get("/notifications/event-stream", opt.nStreamHandler)
	})

	r.With(healthTimer).Get("/health", opt.healthHandler)
	r.With(createAccesstokenTimer).Post("/oauth/token", opt.oAuthHandler)

	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	r.Handle("/*", staticServer)
}

func createCorsHandler(c config.CORSConfig) func(next http.Handler) http.Handler {
	options := cors.Options{
		AllowedOrigins:   c.AllowedOrigins,
		AllowedMethods:   c.AllowedMethods,
		AllowedHeaders:   c.AllowedHeaders,
		ExposedHeaders:   c.ExposedHeaders,
		AllowCredentials: c.AllowedCredentials,
		MaxAge:           c.MaxAge,
	}

	return cors.Handler(options)
}
