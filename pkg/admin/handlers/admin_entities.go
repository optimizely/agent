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

// Package handlers //
package handlers

import (
	"errors"
	"net/http"
	"os"

	"github.com/go-chi/render"
	"golang.org/x/sync/errgroup"
)

// JSON is a map alias, just for convenience
type JSON map[string]interface{}

// AliveChecker is the interface to check if something is alive
type AliveChecker interface {
	IsAlive() bool
}

// Admin is holding info to pass to admin handlers
type Admin struct {
	Version string `json:"version,omitempty"`
	Author  string `json:"author,omitempty"`
	AppName string `json:"app_name,omitempty"`
	Host    string `json:"host,omitempty"`

	srvcs []AliveChecker
}

// NewAdmin initializes admin
func NewAdmin(version, author, appName string, srvcs []AliveChecker) *Admin {
	return &Admin{Version: version, Author: author, AppName: appName, srvcs: srvcs}
}

// Health displays health status
func (a Admin) Health(w http.ResponseWriter, r *http.Request) {

	var eg errgroup.Group
	var err error
	if len(a.srvcs) > 0 {
		for _, s := range a.srvcs {
			s := s
			eg.Go(func() error {
				if !s.IsAlive() {
					return errors.New("failed")
				}
				return nil
			})
		}
		err = eg.Wait()
	} else {
		err = errors.New("no services")
	}
	var status string
	if err == nil {
		status = "ok"
	} else {
		status = "error"
	}
	render.JSON(w, r, JSON{"status": status})
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
