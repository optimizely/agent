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

// Package middlewre //
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/entities"
)

var defaultClient = optimizely.OptlyClient{}
var expectedClient = optimizely.OptlyClient{}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) GetClient(key string) (*optimizely.OptlyClient, error) {
	args := m.Called(key)
	return args.Get(0).(*optimizely.OptlyClient), args.Error(1)
}

type OptlyMiddlewareTestSuite struct {
	suite.Suite
	mw *CachedOptlyMiddleware
	tc *optimizelytest.TestClient
}

func (suite *OptlyMiddlewareTestSuite) SetupTest() {
	mockCache := new(MockCache)
	mockCache.On("GetClient", "ERROR").Return(new(optimizely.OptlyClient), fmt.Errorf("error"))
	mockCache.On("GetClient", "403").Return(new(optimizely.OptlyClient), fmt.Errorf("403 forbidden"))
	mockCache.On("GetClient", "EXPECTED").Return(&expectedClient, nil)
	suite.mw = &CachedOptlyMiddleware{mockCache}

	suite.tc = optimizelytest.NewClient()
	suite.tc.AddFeature(entities.Feature{
		ID:  "1",
		Key: "one",
	})
	suite.tc.AddExperiment("expOne", []entities.Variation{{
		ID:  "9999",
		Key: "variation_1",
	}})
	clientWithConfig := optimizely.OptlyClient{
		OptimizelyClient: suite.tc.OptimizelyClient,
	}
	mockCache.On("GetClient", "WITH_TEST_CLIENT").Return(&clientWithConfig, nil)
}

func (suite *OptlyMiddlewareTestSuite) TestGetError() {
	handler := suite.mw.ClientCtx(ErrorHandler())
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add(OptlySDKHeader, "ERROR")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	suite.Equal(http.StatusInternalServerError, rec.Code)
}

func (suite *OptlyMiddlewareTestSuite) TestGetInvalid() {
	handler := suite.mw.ClientCtx(ErrorHandler())
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add(OptlySDKHeader, "403")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	suite.Equal(http.StatusForbidden, rec.Code)
}

