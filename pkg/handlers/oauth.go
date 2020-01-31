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

// Package handlers //
package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/jwtauth"
	"github.com/optimizely/agent/pkg/middleware"

	"github.com/go-chi/render"
)

// ClientCredentials has all info for client credentials
type ClientCredentials struct {
	ID     string
	TTL    time.Duration
	Secret []byte
}

// OAuthHandler provides handler for auth
type OAuthHandler struct {
	ClientCredentials map[string]ClientCredentials
	hmacSecret        []byte
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func renderAccessTokenResponse(w http.ResponseWriter, r *http.Request, accessToken string, expires int64) {
	// TODO: expires_in should be in seconds, per https://tools.ietf.org/html/rfc6749#section-5.1
	render.JSON(w, r, tokenResponse{accessToken, "bearer", expires})
}

// NewOAuthHandler creates new handler for auth
func NewOAuthHandler(authConfig *config.ServiceAuthConfig) *OAuthHandler {

	clientCredentials := make(map[string]ClientCredentials)
	for _, clientCreds := range authConfig.Clients {
		clientCredentials[clientCreds.ID] = ClientCredentials{
			ID:     clientCreds.ID,
			Secret: []byte(clientCreds.Secret),
			TTL:    authConfig.TTL,
		}
	}

	h := &OAuthHandler{
		hmacSecret:        []byte(authConfig.HMACSecret),
		ClientCredentials: clientCredentials,
	}
	return h
}

func (h *OAuthHandler) verifyClientCredentials(r *http.Request) (*ClientCredentials, int, error) {
	queryParams := r.URL.Query()
	grantType := queryParams.Get("grant_type")
	if grantType == "" {
		return nil, http.StatusBadRequest, errors.New("grant_type query parameter required")
	}
	if grantType != "client_credentials" {
		return nil, http.StatusBadRequest, fmt.Errorf("unsupported grant_type %v", grantType)

	}
	clientID := queryParams.Get("client_id")
	if clientID == "" {
		return nil, http.StatusUnauthorized, errors.New("client_id query parameter required")
	}
	clientSecret := queryParams.Get("client_secret")
	if clientSecret == "" {
		return nil, http.StatusUnauthorized, errors.New("client_secret query parameter required")
	}
	clientCreds, ok := h.ClientCredentials[clientID]
	if !ok || !jwtauth.MatchClientSecret(clientSecret, clientCreds.Secret) {
		return nil, http.StatusForbidden, errors.New("invalid client_id or client_secret")
	}
	return &clientCreds, http.StatusOK, nil
}

// GetAPIAccessToken returns a JWT access token for the API service
func (h *OAuthHandler) GetAPIAccessToken(w http.ResponseWriter, r *http.Request) {

	clientCreds, httpCode, e := h.verifyClientCredentials(r)
	if e != nil {
		// TODO: set correct error property in response body as described here: https://tools.ietf.org/html/rfc6749#section-5.2
		RenderError(e, httpCode, w, r)
		return
	}

	sdkKey := r.Header.Get(middleware.OptlySDKHeader)
	if sdkKey == "" {
		RenderError(errors.New("sdk_key required in the header"), http.StatusBadRequest, w, r)
		return
	}

	accessToken, expires, err := jwtauth.BuildAPIAccessToken(sdkKey, clientCreds.TTL, h.hmacSecret)
	if err != nil {
		middleware.GetLogger(r).Error().Err(err).Msg("Calling jwt BuildAPIAccessToken")
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	renderAccessTokenResponse(w, r, accessToken, expires)
}

// GetAdminAccessToken returns a JWT access token for the Admin service
func (h *OAuthHandler) GetAdminAccessToken(w http.ResponseWriter, r *http.Request) {

	clientCreds, httpCode, e := h.verifyClientCredentials(r)
	if e != nil {
		RenderError(e, httpCode, w, r)
		return
	}

	accessToken, expires, err := jwtauth.BuildAdminAccessToken(clientCreds.TTL, h.hmacSecret)
	if err != nil {
		middleware.GetLogger(r).Error().Err(err).Msg("Calling jwt BuildAdminAccessToken")
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	renderAccessTokenResponse(w, r, accessToken, expires)
}
