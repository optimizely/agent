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

// Package handler //
package handler

import (
	"expvar"
	"net/http"
	"os"

	"github.com/go-chi/render"
)

// JSON is a map alias, just for convenience
type JSON map[string]interface{}

// Health is holding info about health checks
type Health struct {
	Status  string   `json:"status,omitempty"`
	Reasons []string `json:"reasons,omitempty"`
}

// Admin is holding info to pass to admin handlers
type Admin struct {
	Version string `json:"version,omitempty"`
	Author  string `json:"author,omitempty"`
	AppName string `json:"app_name,omitempty"`
	Host    string `json:"host,omitempty"`
}

// NewAdmin initializes admin
func NewAdmin(version, author, appName string) *Admin {
	return &Admin{Version: version, Author: author, AppName: appName}
}

// Health displays health status
func (a Admin) Health(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Health{Status: "ok"})
}

// AppInfo returns custom app-info
func (a Admin) AppInfo(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, a)
}

// AppInfoHeader adds custom app-info to the response header
func (a Admin) AppInfoHeader(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Author", a.Author)
		w.Header().Set("App-Name", a.AppName)
		w.Header().Set("App-Version", a.Version)
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
