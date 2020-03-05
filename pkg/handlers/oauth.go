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
	"net/http"
	"time"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/jwtauth"
	"github.com/optimizely/agent/pkg/middleware"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

// ClientCredentials has all info for client credentials
type ClientCredentials struct {
	ID         string
	TTL        time.Duration
	SecretHash []byte
}

// OAuthHandler provides handler for auth
type OAuthHandler struct {
	ClientCredentials map[string]ClientCredentials
	hmacSecret        []byte
}

type tokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func renderAccessTokenResponse(w http.ResponseWriter, r *http.Request, accessToken string, ttl time.Duration) {
	render.JSON(w, r, tokenResponse{
		accessToken,
		"bearer",
		int64(ttl.Seconds()),
	})
}

// NewOAuthHandler creates new handler for auth
func NewOAuthHandler(authConfig *config.ServiceAuthConfig) *OAuthHandler {

	clientCredentials := make(map[string]ClientCredentials)
	// TODO: need to validate all client IDs are unique
	for _, clientCreds := range authConfig.Clients {
		secretHashBytes, err := jwtauth.DecodeSecretHashFromConfig(clientCreds.SecretHash)
		if err != nil {
			log.Error().Err(err).Msgf("error decoding client creds secret (paired with client ID: %v)", clientCreds.ID)
			continue
		}
		clientCredentials[clientCreds.ID] = ClientCredentials{
			ID:         clientCreds.ID,
			SecretHash: secretHashBytes,
			TTL:        authConfig.TTL,
		}
	}

	hmacSecret := ""
	if len(authConfig.HMACSecrets) > 0 {
		hmacSecret = authConfig.HMACSecrets[0]
	}

	h := &OAuthHandler{
		hmacSecret:        []byte(hmacSecret),
		ClientCredentials: clientCredentials,
	}
	return h
}

// ClientCredentialsError is the response body returned when the provided client credentials are invalid
type ClientCredentialsError struct {
	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

func (err *ClientCredentialsError) Error() string {
	return err.ErrorCode
}

func (h *OAuthHandler) verifyClientCredentials(r *http.Request) (*ClientCredentials, int, error) {
	var reqBody tokenRequest
	err := ParseRequestBody(r, &reqBody)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	if reqBody.GrantType == "" {
		return nil, http.StatusBadRequest, &ClientCredentialsError{
			ErrorCode:        "invalid_request",
			ErrorDescription: "grant_type missing from request body",
		}
	}
	if reqBody.GrantType != "client_credentials" {
		return nil, http.StatusBadRequest, &ClientCredentialsError{
			ErrorCode: "unsupported_grant_type",
		}
	}

	if reqBody.ClientID == "" {
		return nil, http.StatusUnauthorized, &ClientCredentialsError{
			ErrorCode:        "invalid_client",
			ErrorDescription: "client_id missing from request body",
		}
	}

	if reqBody.ClientSecret == "" {
		return nil, http.StatusUnauthorized, &ClientCredentialsError{
			ErrorCode:        "invalid_client",
			ErrorDescription: "client_secret missing from request body",
		}
	}
	clientCreds, ok := h.ClientCredentials[reqBody.ClientID]
	if !ok {
		return nil, http.StatusUnauthorized, &ClientCredentialsError{
			ErrorCode:        "invalid_client",
			ErrorDescription: "invalid client_id or client_secret",
		}
	}

	isValid, err := jwtauth.ValidateClientSecret(reqBody.ClientSecret, clientCreds.SecretHash)
	if err != nil {
		middleware.GetLogger(r).Info().Err(err).Msg("validating request secret")
	}
	if !isValid {
		return nil, http.StatusUnauthorized, &ClientCredentialsError{
			ErrorCode:        "invalid_client",
			ErrorDescription: "invalid client_id or client_secret",
		}
	}

	return &clientCreds, http.StatusOK, nil
}

func renderClientCredentialsError(err error, status int, w http.ResponseWriter, r *http.Request) {
	middleware.GetLogger(r).Debug().Err(err).Int("status", status).Msg("render client credentials error")
	render.Status(r, status)
	render.JSON(w, r, err)
}

// CreateAPIAccessToken returns a JWT access token for the API service
func (h *OAuthHandler) CreateAPIAccessToken(w http.ResponseWriter, r *http.Request) {

	clientCreds, httpCode, err := h.verifyClientCredentials(r)
	if err != nil {
		renderClientCredentialsError(err, httpCode, w, r)
		return
	}

	if len(h.hmacSecret) == 0 {
		middleware.GetLogger(r).Error().Msg("Invalid hmac secret in configuration, can't issue token")
		render.Status(r, http.StatusInternalServerError)
		render.PlainText(w, r, "Invalid server configuration, can't issue token")
		return
	}

	sdkKey := r.Header.Get(middleware.OptlySDKHeader)
	if sdkKey == "" {
		renderClientCredentialsError(&ClientCredentialsError{
			ErrorCode:        "invalid_request",
			ErrorDescription: "X-Optimizely-Sdk-Key header required",
		}, http.StatusBadRequest, w, r)
		return
	}

	accessToken, err := jwtauth.BuildAPIAccessToken(sdkKey, clientCreds.TTL, h.hmacSecret)
	if err != nil {
		middleware.GetLogger(r).Error().Err(err).Msg("Calling jwt BuildAPIAccessToken")
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	renderAccessTokenResponse(w, r, accessToken, clientCreds.TTL)
}

// CreateAdminAccessToken returns a JWT access token for the Admin service
func (h *OAuthHandler) CreateAdminAccessToken(w http.ResponseWriter, r *http.Request) {

	clientCreds, httpCode, err := h.verifyClientCredentials(r)
	if err != nil {
		renderClientCredentialsError(err, httpCode, w, r)
		return
	}

	if len(h.hmacSecret) == 0 {
		middleware.GetLogger(r).Error().Msg("Invalid hmac secret in configuration, can't issue token")
		render.Status(r, http.StatusInternalServerError)
		render.PlainText(w, r, "Invalid server configuration, can't issue token")
		return
	}

	accessToken, err := jwtauth.BuildAdminAccessToken(clientCreds.TTL, h.hmacSecret)
	if err != nil {
		middleware.GetLogger(r).Error().Err(err).Msg("Calling jwt BuildAdminAccessToken")
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	renderAccessTokenResponse(w, r, accessToken, clientCreds.TTL)
}
