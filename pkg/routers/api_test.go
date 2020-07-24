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
	"bytes"
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
const contentTypeHeaderKey = "Content-Type"
const originHeaderKey = "Origin"
const corsOriginHeaderKey = "Access-Control-Allow-Origin"
const corsRequestMethodHeaderKey = "Access-Control-Request-Method"
const corsAllowMethodHeaderKey = "Access-Control-Allow-Methods"
const corsExposeHeaderKey = "Access-Control-Expose-Headers"
const corsRequestHeadersKey = "Access-Control-Request-Headers"
const corsAllowedHeadersKey = "Access-Control-Allow-Headers"
const corsCredentialsHeaderKey = "Access-Control-Allow-Credentials"
const corsMaxAgeHeaderKey = "Access-Control-Max-Age"

const validOrigin = "http://localhost.com"

var testOptlyMiddleware = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(clientHeaderKey, "expected")
		next.ServeHTTP(w, r)
	})
}

type MockCache struct{}

func (m MockCache) GetClient(_ string) (*optimizely.OptlyClient, error) {
	return &optimizely.OptlyClient{}, nil
}

func (m MockCache) UpdateConfigs(_ string) {
}

var testHandler = func(val string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(methodHeaderKey, val)
	}
}

const middlewareHeaderKey = "X-Middleware-Header"

var testAuthMiddleware = func(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(middlewareHeaderKey, "mockMiddleware")
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

var opts *APIOptions

var corsConfig = config.CORSConfig{
	AllowedOrigins:     []string{validOrigin},
	AllowedMethods:     []string{"OPTIONS", "GET", "POST"},
	AllowedHeaders:     []string{"Origin", "Accept", "Content-Type"},
	ExposedHeaders:     []string{"Header1", "Header2"},
	AllowedCredentials: true,
	MaxAge:             500,
}

var testCorsHandler = createCorsHandler(corsConfig)

type APIV1TestSuite struct {
	suite.Suite
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

func (suite *APIV1TestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	suite.tc = testClient

	opts = &APIOptions{
		maxConns:        1,
		sdkMiddleware:   testOptlyMiddleware,
		configHandler:   testHandler("config"),
		activateHandler: testHandler("activate"),
		overrideHandler: testHandler("override"),
		trackHandler:    testHandler("track"),
		nStreamHandler:  testHandler("notifications/event-stream"),
		oAuthHandler:    testHandler("oauth/token"),
		oAuthMiddleware: testAuthMiddleware,
		metricsRegistry: metricsRegistry,
		corsHandler:     testCorsHandler,
	}

	suite.mux = NewAPIRouter(opts)
}

func (suite *APIV1TestSuite) TestValidRoutes() {

	opts.corsHandler = func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(originHeaderKey, "corsMiddleware")
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
	suite.mux = NewAPIRouter(opts)

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "config"},
		{"POST", "activate"},
		{"POST", "track"},
		{"POST", "override"},
		{"GET", "notifications/event-stream"},
	}

	for _, route := range routes {
		req := httptest.NewRequest(route.method, "/v1/"+route.path, nil)
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)
		suite.Equal(http.StatusOK, rec.Code)

		suite.Equal("expected", rec.Header().Get(clientHeaderKey))
		suite.Equal(route.path, rec.Header().Get(methodHeaderKey))
		suite.Equal("mockMiddleware", rec.Header().Get(middlewareHeaderKey))
		suite.Equal("corsMiddleware", rec.Header().Get(originHeaderKey))
	}
}

func (suite *APIV1TestSuite) TestStaticContent() {
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"POST", "/openapi.yaml"},
	}

	for _, route := range routes {
		req := httptest.NewRequest(route.method, route.path, nil)
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)
		suite.Equal(http.StatusOK, rec.Code)
	}
}

func (suite *APIV1TestSuite) TestCreateAccessToken() {
	req := httptest.NewRequest("POST", "/oauth/token", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
	suite.Equal("oauth/token", rec.Header().Get(methodHeaderKey))
}

func (suite *APIV1TestSuite) TestCORSAllowedOrigins() {
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "config"},
		{"POST", "activate"},
		{"POST", "track"},
		{"POST", "override"},
		{"GET", "notifications/event-stream"},
	}

	// Allowed Origin
	for _, route := range routes {
		req := httptest.NewRequest(route.method, "/v1/"+route.path, nil)
		req.Header.Add(originHeaderKey, validOrigin)
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)
		suite.Equal(http.StatusOK, rec.Code)
		suite.Equal(validOrigin, rec.Header().Get(corsOriginHeaderKey))
	}

	// Unallowed Origin
	for _, route := range routes {
		req := httptest.NewRequest(route.method, "/v1/"+route.path, nil)
		req.Header.Add(originHeaderKey, "http://test.com")
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)
		suite.Equal(http.StatusOK, rec.Code)
		suite.Equal("", rec.Header().Get(corsOriginHeaderKey))
	}
}

func (suite *APIV1TestSuite) TestCORSAllowedMethods() {

	// Allowed Method
	req := httptest.NewRequest("OPTIONS", "/v1/config", nil)
	req.Header.Add(originHeaderKey, validOrigin)
	req.Header.Add(corsRequestMethodHeaderKey, "GET")
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
	suite.Equal("GET", rec.Header().Get(corsAllowMethodHeaderKey))

	// Unallowed Method
	req = httptest.NewRequest("OPTIONS", "/v1/config", nil)
	req.Header.Add(originHeaderKey, validOrigin)
	req.Header.Add(corsRequestMethodHeaderKey, "PATCH")
	rec = httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
	suite.Equal("", rec.Header().Get(corsAllowMethodHeaderKey))
}

