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
	"crypto/subtle"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/optimizely/agent/config"
	"net/http"
	"time"
)

type clientCredentials struct {
	id        string
	ttl       time.Duration
	secret    []byte
}

type OAuthHandler struct {
	clientCredentials []clientCredentials
}

func NewOAuthHandler(authConfigs []*config.ServiceAuthConfig) *OAuthHandler {
	h := &OAuthHandler{
		clientCredentials: []clientCredentials{},
	}
	for _, authConfig := range authConfigs {
		for _, clientCreds := range authConfig.Clients {
			h.clientCredentials = append(h.clientCredentials, clientCredentials{
				id:        clientCreds.ID,
				secret:    []byte(clientCreds.Secret),
				ttl:       authConfig.TTL,
			})
		}
	}
	return h
}

func matchClientSecret(reqSecretStr string, configSecret []byte) bool {
	reqSecret := []byte(reqSecretStr)
	if len(configSecret) != len(reqSecret) {
		return false
	}
	return subtle.ConstantTimeCompare(reqSecret, configSecret) == 1
}

func renderTokenResponse(sdkKey string, ttl time.Duration, w http.ResponseWriter, r *http.Request) {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sdk_key": sdkKey,
		"expires": time.Now().Add(ttl).Unix(),
	})
	// TODO: get signing secret from config
	tokenString, err := token.SignedString([]byte("hmacseekrit"))
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
	}
	w.Header().Set("Content-Type", "application/jwt")
	w.WriteHeader(200)
	w.Write([]byte(tokenString))
}

// GetAPIAccessToken returns a JWT access token for an Agent service, derived from the provided client ID and client secret
func (h *OAuthHandler) GetAPIAccessToken(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	grantType := queryParams.Get("grant_type")
	if grantType == "" {
		RenderError(errors.New("grant_type query parameter required"), http.StatusBadRequest, w, r)
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
	sdkKey := queryParams.Get("sdk_key")
	if clientSecret == "" {
		RenderError(errors.New("sdk_key query parameter required"), http.StatusBadRequest, w, r)
		return
	}

	for _, clientCreds := range h.clientCredentials {
		if clientCreds.id == clientID && matchClientSecret(clientSecret, clientCreds.secret) {
			renderTokenResponse(sdkKey, clientCreds.ttl, w, r)
			return
		}
	}

	RenderError(errors.New("Invalid client_id or client_secret"), http.StatusUnauthorized, w, r)
}
