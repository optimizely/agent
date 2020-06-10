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

// Package handlers //
package handlers

import (
	"expvar"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/render"

	"github.com/optimizely/agent/config"
)

var startTime = time.Now()

// JSON is a map alias, just for convenience
type JSON map[string]interface{}

// Info holds the detail to support the info endpoint
type Info struct {
	Version string `json:"version,omitempty"`
	Author  string `json:"author,omitempty"`
	AppName string `json:"app_name,omitempty"`
	Uptime  string `json:"uptime"`
	Host    string `json:"host,omitempty"`
}

// Admin is holding info to pass to admin handlers
type Admin struct {
	Config config.AgentConfig
	Info   Info
}

// NewAdmin initializes admin
func NewAdmin(conf config.AgentConfig) *Admin {
	info := Info{
		Version: conf.Version,
		Author:  conf.Author,
		AppName: conf.Name,
	}

	return &Admin{Config: conf, Info: info}
}

// AppInfo returns custom app-info
func (a Admin) AppInfo(w http.ResponseWriter, r *http.Request) {
	a.Info.Uptime = time.Since(startTime).String()
	render.JSON(w, r, a.Info)
}

// AppConfig returns the agent configuration
func (a Admin) AppConfig(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, a.Config)
}

// AppInfoHeader adds custom app-info to the response header
func (a Admin) AppInfoHeader(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Author", a.Info.Author)
		w.Header().Set("App-Name", a.Info.AppName)
		w.Header().Set("App-Version", a.Info.Version)
		if host := os.Getenv("HOST"); host != "" {
			w.Header().Set("Host", host)
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// Metrics returns expvar info
func (a Admin) Metrics(w http.ResponseWriter, r *http.Request) {
	expvar.Handler().ServeHTTP(w, r)
}
