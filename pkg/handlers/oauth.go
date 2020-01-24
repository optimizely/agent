/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
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
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/render"
	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/middleware"
	"net/http"
	"time"
)

type OAuthHandler struct {
	hmacSecret        []byte
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	Expires     int64  `json:"expires"`
}

func NewOAuthHandler(authConfig *config.ServiceAuthConfig) *OAuthHandler {
	h := &OAuthHandler{
		hmacSecret:     []byte(authConfig.HMACSecret),
	}
	return h
}

func renderTokenResponse(sdkKey string, ttl time.Duration, hmacSecret []byte, w http.ResponseWriter, r *http.Request) {
	expires := time.Now().Add(ttl).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sdk_key": sdkKey,
		"expires": expires,
	})
	tokenString, err := token.SignedString(hmacSecret)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
	}
	render.JSON(w, r, tokenResponse{tokenString, expires})
}

// GetAccessToken returns a JWT access token for an Agent service, derived from the provided client ID and client secret
func (h *OAuthHandler) GetAccessToken(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	sdkKey := queryParams.Get("sdk_key")
	if sdkKey == "" {
		RenderError(errors.New("sdk_key query parameter required"), http.StatusBadRequest, w, r)
		return
	}

	clientCreds, err := middleware.GetClientCreds(r)
	if err != nil {
		middleware.GetLogger(r).Error().Err(err).Msg("Calling middleware GetClientCreds")
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	renderTokenResponse(sdkKey, clientCreds.TTL, h.hmacSecret, w, r)
}

