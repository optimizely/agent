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
	mux.Get("/api/token", s.handler.GetAPIAccessToken)
	mux.Get("/admin/token", s.handler.GetAdminAccessToken)
	s.mux = mux
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenMissingGrantType() {
	req := httptest.NewRequest("GET", "/api/token?client_id=optly_user&client_secret=client_seekrit", nil)
	req.Header.Set(middleware.OptlySDKHeader, "123")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusBadRequest, rec.Code)
}
func (s *OAuthTestSuite) TestGetAdminAccessTokenMissingGrantType() {
	req := httptest.NewRequest("GET", "/admin/token?client_id=optly_user&client_secret=client_seekrit", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenUnsupportedGrantType() {
	req := httptest.NewRequest("GET", "/api/token?grant_type=authorizationclient_id=optly_user&client_secret=client_seekrit", nil)
	req.Header.Set(middleware.OptlySDKHeader, "123")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *OAuthTestSuite) TestGetAdminAccessTokenUnsupportedGrantType() {
	req := httptest.NewRequest("GET", "/admin/token?grant_type=authorizationclient_id=optly_user&client_secret=client_seekrit", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenMissingClientId() {
	req := httptest.NewRequest("GET", "/api/token?grant_type=client_credentials&client_secret=client_seekrit", nil)
	req.Header.Set(middleware.OptlySDKHeader, "123")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *OAuthTestSuite) TestGetAdminAccessTokenMissingClientId() {
	req := httptest.NewRequest("GET", "/admin/token?grant_type=client_credentials&client_secret=client_seekrit", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenMissingClientSecret() {
	req := httptest.NewRequest("GET", "/api/token?grant_type=client_credentials&client_id=optly_user", nil)
	req.Header.Set(middleware.OptlySDKHeader, "123")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *OAuthTestSuite) TestGetAdminAccessTokenMissingClientSecret() {
	req := httptest.NewRequest("GET", "/admin/token?grant_type=client_credentials&client_id=optly_user", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenInvalidClientId() {
	req := httptest.NewRequest("GET", "/api/token?grant_type=client_credentials&client_id=not_an_optly_user&client_secret=client_seekrit", nil)
	req.Header.Set(middleware.OptlySDKHeader, "123")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusForbidden, rec.Code)
}

func (s *OAuthTestSuite) TestGetAdminAccessTokenInvalidClientId() {
	req := httptest.NewRequest("GET", "/admin/token?grant_type=client_credentials&client_id=not_an_optly_user&client_secret=client_seekrit", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusForbidden, rec.Code)
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenInvalidClientSecret() {
	req := httptest.NewRequest("GET", "/api/token?grant_type=client_credentials&client_id=optly_user&client_secret=invalid_seekret", nil)
	req.Header.Set(middleware.OptlySDKHeader, "123")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusForbidden, rec.Code)
}

func (s *OAuthTestSuite) TestGetAdminAccessTokenInvalidClientSecret() {
	req := httptest.NewRequest("GET", "/admin/token?grant_type=client_credentials&client_id=optly_user&client_secret=invalid_seekret", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusForbidden, rec.Code)
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenMissingSDKKey() {
	req := httptest.NewRequest("GET", "/api/token?grant_type=client_credentials&client_id=optly_user&client_secret=client_seekrit", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *OAuthTestSuite) TestGetAPIAccessTokenSuccess() {
	req := httptest.NewRequest("GET", "/api/token?grant_type=client_credentials&client_id=optly_user&client_secret=client_seekrit", nil)
	req.Header.Set(middleware.OptlySDKHeader, "123")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	var actual tokenResponse
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	s.NoError(err)
}

func (s *OAuthTestSuite) TestGetAdminAccessTokenSuccess() {
	req := httptest.NewRequest("GET", "/admin/token?grant_type=client_credentials&client_id=optly_user&client_secret=client_seekrit", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	var actual tokenResponse
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	s.NoError(err)
}

func TestOAuthTestSuite(t *testing.T) {
	suite.Run(t, new(OAuthTestSuite))
}
