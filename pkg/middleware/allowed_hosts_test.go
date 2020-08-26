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
	s.handler = AllowedHosts([]string{"76.125.27.44", "example.com"})(okHandler)
}

func (s *AllowedHostsTestSuite) TestURLHost() {
	scenarios := []struct {
		inputUrl       string
		expectedStatus int
	}{
		// matches first allowedHost, expect StatusOK
		{"https://76.125.27.44:8080/v1/config", http.StatusOK},
		{"https://76.125.27.44/v1/config", http.StatusOK},

		// matches second allowedHost, expect StatusOK
		{"https://example.com:8080/v1/config", http.StatusOK},
		{"https://example.com/v1/config", http.StatusOK},

		// wrong URL host, expect http.StatusNotFound
		{ "https://evil.com:8080/v1/config", http.StatusNotFound},
		{ "https://evil.com/v1/config", http.StatusNotFound},
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
		{"X-Forwarded-Host", "example.com", http.StatusOK},

		// X-Forwarded-Host is invalid, expect http.statusNotFound
		{"X-Forwarded-Host", "evil.com:8080", http.StatusNotFound},
		{"X-Forwarded-Host", "evil.com", http.StatusNotFound},

		// Forwarded is valid, expect http.statusOK
		{"Forwarded", "host=76.125.27.44:8080", http.StatusOK},
		{"Forwarded", "host=76.125.27.44", http.StatusOK},

		// Forwarded is invalid, expect http.statusOK
		{"Forwarded", "host=77.125.26.44:8080", http.StatusNotFound},
		{"Forwarded", "host=77.125.26.44", http.StatusNotFound},
	}

	for _, scenario := range scenarios {
		req := httptest.NewRequest("GET", "https://company-proxy.com:8080/v1/config", nil)
		req.Header.Set(scenario.inputHeaderKey, scenario.inputHeaderVal)
		rec := httptest.NewRecorder()
		s.handler.ServeHTTP(rec, req)
		s.Equal(scenario.expectedStatus, rec.Code)
	}
}

func TestAllowedHostsTestSuite(t *testing.T) {
	suite.Run(t, new(AllowedHostsTestSuite))
}
