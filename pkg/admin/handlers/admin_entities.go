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
	"strings"

	"github.com/go-chi/render"
	"golang.org/x/sync/errgroup"
)

// JSON is a map alias, just for convenience
type JSON map[string]interface{}

// HealthChecker is the interface to check if something is healthy
type HealthChecker interface {
	IsHealthy() (bool, string)
}

// Health is holding info about health checks
type Health struct {
	Status string `json:"status,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// Admin is holding info to pass to admin handlers
type Admin struct {
	Version string `json:"version,omitempty"`
	Author  string `json:"author,omitempty"`
	AppName string `json:"app_name,omitempty"`
	Host    string `json:"host,omitempty"`

	checks []HealthChecker
}

// NewAdmin initializes admin
func NewAdmin(version, author, appName string, checks []HealthChecker) *Admin {
	return &Admin{Version: version, Author: author, AppName: appName, checks: checks}
}

// Health displays health status
func (a Admin) Health(w http.ResponseWriter, r *http.Request) {

	var eg errgroup.Group
	var err error
	msgList := []string{}
	if len(a.checks) > 0 {
		for _, s := range a.checks {
			s := s
			eg.Go(func() error {
				healthy, msg := s.IsHealthy()
				if !healthy {
					msgList = append(msgList, msg)
					return errors.New(msg)

				}
				return nil
			})
		}
		err = eg.Wait()
	} else {
		msg := "no services"
		msgList = append(msgList, msg)
		err = errors.New(msg)
	}
	var status, reason string
	if err == nil {
		status = "ok"
	} else {
		status = "error"
		reason = strings.Join(msgList, ", ")
	}
	render.JSON(w, r, Health{status, reason})
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
