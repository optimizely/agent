/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
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

// Package api //
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/optimizely/sidedoor/pkg/optimizelytest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const clientHeaderKey = "X-Client-Header"
const userHeaderKey = "X-User-Header"

type MockOptlyMiddleware struct{}

func (m *MockOptlyMiddleware) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(clientHeaderKey, "expected")
		next.ServeHTTP(w, r)
	})
}

func (m *MockOptlyMiddleware) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(userHeaderKey, "expected")
		next.ServeHTTP(w, r)
	})
}

type MockExperimentAPI struct{}

func (m *MockExperimentAPI) ListExperiments(w http.ResponseWriter, r *http.Request) {
	renderPathParams(w, r)
}
func (m *MockExperimentAPI) GetExperiment(w http.ResponseWriter, r *http.Request) {
	renderPathParams(w, r)
}

type MockFeatureAPI struct{}

func (m *MockFeatureAPI) ListFeatures(w http.ResponseWriter, r *http.Request) {
	renderPathParams(w, r)
}
func (m *MockFeatureAPI) GetFeature(w http.ResponseWriter, r *http.Request) {
	renderPathParams(w, r)
}

type MockUserEventAPI struct{}

func (m *MockUserEventAPI) AddUserEvent(w http.ResponseWriter, r *http.Request) {
	renderPathParams(w, r)
}

type MockUserAPI struct{}

func (m *MockUserAPI) TrackEvent(w http.ResponseWriter, r *http.Request) {
	renderPathParams(w, r)
}

func (m *MockUserAPI) GetFeature(w http.ResponseWriter, r *http.Request) {
	renderPathParams(w, r)
}

func (m *MockUserAPI) TrackFeature(w http.ResponseWriter, r *http.Request) {
	renderPathParams(w, r)
}

func (m *MockUserAPI) ListFeatures(w http.ResponseWriter, r *http.Request) {
	renderPathParams(w, r)
}

func (m *MockUserAPI) TrackFeatures(w http.ResponseWriter, r *http.Request) {
	renderPathParams(w, r)
}

func (m *MockUserAPI) SetForcedVariation(w http.ResponseWriter, r *http.Request) {
	renderPathParams(w, r)
}

func (m *MockUserAPI) RemoveForcedVariation(w http.ResponseWriter, r *http.Request) {
	renderPathParams(w, r)
}

func (m *MockUserAPI) GetVariation(w http.ResponseWriter, r *http.Request) {
	renderPathParams(w, r)
}

func (m *MockUserAPI) ActivateExperiment(w http.ResponseWriter, r *http.Request) {
	renderPathParams(w, r)
}

func renderPathParams(w http.ResponseWriter, r *http.Request) {
	pathParams := make(map[string]string)
	rctx := chi.RouteContext(r.Context())
	for i, k := range rctx.URLParams.Keys {
		if k == "*" {
			continue
		}
		pathParams[k] = rctx.URLParams.Values[i]
	}

	render.JSON(w, r, pathParams)
}

