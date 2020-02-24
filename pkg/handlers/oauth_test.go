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
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/middleware"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type OAuthTestSuite struct {
	suite.Suite
	handler *OAuthHandler
	mux     *chi.Mux
}

func (s *OAuthTestSuite) SetupTest() {
	config := config.ServiceAuthConfig{
		Clients: []config.OAuthClientCredentials{
			{
				ID:     "optly_user",
				Secret: "client_seekrit",
			},
		},
		HMACSecret: "hmac_seekrit",
		TTL:        30 * time.Minute,
	}
	s.handler = NewOAuthHandler(&config)

	mux := chi.NewMux()
	mux.Post("/api/token", s.handler.CreateAPIAccessToken)
	mux.Post("/admin/token", s.handler.CreateAdminAccessToken)
	s.mux = mux
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenMissingGrantType() {
	bodyBytes, _ := json.Marshal(map[string]string{
		"client_id":     "optly",
		"client_secret": "client_seekrit",
	})
	req := httptest.NewRequest("POST", "/api/token", bytes.NewReader(bodyBytes))
	req.Header.Set(middleware.OptlySDKHeader, "123")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *OAuthTestSuite) TestGetAdminAccessTokenMissingGrantType() {
	bodyBytes, _ := json.Marshal(map[string]string{
		"client_id":     "optly",
		"client_secret": "client_seekrit",
	})
	req := httptest.NewRequest("POST", "/admin/token", bytes.NewReader(bodyBytes))
	req.Header.Set(middleware.OptlySDKHeader, "123")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenUnsupportedGrantType() {
	bodyBytes, _ := json.Marshal(map[string]string{
		"grant_type":    "authorization",
		"client_id":     "optly_user",
		"client_secret": "client_seekrit",
	})
	req := httptest.NewRequest("POST", "/api/token", bytes.NewReader(bodyBytes))
	req.Header.Set(middleware.OptlySDKHeader, "123")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *OAuthTestSuite) TestGetAdminAccessTokenUnsupportedGrantType() {
	bodyBytes, _ := json.Marshal(map[string]string{
		"grant_type":    "authorization",
		"client_id":     "optly_user",
		"client_secret": "client_seekrit",
	})
	req := httptest.NewRequest("POST", "/admin/token", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenMissingClientId() {
	bodyBytes, _ := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_secret": "client_seekrit",
	})
	req := httptest.NewRequest("POST", "/api/token", bytes.NewReader(bodyBytes))
	req.Header.Set(middleware.OptlySDKHeader, "123")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *OAuthTestSuite) TestGetAdminAccessTokenMissingClientId() {
	bodyBytes, _ := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_secret": "client_seekrit",
	})
	req := httptest.NewRequest("POST", "/admin/token", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenMissingClientSecret() {
	bodyBytes, _ := json.Marshal(map[string]string{
		"grant_type": "client_credentials",
		"client_id":  "optly_user",
	})
	req := httptest.NewRequest("POST", "/api/token", bytes.NewReader(bodyBytes))
	req.Header.Set(middleware.OptlySDKHeader, "123")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *OAuthTestSuite) TestGetAdminAccessTokenMissingClientSecret() {
	bodyBytes, _ := json.Marshal(map[string]string{
		"grant_type": "client_credentials",
		"client_id":  "optly_user",
	})
	req := httptest.NewRequest("POST", "/admin/token", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenInvalidClientId() {
	bodyBytes, _ := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     "not_an_optly_user",
		"client_secret": "client_seekrit",
	})
	req := httptest.NewRequest("POST", "/api/token", bytes.NewReader(bodyBytes))
	req.Header.Set(middleware.OptlySDKHeader, "123")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusForbidden, rec.Code)
}

func (s *OAuthTestSuite) TestGetAdminAccessTokenInvalidClientId() {
	bodyBytes, _ := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     "not_an_optly_user",
		"client_secret": "client_seekrit",
	})
	req := httptest.NewRequest("POST", "/admin/token", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusForbidden, rec.Code)
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenInvalidClientSecret() {
	bodyBytes, _ := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     "optly_user",
		"client_secret": "invalid_seekret",
	})
	req := httptest.NewRequest("POST", "/api/token", bytes.NewReader(bodyBytes))
	req.Header.Set(middleware.OptlySDKHeader, "123")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusForbidden, rec.Code)
}

func (s *OAuthTestSuite) TestGetAdminAccessTokenInvalidClientSecret() {
	bodyBytes, _ := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     "optly_user",
		"client_secret": "invalid_seekret",
	})
	req := httptest.NewRequest("POST", "/admin/token", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusForbidden, rec.Code)
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenMissingSDKKey() {
	bodyBytes, _ := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     "optly_user",
		"client_secret": "client_seekrit",
	})
	req := httptest.NewRequest("POST", "/api/token", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenSuccess() {
	bodyBytes, _ := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     "optly_user",
		"client_secret": "client_seekrit",
	})
	req := httptest.NewRequest("POST", "/api/token", bytes.NewReader(bodyBytes))
	req.Header.Set(middleware.OptlySDKHeader, "123")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	var actual tokenResponse
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	s.NoError(err)
	s.Equal("bearer", actual.TokenType)
	s.NotEmpty(actual.AccessToken)
	s.NotEmpty(actual.ExpiresIn)
}

func (s *OAuthTestSuite) TestGetAdminAccessTokenSuccess() {
	bodyBytes, _ := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     "optly_user",
		"client_secret": "client_seekrit",
	})
	req := httptest.NewRequest("POST", "/admin/token", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	var actual tokenResponse
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	s.NoError(err)
	s.Equal("bearer", actual.TokenType)
	s.NotEmpty(actual.AccessToken)
	s.NotEmpty(actual.ExpiresIn)
}

func TestOAuthTestSuite(t *testing.T) {
	suite.Run(t, new(OAuthTestSuite))
}
