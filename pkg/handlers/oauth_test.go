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

func (s *OAuthTestSuite) TestVerifyClientCredentialsSuccess() {
	req := httptest.NewRequest("GET", "/api/token?grant_type=client_credentials&client_id=optly_user&client_secret=client_seekrit", nil)
	clientCreds, httpCode, err := s.handler.verifyClientCredentials(req)
	s.Equal(&ClientCredentials{ID: "optly_user", TTL: 30 * time.Minute, Secret: []byte("client_seekrit")}, clientCreds)
	s.Equal(http.StatusOK, httpCode)
	s.NoError(err)
}

func (s *OAuthTestSuite) TestVerifyClientCredentialsMissingGrantType() {
	req := httptest.NewRequest("GET", "/api/token?client_id=optly_user&client_secret=client_seekrit", nil)
	_, httpCode, err := s.handler.verifyClientCredentials(req)
	s.Equal(http.StatusBadRequest, httpCode)
	s.Error(err)
}

func (s *OAuthTestSuite) TestVerifyClientCredentialsUnsupportedGrantType() {
	req := httptest.NewRequest("GET", "/api/token?grant_type=authorizationclient_id=optly_user&client_secret=client_seekrit", nil)
	_, httpCode, err := s.handler.verifyClientCredentials(req)
	s.Equal(http.StatusBadRequest, httpCode)
	s.Error(err)
}

func (s *OAuthTestSuite) TestVerifyClientCredentialsMissingClientId() {
	req := httptest.NewRequest("GET", "/api/token?grant_type=client_credentials&client_secret=client_seekrit", nil)
	_, httpCode, err := s.handler.verifyClientCredentials(req)
	s.Equal(http.StatusUnauthorized, httpCode)
	s.Error(err)
}

func (s *OAuthTestSuite) TestVerifyClientCredentialsMissingClientSecret() {
	req := httptest.NewRequest("GET", "/api/token?grant_type=client_credentials&client_id=optly_user", nil)
	_, httpCode, err := s.handler.verifyClientCredentials(req)
	s.Equal(http.StatusUnauthorized, httpCode)
	s.Error(err)
}

func (s *OAuthTestSuite) TestVerifyClientCredentialsInvalidClientId() {
	req := httptest.NewRequest("GET", "/api/token?grant_type=client_credentials&client_id=not_an_optly_user&client_secret=client_seekrit", nil)
	_, httpCode, err := s.handler.verifyClientCredentials(req)
	s.Equal(http.StatusForbidden, httpCode)
	s.Error(err)
}

func (s *OAuthTestSuite) TestVerifyClientCredentialsInvalidClientSecret() {
	req := httptest.NewRequest("GET", "/api/token?grant_type=client_credentials&client_id=optly_user&client_secret=invalid_seekret", nil)
	_, httpCode, err := s.handler.verifyClientCredentials(req)
	s.Equal(http.StatusForbidden, httpCode)
	s.Error(err)
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

func (s *OAuthTestSuite) TestGetAPIAccessTokenForbidden() {
	req := httptest.NewRequest("GET", "/api/token?grant_type=client_credentials&client_id=optly_user&client_secret=blablabla", nil)
	req.Header.Set(middleware.OptlySDKHeader, "123")
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

func (s *OAuthTestSuite) TestGetAdminAccessTokenSuccess() {
	req := httptest.NewRequest("GET", "/admin/token?grant_type=client_credentials&client_id=optly_user&client_secret=client_seekrit", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	var actual tokenResponse
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	s.NoError(err)
}

func (s *OAuthTestSuite) TestGetAdminAccessTokenForbidden() {
	req := httptest.NewRequest("GET", "/admin/token?grant_type=client_credentials&client_id=not_an_optly_user&client_secret=client_seekrit", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	s.Equal(http.StatusForbidden, rec.Code)
}

func TestOAuthTestSuite(t *testing.T) {
	suite.Run(t, new(OAuthTestSuite))
}
