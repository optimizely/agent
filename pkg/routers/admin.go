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
	"encoding/json"
	"expvar"
	"fmt"
	"github.com/optimizely/agent/pkg/optimizely"
	"go.opencensus.io/stats/view"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/handlers"
	"github.com/optimizely/agent/pkg/middleware"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/go-chi/chi"
	chimw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

type expvarCollector struct {
	exports map[string]*prom.Desc
}

func NewExpvarCollector(exports map[string]*prom.Desc) prom.Collector {
	return &expvarCollector{
		exports: exports,
	}
}

// Describe implements Collector.
func (e *expvarCollector) Describe(ch chan<- *prom.Desc) {
	for _, desc := range e.exports {
		ch <- desc
	}
}

func (e *expvarCollector) Collect(ch chan<- prom.Metric) {
	for name, desc := range e.exports {
		var m prom.Metric
		var v interface{}
		//if strings.Contains(name, "responseTime") {
		//
		//}
		//if strings.Contains(name, "hits") {
		//	expVar := expvar.Get(name)
		//	if expVar == nil {
		//		continue
		//	}
		//	if err := json.Unmarshal([]byte(expVar.String()), &v); err != nil {
		//		ch <- prom.NewInvalidMetric(desc, err)
		//		continue
		//	}
		//	switch v := v.(type) {
		//	case float64:
		//		m, _ = prom.NewConstMetric(desc, prom.UntypedValue, v)
		//	default:
		//		continue
		//	}
		//	ch <- m
		//}
		if strings.Contains(name, "responseTimeHist") {
			for _, val := range []string{"p50", "p90", "p95", "p99"} {
				histName := fmt.Sprintf("%s.%s", name, val)
				expVar := expvar.Get(histName)
				if expVar == nil {
					continue
				}
				if err := json.Unmarshal([]byte(expVar.String()), &v); err != nil {
					ch <- prom.NewInvalidMetric(desc, err)
					continue
				}
				switch v := v.(type) {
				case float64:
					m, _ = prom.NewConstMetric(desc, prom.UntypedValue, v, val)
				default:
					continue
				}
				ch <- m
			}
			continue
		}
		expVar := expvar.Get(name)
		if expVar == nil {
			continue
		}
		if err := json.Unmarshal([]byte(expVar.String()), &v); err != nil {
			ch <- prom.NewInvalidMetric(desc, err)
			continue
		}
		switch v := v.(type) {
		case float64:
			m, _ = prom.NewConstMetric(desc, prom.UntypedValue, v)
		default:
			continue
		}
		ch <- m
	}
}

// NewAdminRouter returns HTTP admin router
func NewAdminRouter(conf config.AgentConfig, metricsRegistry *optimizely.MetricsRegistry) http.Handler {
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

	expvarCollector := NewExpvarCollector(map[string]*prom.Desc{
		"memstats": prom.NewDesc(
			"expvar_memstats",
			"All numeric memstats as one metric family. Not a good role-model, actually... ;-)",
			[]string{"type"}, nil,
		),
		"timer.activate.responseTimeHist": prom.NewDesc(
			"aa_timer_activate",
			"Just an expvar int as an example.",
			[]string{"le"}, nil,
		),
		"timer.activate.responseTime": prom.NewDesc(
			"aa_timer_activate_resp",
			"Just an expvar int as an example.",
			nil, nil,
		),
		"timer.activate.hits": prom.NewDesc(
			"aa_timer_activate_hits",
			"Just an expvar int as an example.",
			nil, nil,
		),
		//"memstats.PauseNs": prom.NewDesc(
		//	"aa_timer_get_datafile",
		//	"How many http requests processed, partitioned by status code and http method.",
		//	nil, nil,
		//),
	})

	foo := expvar.Get("timer.save.responseTimeHist.h")

	fmt.Printf("%q", foo)

	registry := prom.DefaultRegisterer.(*prom.Registry)
	registry.MustRegister(expvarCollector)
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: strings.Replace("deploymentName", "-", "_", -1),
		Registry:  registry,
	})
	if err != nil {
		fmt.Print("failed to create prometheus registry")
	}
	view.RegisterExporter(pe)

	optlyAdmin := handlers.NewAdmin(conf)
	r.Use(optlyAdmin.AppInfoHeader)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.With(authProvider.AuthorizeAdmin).Get("/config", optlyAdmin.AppConfig)
	r.With(authProvider.AuthorizeAdmin).Get("/info", optlyAdmin.AppInfo)
	r.Handle("/metrics", pe)

	r.With(authProvider.AuthorizeAdmin).Get("/debug/pprof/*", pprof.Index)
	r.With(authProvider.AuthorizeAdmin).Get("/debug/pprof/cmdline", pprof.Cmdline)
	r.With(authProvider.AuthorizeAdmin).Get("/debug/pprof/profile", pprof.Profile)
	r.With(authProvider.AuthorizeAdmin).Get("/debug/pprof/symbol", pprof.Symbol)
	r.With(authProvider.AuthorizeAdmin).Get("/debug/pprof/trace", pprof.Trace)

	r.With(chimw.AllowContentType("application/json")).Post("/oauth/token", tokenHandler.CreateAdminAccessToken)
	return r
}
