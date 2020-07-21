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
	s.handler = AllowedHosts([]string{"76.125.27.44", "example.com"}, "8080", true)(okHandler)
}

func (s *AllowedHostsTestSuite) TestURLHostAndPort() {
	scenarios := []struct {
		inputUrl       string
		expectedStatus int
	}{
		// matches first allowedHost, expect StatusOK
		{"https://76.125.27.44:8080/v1/config", http.StatusOK},
		// matches second allowedHost, expect StatusOK
		{"https://example.com:8080/v1/config", http.StatusOK},
		// wrong URL port, expect http.StatusNotFound
		{"https://76.125.27.44:1000/v1/config", http.StatusNotFound},
		// wrong URL host, expect http.StatusNotFound
		{ "https://evil.com:8080/v1/config", http.StatusNotFound},
	}

	for _, scenario := range scenarios {
		req := httptest.NewRequest("GET", scenario.inputUrl, nil)
		rec := httptest.NewRecorder()
		s.handler.ServeHTTP(rec, req)
		s.Equal(scenario.expectedStatus, rec.Code)
	}
}

func (s *AllowedHostsTestSuite) TestCustomHeaders() {
	scenarios := []struct {
		inputHeaderKey string
		inputHeaderVal string
		expectedStatus int
	} {
		// X-Forwarded-Host is valid, expect http.statusOK
		{"X-Forwarded-Host", "example.com:8080", http.StatusOK},
		// X-Forwarded-Host is invalid, expect http.statusNotFound
		{"X-Forwarded-Host", "evil.com:8080", http.StatusNotFound},
		// Forwarded is valid, expect http.statusOK
		{"Forwarded", "host=76.125.27.44:8080", http.StatusOK},
		// Forwarded is invalid, expect http.statusOK
		{"Forwarded", "host=77.125.26.44:8080", http.StatusNotFound},
	}

	for _, scenario := range scenarios {
		req := httptest.NewRequest("GET", "https://company-proxy.com:8080/v1/config", nil)
		req.Header.Set(scenario.inputHeaderKey, scenario.inputHeaderVal)
		rec := httptest.NewRecorder()
		s.handler.ServeHTTP(rec, req)
		s.Equal(scenario.expectedStatus, rec.Code)
	}
}

func (s *AllowedHostsTestSuite) TestDefaultPortNoTLSValid() {
	noTLSHandler := AllowedHosts([]string{"76.125.27.44", "example.com"}, "80", false)(okHandler)
	// URL contains no explicit port. Request should be allowed as server is running on port 80 with no TLS.
	req := httptest.NewRequest("GET", "http://76.125.27.44/v1/config", nil)
	rec := httptest.NewRecorder()
	noTLSHandler.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *AllowedHostsTestSuite) TestDefaultPortWithTLSValid() {
	noTLSHandler := AllowedHosts([]string{"76.125.27.44", "example.com"}, "443", true)(okHandler)
	// URL contains no explicit port. Request should be allowed as server is running on port 443 with TLS.
	req := httptest.NewRequest("GET", "http://76.125.27.44/v1/config", nil)
	rec := httptest.NewRecorder()
	noTLSHandler.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func TestAllowedHostsTestSuite(t *testing.T) {
	suite.Run(t, new(AllowedHostsTestSuite))
}
