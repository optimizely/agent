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

// Package middleware
package middleware

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"github.com/optimizely/agent/config"
	"net/http"
	"time"
)

const oAuthClientCreds = contextKey("oAuthClientCreds")

type ClientCredentials struct {
	ID     string
	TTL    time.Duration
	Secret []byte
}

func matchClientSecret(reqSecretStr string, configSecret []byte) bool {
	reqSecret := []byte(reqSecretStr)
	if len(configSecret) != len(reqSecret) {
		return false
	}
	return subtle.ConstantTimeCompare(reqSecret, configSecret) == 1
}

var ErrInvalidHMACSecret = errors.New("HMACSecret unavailable")

var ErrInvalidClients = errors.New("Clients unavailable")

func OAuthMiddleware(authConfig *config.ServiceAuthConfig) (func(http.Handler) http.Handler, error) {
	if authConfig.HMACSecret == "" {
		return nil, ErrInvalidHMACSecret
	}

	if  len(authConfig.Clients) == 0 {
		return nil, ErrInvalidClients
	}

	clientCredentials := make(map[string]ClientCredentials)
	for _, clientCreds := range authConfig.Clients {
		clientCredentials[clientCreds.ID] = ClientCredentials{
			ID:     clientCreds.ID,
			Secret: []byte(clientCreds.Secret),
			TTL:    authConfig.TTL,
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			queryParams := r.URL.Query()
			grantType := queryParams.Get("grant_type")
			if grantType == "" {
				RenderError(errors.New("grant_type query parameter required"), http.StatusBadRequest, w, r)
				return
			}
			if grantType != "client_credentials" {
				RenderError(fmt.Errorf("unsupported grant_type %v", grantType), http.StatusBadRequest, w, r)
				return

			}
			clientID := queryParams.Get("client_id")
			if clientID == "" {
				RenderError(errors.New("client_id query parameter required"), http.StatusBadRequest, w, r)
				return
			}
			clientSecret := queryParams.Get("client_secret")
			if clientSecret == "" {
				RenderError(errors.New("client_secret query parameter required"), http.StatusBadRequest, w, r)
				return
			}
			clientCreds, ok := clientCredentials[clientID]
			if ok && matchClientSecret(clientSecret, clientCreds.Secret) {
				ctx := context.WithValue(r.Context(), oAuthClientCreds, &clientCreds)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				RenderError(errors.New("Invalid client_id or client_secret"), http.StatusUnauthorized, w, r)
			}
		})
	}, nil
}
