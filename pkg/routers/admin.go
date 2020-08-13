/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
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
	"html/template"
	"net/http"
	"net/http/pprof"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/handlers"
	"github.com/optimizely/agent/pkg/middleware"

	"github.com/go-chi/chi"
	chimw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

// NewAdminRouter returns HTTP admin router
func NewAdminRouter(conf config.AgentConfig) http.Handler {
	r := chi.NewRouter()

	authProvider := middleware.NewAuth(&conf.Admin.Auth)

	if authProvider == nil {
		log.Error().Msg("unable to initialize admin auth middleware.")
		return nil
	}

	tokenHandler := handlers.NewOAuthHandler(&conf.Admin.Auth)
	if tokenHandler == nil {
		log.Error().Msg("unable to initialize admin auth handler.")
		return nil
	}

	optlyAdmin := handlers.NewAdmin(conf)
	r.Use(optlyAdmin.AppInfoHeader)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	routes := []Route{
		{"/config", optlyAdmin.AppConfig, "Get application configuration.", true},
		{"/info", optlyAdmin.AppInfo, "Get application info.", true},
		{"/metrics", optlyAdmin.Metrics, "Get application metrics (e.g. expvar).", true},
		{"/cluster", handlers.GetClusterInfo, "Get current cluster state.", true},
		{"/debug/pprof/*", pprof.Index, "Get current cluster state.", false},
		{"/debug/pprof/cmdline", pprof.Cmdline, "Get current cluster state.", false},
		{"/debug/pprof/profile", pprof.Profile, "Get current cluster state.", false},
		{"/debug/pprof/symbol", pprof.Symbol, "Get current cluster state.", false},
		{"/debug/pprof/trace", pprof.Trace, "Get current cluster state.", false},
	}

	index := Index(routes)
	routes = append(routes, Route{"/", index, "index", false})

	for _, route := range routes {
		r.With(authProvider.AuthorizeAdmin).Get(route.Pattern, route.Handler)
	}

	r.With(chimw.AllowContentType("application/json")).Post("/oauth/token", tokenHandler.CreateAdminAccessToken)
	return r
}

type Route struct {
	Pattern string
	Handler http.HandlerFunc
	Desc    string
	Index   bool
}

func Index(routes []Route) func(w http.ResponseWriter, r *http.Request) {
	index := make([]Route, 0, len(routes))
	for _, route := range routes {
		if route.Index {
			index = append(index, route)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := indexTmpl.Execute(w, index); err != nil {
			log.Print(err)
		}
	}
}

var indexTmpl = template.Must(template.New("index").Parse(`<html>
<head>
<title>Agent Admin</title>
</head>
<body>
<br>
<p>
Admin Routes:
<ul>
{{range .}}
<li><a href={{.Pattern}}>{{.Pattern}}</a>: {{.Desc}}</li>
{{end}}
</ul>
</p>
</body>
</html>
`))