func (suite *OptlyMiddlewareTestSuite) TestGetClientMissingHeader() {
	handler := suite.mw.ClientCtx(AssertOptlyClientHandler(suite, &defaultClient))
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *OptlyMiddlewareTestSuite) TestGetClient() {
	handler := suite.mw.ClientCtx(AssertOptlyClientHandler(suite, &expectedClient))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add(OptlySDKHeader, "EXPECTED")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *OptlyMiddlewareTestSuite) TestGetUserContext() {
	attributes := map[string]interface{}{
		"foo": "true",
		"bar": "yes",
		"baz": "100",
	}
	expected := optimizely.NewContext("test", attributes)

	mux := chi.NewMux()
	handler := AssertOptlyContextHandler(suite, expected)
	mux.With(suite.mw.UserCtx).Get("/{userID}", handler)

	req := httptest.NewRequest("GET", "/test?foo=true&bar=yes&baz=100", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *OptlyMiddlewareTestSuite) TestGetUserContextError() {
	mux := chi.NewMux()
	handler := ErrorHandler()
	mux.With(suite.mw.UserCtx).Get("/{userID}/features", handler)

	req := httptest.NewRequest("GET", "//features?foo=true&bar=yes&baz=100", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *OptlyMiddlewareTestSuite) TestFeatureCtxFeatureFound() {
	mux := chi.NewMux()
	handler := func(w http.ResponseWriter, r *http.Request) {
		actual, ok := r.Context().Value(OptlyFeatureKey).(*config.OptimizelyFeature)
		expected := &config.OptimizelyFeature{
			ExperimentsMap: make(map[string]config.OptimizelyExperiment),
			ID:             "1",
			Key:            "one",
			VariablesMap:   make(map[string]config.OptimizelyVariable),
		}
		suite.True(ok)
		suite.Equal(expected, actual)
	}
	mux.With(suite.mw.FeatureCtx).Get("/features/{featureKey}", handler)
	req := httptest.NewRequest("GET", "/features/one", nil)
	req = req.WithContext(context.WithValue(req.Context(), OptlyClientKey, &optimizely.OptlyClient{
		OptimizelyClient: suite.tc.OptimizelyClient,
	}))
	req.Header.Add(OptlySDKHeader, "WITH_TEST_CLIENT")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *OptlyMiddlewareTestSuite) TestFeatureCtxFeatureNotFound() {
	mux := chi.NewMux()
	handler := func(w http.ResponseWriter, r *http.Request) {
		suite.Fail("FeatureCtx should have returned 404 response without calling handler")
	}
	mux.With(suite.mw.FeatureCtx).Get("/features/{featureKey}", handler)
	req := httptest.NewRequest("GET", "/features/two", nil)
	req = req.WithContext(context.WithValue(req.Context(), OptlyClientKey, &optimizely.OptlyClient{
		OptimizelyClient: suite.tc.OptimizelyClient,
	}))
	req.Header.Add(OptlySDKHeader, "WITH_TEST_CLIENT")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusNotFound, rec.Code)
}

func (suite *OptlyMiddlewareTestSuite) TestFeatureCtxNoURLParam() {
	mux := chi.NewMux()
	handler := func(w http.ResponseWriter, r *http.Request) {
		suite.Fail("FeatureCtx should have returned 400 response without calling handler")
	}
	mux.With(suite.mw.FeatureCtx).Get("/features/{featureKey}/", handler)
	req := httptest.NewRequest("GET", "/features//", nil)
	req = req.WithContext(context.WithValue(req.Context(), OptlyClientKey, &optimizely.OptlyClient{
		OptimizelyClient: suite.tc.OptimizelyClient,
	}))
	req.Header.Add(OptlySDKHeader, "WITH_TEST_CLIENT")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *OptlyMiddlewareTestSuite) TestExperimentCtxExperimentFound() {
	mux := chi.NewMux()
	handler := func(w http.ResponseWriter, r *http.Request) {
		actual, ok := r.Context().Value(OptlyExperimentKey).(*config.OptimizelyExperiment)
		expected := &config.OptimizelyExperiment{
			ID:  suite.tc.ProjectConfig.ExperimentKeyToIDMap["expOne"],
			Key: "expOne",
			VariationsMap: map[string]config.OptimizelyVariation{
				"variation_1": {
					ID:           "9999",
					Key:          "variation_1",
					VariablesMap: map[string]config.OptimizelyVariable{},
				},
			},
		}
		suite.True(ok)
		suite.Equal(expected, actual)
	}
	mux.With(suite.mw.ExperimentCtx).Get("/experiments/{experimentKey}", handler)
	req := httptest.NewRequest("GET", "/experiments/expOne", nil)
	req = req.WithContext(context.WithValue(req.Context(), OptlyClientKey, &optimizely.OptlyClient{
		OptimizelyClient: suite.tc.OptimizelyClient,
	}))
	req.Header.Add(OptlySDKHeader, "WITH_TEST_CLIENT")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *OptlyMiddlewareTestSuite) TestExperimentCtxExperimentNotFound() {
	mux := chi.NewMux()
	handler := func(w http.ResponseWriter, r *http.Request) {
		suite.Fail("ExperimentCtx should have returned 404 response without calling handler")
	}
	mux.With(suite.mw.ExperimentCtx).Get("/experiments/{experimentKey}", handler)
	req := httptest.NewRequest("GET", "/experiments/expTwo", nil)
	req = req.WithContext(context.WithValue(req.Context(), OptlyClientKey, &optimizely.OptlyClient{
		OptimizelyClient: suite.tc.OptimizelyClient,
	}))
	req.Header.Add(OptlySDKHeader, "WITH_TEST_CLIENT")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusNotFound, rec.Code)
}

func (suite *OptlyMiddlewareTestSuite) TestExperimentCtxNoURLParam() {
	mux := chi.NewMux()
	handler := func(w http.ResponseWriter, r *http.Request) {
		suite.Fail("ExperimentCtx should have returned 400 response without calling handler")
	}
	mux.With(suite.mw.ExperimentCtx).Get("/experiments/{experimentKey}/", handler)
	req := httptest.NewRequest("GET", "/experiments//", nil)
	req = req.WithContext(context.WithValue(req.Context(), OptlyClientKey, &optimizely.OptlyClient{
		OptimizelyClient: suite.tc.OptimizelyClient,
	}))
	req.Header.Add(OptlySDKHeader, "WITH_TEST_CLIENT")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

// ErrorHandler will panic if reached.
func ErrorHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		panic("test entered test handler, this should not happen")
	}
}

func AssertOptlyClientHandler(suite *OptlyMiddlewareTestSuite, expected *optimizely.OptlyClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actual, err := GetOptlyClient(r)
		suite.NoError(err)
		suite.Same(expected, actual)

	}
}

func AssertOptlyContextHandler(suite *OptlyMiddlewareTestSuite, expected *optimizely.OptlyContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actual, err := GetOptlyContext(r)
		suite.NoError(err)
		suite.Equal(expected, actual)

	}
}

func TestOptlyMiddleware(t *testing.T) {
	suite.Run(t, new(OptlyMiddlewareTestSuite))
}
