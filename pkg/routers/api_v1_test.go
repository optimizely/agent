/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                        *
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
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
)

const methodHeaderKey = "X-Method-Header"

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
	}
}

func TestAPIV1TestSuite(t *testing.T) {
	suite.Run(t, new(APIV1TestSuite))
}

func TestNewDefaultClientRouter(t *testing.T) {
	client := NewDefaultAPIRouter(MockCache{}, config.APIConfig{}, metricsRegistry)
	assert.NotNil(t, client)
}