type RouterTestSuite struct {
	suite.Suite
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

func (suite *RouterTestSuite) SetupTest() {

	testClient := optimizelytest.NewClient()
	suite.tc = testClient

	opts := &RouterOptions{
		maxConns:      1,
		experimentAPI: new(MockExperimentAPI),
		featureAPI:    new(MockFeatureAPI),
		userAPI:       new(MockUserAPI),
		middleware:    new(MockOptlyMiddleware),
	}

	suite.mux = NewRouter(opts)
}

func (suite *RouterTestSuite) TestListFeatures() {
	req := httptest.NewRequest("GET", "/features", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal("expected", rec.Header().Get(clientHeaderKey))
	suite.Empty(rec.Header().Get(userHeaderKey))

	expected := map[string]string{}
	suite.assertValid(rec, expected)
}

func (suite *RouterTestSuite) TestGetFeature() {
	req := httptest.NewRequest("GET", "/features/one", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal("expected", rec.Header().Get(clientHeaderKey))
	suite.Empty(rec.Header().Get(userHeaderKey))

	expected := map[string]string{
		"featureKey": "one",
	}
	suite.assertValid(rec, expected)
}

func (suite *RouterTestSuite) TestGetUserFeatures() {
	req := httptest.NewRequest("GET", "/users/me/features", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal("expected", rec.Header().Get(clientHeaderKey))
	suite.Equal("expected", rec.Header().Get(userHeaderKey))

	expected := map[string]string{
		"userID": "me",
	}
	suite.assertValid(rec, expected)
}

func (suite *RouterTestSuite) TestGetUserFeature() {
	req := httptest.NewRequest("GET", "/users/me/features/one", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal("expected", rec.Header().Get(clientHeaderKey))
	suite.Equal("expected", rec.Header().Get(userHeaderKey))

	expected := map[string]string{
		"userID":     "me",
		"featureKey": "one",
	}
	suite.assertValid(rec, expected)
}

func (suite *RouterTestSuite) TestTrackUserFeatures() {
	req := httptest.NewRequest("POST", "/users/me/features", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal("expected", rec.Header().Get(clientHeaderKey))
	suite.Equal("expected", rec.Header().Get(userHeaderKey))

	expected := map[string]string{
		"userID": "me",
	}
	suite.assertValid(rec, expected)
}

func (suite *RouterTestSuite) TestTrackEvent() {
	req := httptest.NewRequest("POST", "/users/me/events/key", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal("expected", rec.Header().Get(clientHeaderKey))
	suite.Equal("expected", rec.Header().Get(userHeaderKey))

	expected := map[string]string{
		"userID":   "me",
		"eventKey": "key",
	}
	suite.assertValid(rec, expected)
}

func (suite *RouterTestSuite) TestGetVariation() {
	req := httptest.NewRequest("GET", "/users/me/experiments/key", nil)
	rec := httptest.NewRecorder()

	suite.mux.ServeHTTP(rec, req)

	suite.Equal("expected", rec.Header().Get(clientHeaderKey))
	suite.Equal("expected", rec.Header().Get(userHeaderKey))

	expected := map[string]string{
		"userID":        "me",
		"experimentKey": "key",
	}
	suite.assertValid(rec, expected)
}

func (suite *RouterTestSuite) TestActivateExperiment() {
	req := httptest.NewRequest("POST", "/users/me/experiments/key", nil)
	rec := httptest.NewRecorder()

	suite.mux.ServeHTTP(rec, req)

	suite.Equal("expected", rec.Header().Get(clientHeaderKey))
	suite.Equal("expected", rec.Header().Get(userHeaderKey))

	expected := map[string]string{
		"userID":        "me",
		"experimentKey": "key",
	}
	suite.assertValid(rec, expected)
}

func (suite *RouterTestSuite) TestSetForcedVariation() {
	req := httptest.NewRequest("PUT", "/users/me/experiments/exp_key/variations/var_key", nil)
	rec := httptest.NewRecorder()

	suite.mux.ServeHTTP(rec, req)

	suite.Equal("expected", rec.Header().Get(clientHeaderKey))
	suite.Equal("expected", rec.Header().Get(userHeaderKey))

	expected := map[string]string{
		"userID":        "me",
		"experimentKey": "exp_key",
		"variationKey":  "var_key",
	}
	suite.assertValid(rec, expected)
}

func (suite *RouterTestSuite) TestRemoveForcedVariation() {
	req := httptest.NewRequest("DELETE", "/users/me/experiments/exp_key/variations", nil)
	rec := httptest.NewRecorder()

	suite.mux.ServeHTTP(rec, req)

	suite.Equal("expected", rec.Header().Get(clientHeaderKey))
	suite.Equal("expected", rec.Header().Get(userHeaderKey))

	expected := map[string]string{
		"userID":        "me",
		"experimentKey": "exp_key",
	}
	suite.assertValid(rec, expected)
}

func (suite *RouterTestSuite) TestThrottleConfig() {
	req := httptest.NewRequest("GET", "/throttled", nil)

	wg1 := sync.WaitGroup{}
	wg1.Add(1)
	suite.mux.Get("/throttled", func(w http.ResponseWriter, r *http.Request) {
		wg1.Wait()
	})

	wg2 := sync.WaitGroup{}
	wg2.Add(1)
	go func() {
		wg2.Done()
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)
	}()
	wg2.Wait()

	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusServiceUnavailable, rec.Code)
	wg1.Done()
}

func (suite *RouterTestSuite) assertValid(rec *httptest.ResponseRecorder, expected map[string]string) {
	suite.Equal(http.StatusOK, rec.Code)

	var actual map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), expected, actual)
}

func TestRouter(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}
