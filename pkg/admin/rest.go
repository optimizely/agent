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

// Package admin //
package admin

import (
	"net/http"
	"os"

	"github.com/go-chi/render"
)

// JSON is a map alias, just for convenience
type JSON map[string]interface{}

// Admin is holding info to pass to admin handlers
type Admin struct {
	version string
	author  string
	appName string
}

// NewAdmin initializes admin
func NewAdmin(version, author, appName string) *Admin {
	return &Admin{version: version, author: author, appName: appName}
}

// Health displays health status
func (a Admin) Health(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, JSON{"status": "ok"})
}

// AppInfo returns custom app-info
func (a Admin) AppInfo(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, JSON{"author": a.author, "app_name": a.appName, "app_version": a.version, "host": os.Getenv("HOST")})
}

// AppInfoHeader adds custom app-info to the response header
func (a Admin) AppInfoHeader(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Author", a.author)
		w.Header().Set("App-Name", a.appName)
		w.Header().Set("App-Version", a.version)
		if host := os.Getenv("HOST"); host != "" {
			w.Header().Set("Host", host)
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
