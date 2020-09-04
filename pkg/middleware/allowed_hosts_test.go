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

	"github.com/stretchr/testify/assert"
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
		{"https://evil.com:8080/v1/config", http.StatusNotFound},
		{"https://evil.com/v1/config", http.StatusNotFound},
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
	}{
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

func TestAllowedHostsSuffixMatching(t *testing.T) {
	handler := AllowedHosts([]string{".125.27.44", ".mydomain.example.com"})(okHandler)
	scenarios := []struct {
		inputUrl       string
		expectedStatus int
	}{
		// subdomains of .125.27.44 should be allowed
		{"https://76.125.27.44:8080/v1/config", http.StatusOK},
		{"https://123.86.125.27.44/v1/config", http.StatusOK},

		// subdomains of .mydomain.example.com should be allowed
		{"https://hello.mydomain.example.com:8080/v1/config", http.StatusOK},
		{"https://opti.mizely.mydomain.example.com/v1/config", http.StatusOK},

		// Non-matches should be rejected
		{"https://evil.com:8080/v1/config", http.StatusNotFound},
		{"https://hello.evil.com/v1/config", http.StatusNotFound},
		{"https://123.86.125.28.44/v1/config", http.StatusNotFound},
		{"https://opti.mydomain.example.com.biz", http.StatusNotFound},
	}
	for _, scenario := range scenarios {
		req := httptest.NewRequest("GET", scenario.inputUrl, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, scenario.expectedStatus, rec.Code)
	}

	handler = AllowedHosts([]string{".example.com", ".net"})(okHandler)
	// Subdomains of .net and .example.com should be allowed
	urls := []string{"http://ab.cd.ef.g.example.com", "http://example.net", "http://my.example.net"}
	for _, url := range urls {
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestAllowedHostsAllowAll(t *testing.T) {
	handler := AllowedHosts([]string{"."})(okHandler)
	// Any host should be allowed
	urls := []string{
		"https://opti.example.com/v1/config",
		"https://heyyo.some.domain/v1/config",
		"https://evil.com:8080/v1/config",
		"https://hello.evil.com/v1/config",
		"https://76.125.27.44:8080/v1/config",
	}
	for _, url := range urls {
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}