func (suite *APIV1TestSuite) TestCORSAllowedHeaders() {

	// Allowed Headers
	req := httptest.NewRequest("OPTIONS", "/v1/config", nil)
	req.Header.Add(originHeaderKey, validOrigin)
	req.Header.Add(corsRequestMethodHeaderKey, "GET")
	req.Header.Add(corsRequestHeadersKey, "Accept")
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
	suite.Equal("Accept", rec.Header().Get(corsAllowedHeadersKey))

	// Unallowed Headers
	req = httptest.NewRequest("OPTIONS", "/v1/config", nil)
	req.Header.Add(originHeaderKey, validOrigin)
	req.Header.Add(corsRequestMethodHeaderKey, "GET")
	req.Header.Add(corsRequestHeadersKey, "Invalid")
	rec = httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
	suite.Equal("", rec.Header().Get(corsAllowedHeadersKey))
}

func (suite *APIV1TestSuite) TestCORSExposedHeaders() {
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "config"},
		{"POST", "activate"},
	}

	for _, route := range routes {
		req := httptest.NewRequest(route.method, "/v1/"+route.path, nil)
		req.Header.Add(originHeaderKey, validOrigin)
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)
		suite.Equal(http.StatusOK, rec.Code)
		suite.Equal("Header1, Header2", rec.Header().Get(corsExposeHeaderKey))
	}
}

func (suite *APIV1TestSuite) TestCORSAllowCredentials() {
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "config"},
		{"POST", "activate"},
	}

	for _, route := range routes {
		req := httptest.NewRequest(route.method, "/v1/"+route.path, nil)
		req.Header.Add(originHeaderKey, validOrigin)
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)
		suite.Equal(http.StatusOK, rec.Code)
		suite.Equal("true", rec.Header().Get(corsCredentialsHeaderKey))
	}
}

func (suite *APIV1TestSuite) TestCORSMaxAge() {
	req := httptest.NewRequest("OPTIONS", "/v1/config", nil)
	req.Header.Add(originHeaderKey, validOrigin)
	req.Header.Add(corsRequestMethodHeaderKey, "GET")
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
	suite.Equal("500", rec.Header().Get(corsMaxAgeHeaderKey))
}

func TestAPIV1TestSuite(t *testing.T) {
	suite.Run(t, new(APIV1TestSuite))
}

func TestNewDefaultAPIV1Router(t *testing.T) {
	client := NewDefaultAPIRouter(MockCache{}, config.APIConfig{}, metricsRegistry)
	assert.NotNil(t, client)
}

func TestNewDefaultAPIV1RouterInvalidHandlerConfig(t *testing.T) {
	invalidAPIConfig := config.APIConfig{
		Auth: config.ServiceAuthConfig{
			Clients: []config.OAuthClientCredentials{
				{
					ID:         "id1",
					SecretHash: "JDJhJDEyJFBQM3dSdnNERnVSQmZPNnA4MGcvLk9Eb1RVWExYMm5FZ2VhZXpsS1VmR3hPdFJUT3ViaXVX",
					SDKKeys:    []string{"123"},
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
	client := NewDefaultAPIRouter(MockCache{}, invalidAPIConfig, metricsRegistry)
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
	client := NewDefaultAPIRouter(MockCache{}, invalidAPIConfig, metricsRegistry)
	assert.Nil(t, client)
}

func TestForbiddenRoutes(t *testing.T) {
	conf := config.APIConfig{}
	mux := NewDefaultAPIRouter(MockCache{}, conf, metricsRegistry)

	routes := []struct {
		method string
		path   string
		error  string
	}{
		{"POST", "override", "Overrides not enabled\n"},
		{"GET", "notifications/event-stream", "Notification stream not enabled\n"},
	}

	for _, route := range routes {
		req := httptest.NewRequest(route.method, "/v1/"+route.path, nil)
		req.Header.Add("X-Optimizely-SDK-Key", "something")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)

		response := string(rec.Body.Bytes())
		assert.Equal(t, route.error, response)
	}
}

func (suite *APIV1TestSuite) TestAllowedContentTypeMiddleware() {

	routes := []struct {
		method string
		path   string
	}{
		{"POST", "activate"},
		{"POST", "track"},
		{"POST", "override"},
	}

	for _, route := range routes {

		// Testing unsupported content type
		body := "<request> <parameters> <email>test@123.com</email> </parameters> </request>"
		req := httptest.NewRequest(route.method, "/v1/"+route.path, bytes.NewBuffer([]byte(body)))
		req.Header.Add(contentTypeHeaderKey, "application/xml")
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)
		suite.Equal(http.StatusUnsupportedMediaType, rec.Code)

		// Testing supported content type
		body = `{"email":"test@123.com"}`
		req = httptest.NewRequest(route.method, "/v1/"+route.path, bytes.NewBuffer([]byte(body)))
		req.Header.Add(contentTypeHeaderKey, "application/json")
		rec = httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)
		suite.Equal(http.StatusOK, rec.Code)
	}
}
