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
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/hostrouter"
	"github.com/rs/zerolog/log"
)

func createAllowedHostsRouter(r chi.Router, allowedHosts []string, allowedPort string) http.Handler {
	hr := hostrouter.New()
	for _, allowedHost := range allowedHosts {
		hr.Map(fmt.Sprintf("%v:%v", allowedHost, allowedPort), r)
	}

	hostCheckFailedRouter := chi.NewRouter()
	hostCheckFailedRouter.Mount("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Strs("allowedHosts", allowedHosts).Str("allowedPort", allowedPort).Str("Host", r.Host).Str("X-Forwarded-Host", r.Header.Get("X-Forwarded-Host")).Str("Forwarded", r.Header.Get("Forwarded")).Msg("Request failed allowed hosts check")
		http.Error(w, "invalid request host", http.StatusNotFound)
	}))
	hr.Map("*", hostCheckFailedRouter)
	hr.Map("", hostCheckFailedRouter)

	log.Debug().Msgf("%v", hr.Routes())

	return hr
}
