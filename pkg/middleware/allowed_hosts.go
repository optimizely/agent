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

// Package middleware //
package middleware

import (
	"errors"
	"fmt"
	"net/http"
)

// AllowedHosts returns a middleware function that rejects requests whose host value does not match any host in allowedHosts.
// Request host is determined in the following priority order:
// 1. X-Forwarded-Host header value
// 2. Forwarded header host= directive value
// 3. Host property of request (see Host under https://golang.org/pkg/net/http/#Request)
func AllowedHosts(allowedHosts []string, allowedPort string) func(next http.Handler) http.Handler {
	allowedMap := make(map[string]bool)
	for _, allowedHost := range allowedHosts {
		allowedMap[fmt.Sprintf("%v:%v", allowedHost, allowedPort)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := requestHost(r)
			if allowedMap[host] {
				next.ServeHTTP(w, r)
				return
			}
			logger := GetLogger(r)
			logger.Debug().Strs("allowedHosts", allowedHosts).Str("allowedPort", allowedPort).Str("Host", r.Host).Str("X-Forwarded-Host", r.Header.Get("X-Forwarded-Host")).Str("Forwarded", r.Header.Get("Forwarded")).Msg("Request failed allowed hosts check")
			RenderError(errors.New("invalid request host"), http.StatusNotFound, w, r)
		})
	}
}
