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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/go-sdk/pkg/entities"

	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
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

func (m MockCache) UpdateConfigs(_ string) {
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
	mockCache.On("GetClient", "INVALID").Return(new(optimizely.OptlyClient), optimizely.ErrValidationFailure)
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
	assertError(suite.T(), rec, "failed to instantiate Optimizely for SDK Key: ERROR", http.StatusInternalServerError)
}

func (suite *OptlyMiddlewareTestSuite) TestGetForbidden() {
	handler := suite.mw.ClientCtx(ErrorHandler())
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add(OptlySDKHeader, "403")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assertError(suite.T(), rec, "403 forbidden", http.StatusForbidden)
}

func (suite *OptlyMiddlewareTestSuite) TestGetInvalid() {
	handler := suite.mw.ClientCtx(ErrorHandler())
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add(OptlySDKHeader, "INVALID")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assertError(suite.T(), rec, "sdkKey failed validation", http.StatusBadRequest)
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

func TestOptlyMiddleware(t *testing.T) {
	suite.Run(t, new(OptlyMiddlewareTestSuite))
}
