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

// Package routers //
package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/metrics"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
)

var metricsRegistry = metrics.NewRegistry()

const methodHeaderKey = "X-Method-Header"
const clientHeaderKey = "X-Client-Header"

type MockOptlyMiddleware struct{}

func (m *MockOptlyMiddleware) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(clientHeaderKey, "expected")
		next.ServeHTTP(w, r)
	})
}

type MockCache struct{}

func (m MockCache) GetClient(sdkKey string) (*optimizely.OptlyClient, error) {
	panic("implement me")
}

type MockHandlers struct{}

func (m MockHandlers) config(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(methodHeaderKey, "config")
}

func (m MockHandlers) activate(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(methodHeaderKey, "activate")
}

func (m MockHandlers) trackEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(methodHeaderKey, "track")
}

func (m MockHandlers) override(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(methodHeaderKey, "override")
}

var testAuthHandler = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(methodHeaderKey, "oauth/token")
}

const middlewareHeaderKey = "X-Middleware-Header"

var testAuthMiddleware = func(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(middlewareHeaderKey, "mockMiddleware")
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

type APIV1TestSuite struct {
	suite.Suite
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

func (suite *APIV1TestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	suite.tc = testClient

	opts := &APIV1Options{
		maxConns:        1,
		middleware:      &MockOptlyMiddleware{},
		handlers:        MockHandlers{},
		metricsRegistry: metricsRegistry,
		enableOverrides: true,
		oAuthHandler:    testAuthHandler,
		oAuthMiddleware: testAuthMiddleware,
	}

	suite.mux = NewAPIV1Router(opts)
}

func (suite *APIV1TestSuite) TestOverride() {

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "config"},
		{"POST", "activate"},
		{"POST", "track"},
		{"POST", "override"},
	}

	for _, route := range routes {
		req := httptest.NewRequest(route.method, "/v1/"+route.path, nil)
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)
		suite.Equal(http.StatusOK, rec.Code)

		suite.Equal("expected", rec.Header().Get(clientHeaderKey))
		suite.Equal(route.path, rec.Header().Get(methodHeaderKey))
		suite.Equal("mockMiddleware", rec.Header().Get(middlewareHeaderKey))
	}
}

func (suite *APIV1TestSuite) TestDisabledOverride() {

	route := struct {
		method string
		path   string
	}{"POST", "override"}

	opts := &APIV1Options{
		maxConns:        1,
		middleware:      &MockOptlyMiddleware{},
		handlers:        MockHandlers{},
		metricsRegistry: metricsRegistry,
		enableOverrides: false,
		oAuthHandler:    testAuthHandler,
		oAuthMiddleware: testAuthMiddleware,
	}

	mux := NewAPIV1Router(opts)

	req := httptest.NewRequest(route.method, "/v1/"+route.path, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusForbidden, rec.Code)

	// Unmarshal response
	response := string(rec.Body.Bytes())
	suite.Equal("Overrides not enabled\n", response)

}

func (suite *APIV1TestSuite) TestCreateAccessToken() {
	req := httptest.NewRequest("POST", "/oauth/token", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
	suite.Equal("expected", rec.Header().Get(clientHeaderKey))
	suite.Equal("oauth/token", rec.Header().Get(methodHeaderKey))
}

func TestAPIV1TestSuite(t *testing.T) {
	suite.Run(t, new(APIV1TestSuite))
}

func TestNewDefaultAPIV1Router(t *testing.T) {
	client := NewDefaultAPIV1Router(MockCache{}, config.APIConfig{}, metricsRegistry)
	assert.NotNil(t, client)
}

func TestNewDefaultAPIV1RouterInvalidHandlerConfig(t *testing.T) {
	invalidAPIConfig := config.APIConfig{
		Auth: config.ServiceAuthConfig{
			Clients: []config.OAuthClientCredentials{
				{
					ID:         "id1",
					SecretHash: "JDJhJDEyJFBQM3dSdnNERnVSQmZPNnA4MGcvLk9Eb1RVWExYMm5FZ2VhZXpsS1VmR3hPdFJUT3ViaXVX",
				},
			},
			// Empty HMACSecrets, but non-empty Clients, is an invalid config
			HMACSecrets:        []string{},
			TTL:                0,
			JwksURL:            "",
			JwksUpdateInterval: 0,
		},
		MaxConns:            100,
		Port:                "8080",
		EnableNotifications: false,
		EnableOverrides:     false,
	}
	client := NewDefaultAPIV1Router(MockCache{}, invalidAPIConfig, metricsRegistry)
	assert.Nil(t, client)
}

func TestNewDefaultClientRouterInvalidMiddlewareConfig(t *testing.T) {
	invalidAPIConfig := config.APIConfig{
		Auth: config.ServiceAuthConfig{
			JwksURL:            "not-valid",
			JwksUpdateInterval: 0,
		},
		MaxConns:            100,
		Port:                "8080",
		EnableNotifications: false,
		EnableOverrides:     false,
	}
	client := NewDefaultAPIV1Router(MockCache{}, invalidAPIConfig, metricsRegistry)
	assert.Nil(t, client)
}
