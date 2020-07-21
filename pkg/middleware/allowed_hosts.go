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
	"strings"
)

var errInvalidRequestHost = errors.New("invalid request host")

// AllowedHosts returns a middleware function that rejects requests whose host value does not match any host in allowedHosts.
// Request host is determined in the following priority order:
// 1. X-Forwarded-Host header value
// 2. Forwarded header host= directive value
// 3. Host property of request (see Host under https://golang.org/pkg/net/http/#Request)
func AllowedHosts(allowedHosts []string, allowedPort string, usingTLS bool) func(next http.Handler) http.Handler {
	allowedMap := make(map[string]bool)

	// We want to allow requests with hosts that don't contain explicit port when the server is running on the  default
	// port.
	shouldAllowPortOmitted := false
	if usingTLS {
		shouldAllowPortOmitted = allowedPort == "443"
	} else {
		shouldAllowPortOmitted = allowedPort == "80"
	}

	for _, allowedHost := range allowedHosts {
		allowedMap[fmt.Sprintf("%v:%v", allowedHost, allowedPort)] = true
		// When appropriate, create entry in allowedMap without explicit port
		if shouldAllowPortOmitted {
			allowedMap[allowedHost] = true
		}
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
			RenderError(errInvalidRequestHost, http.StatusNotFound, w, r)
		})
	}
}

// requestHost and parseForwarded are taken from https://github.com/go-chi/hostrouter
// (permanent link: https://github.com/go-chi/hostrouter/blob/7bff2694dfd99a31a89c62e5f8a2d9ec2d71da8e/hostrouter.go)
/*
Copyright (c) 2016-Present https://github.com/go-chi authors

MIT License

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/
func requestHost(r *http.Request) (host string) {
	// not standard, but most popular
	host = r.Header.Get("X-Forwarded-Host")
	if host != "" {
		return
	}

	// RFC 7239
	host = r.Header.Get("Forwarded")
	_, _, host = parseForwarded(host)
	if host != "" {
		return
	}

	// if all else fails fall back to request host
	host = r.Host
	return
}

func parseForwarded(forwarded string) (addr, proto, host string) {
	if forwarded == "" {
		return
	}
	for _, forwardedPair := range strings.Split(forwarded, ";") {
		if tv := strings.SplitN(forwardedPair, "=", 2); len(tv) == 2 {
			token, value := tv[0], tv[1]
			token = strings.TrimSpace(token)
			value = strings.TrimSpace(strings.Trim(value, `"`))
			switch strings.ToLower(token) {
			case "for":
				addr = value
			case "proto":
				proto = value
			case "host":
				host = value
			}

		}
	}
	return
}
