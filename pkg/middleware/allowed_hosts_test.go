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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

var okHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
})

type AllowedHostsTestSuite struct {
	suite.Suite
	handler http.Handler
}

func (s *AllowedHostsTestSuite) SetupTest() {
	s.handler = AllowedHosts([]string{"76.125.27.44", "example.com"}, "8080")(okHandler)
}

func (s *AllowedHostsTestSuite) TestRequestHostMatchesFirstAllowed() {
	req := httptest.NewRequest("GET", "https://76.125.27.44:8080/v1/config", nil)
	rec := httptest.NewRecorder()
	s.handler.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *AllowedHostsTestSuite) TestRequestHostMatchesSecondAllowed() {
	req := httptest.NewRequest("GET", "https://example.com:8080/v1/config", nil)
	rec := httptest.NewRecorder()
	s.handler.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *AllowedHostsTestSuite) TestRequestHostWrongPort() {
	req := httptest.NewRequest("GET", "https://76.125.27.44:1000/v1/config", nil)
	rec := httptest.NewRecorder()
	s.handler.ServeHTTP(rec, req)
	s.Equal(http.StatusNotFound, rec.Code)
}

func (s *AllowedHostsTestSuite) TestRequestHostWrongAddr() {
	req := httptest.NewRequest("GET", "https://evil.com:8080/v1/config", nil)
	rec := httptest.NewRecorder()
	s.handler.ServeHTTP(rec, req)
	s.Equal(http.StatusNotFound, rec.Code)
}

func (s *AllowedHostsTestSuite) TestXForwardedHostValid() {
	req := httptest.NewRequest("GET", "https://company-proxy.com:8080/v1/config", nil)
	req.Header.Set("X-Forwarded-Host", "example.com:8080")
	rec := httptest.NewRecorder()
	s.handler.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *AllowedHostsTestSuite) TestXForwardedHostInvalid() {
	req := httptest.NewRequest("GET", "https://company-proxy.com:8080/v1/config", nil)
	req.Header.Set("X-Forwarded-Host", "evil.com:8080")
	rec := httptest.NewRecorder()
	s.handler.ServeHTTP(rec, req)
	s.Equal(http.StatusNotFound, rec.Code)
}

func (s *AllowedHostsTestSuite) TestForwardedHostValid() {
	req := httptest.NewRequest("GET", "https://company-proxy.com:8080/v1/config", nil)
	req.Header.Set("Forwarded", "host=76.125.27.44:8080")
	rec := httptest.NewRecorder()
	s.handler.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *AllowedHostsTestSuite) TestForwardedHostInvalid() {
	req := httptest.NewRequest("GET", "https://company-proxy.com:8080/v1/config", nil)
	req.Header.Set("Forwarded", "host=77.125.26.44:8080")
	rec := httptest.NewRecorder()
	s.handler.ServeHTTP(rec, req)
	s.Equal(http.StatusNotFound, rec.Code)
}

func (s *AllowedHostsTestSuite) TestDefaultPortHTTPHostValid() {
	defaultPortHTTPHandler := AllowedHosts([]string{"76.125.27.44", "example.com"}, "80")(okHandler)
	// Req URL omits port. Default port for scheme http is 80. Since allowed port is 80, this request should be allowed
	req := httptest.NewRequest("GET", "http://76.125.27.44/v1/config", nil)
	rec := httptest.NewRecorder()
	defaultPortHTTPHandler.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *AllowedHostsTestSuite) TestDefaultPortHTTPHostInvalid() {
	defaultPortHTTPHandler := AllowedHosts([]string{"76.125.27.44", "example.com"}, "9000")(okHandler)
	// Req URL omits port. Default port for scheme http is 80. Since allowed port is 9000, this request should not be
	// allowed
	req := httptest.NewRequest("GET", "http://76.125.27.44/v1/config", nil)
	rec := httptest.NewRecorder()
	defaultPortHTTPHandler.ServeHTTP(rec, req)
	s.Equal(http.StatusNotFound, rec.Code)
}

func (s *AllowedHostsTestSuite) TestDefaultPortHTTPSHostValid() {
	defaultPortHTTPSHandler := AllowedHosts([]string{"76.125.27.44", "example.com"}, "443")(okHandler)
	// Req URL omits port. Default port for scheme https is 443. Since allowed port is 443, this request should be
	// allowed
	req := httptest.NewRequest("GET", "https://76.125.27.44/v1/config", nil)
	rec := httptest.NewRecorder()
	defaultPortHTTPSHandler.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *AllowedHostsTestSuite) TestDefaultPortHTTPSHostInvalid() {
	defaultPortHTTPSHandler := AllowedHosts([]string{"76.125.27.44", "example.com"}, "9000")(okHandler)
	// Req URL omits port. Default port for scheme https is 443. Since allowed port is 9000, this request should not be
	// allowed
	req := httptest.NewRequest("GET", "https://76.125.27.44/v1/config", nil)
	rec := httptest.NewRecorder()
	defaultPortHTTPSHandler.ServeHTTP(rec, req)
	s.Equal(http.StatusNotFound, rec.Code)
}

func TestAllowedHostsTestSuite(t *testing.T) {
	suite.Run(t, new(AllowedHostsTestSuite))
}
