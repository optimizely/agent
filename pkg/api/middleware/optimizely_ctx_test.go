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

	"github.com/stretchr/testify/suite"
	"github.com/stretchr/testify/mock"

	"github.com/optimizely/sidedoor/pkg/optimizely"
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

func (m *MockCache) GetDefaultClient() (*optimizely.OptlyClient, error) {
	return &defaultClient, nil
}

type OptlyMiddlewareTestSuite struct {
	suite.Suite
	optlyCtx *OptlyContext
}

func (suite *OptlyMiddlewareTestSuite) SetupTest() {
	mockCache := new(MockCache)
	mockCache.On("GetClient", "ERROR").Return(new(optimizely.OptlyClient), fmt.Errorf("Error"))
	mockCache.On("GetClient", "EXPECTED").Return(&expectedClient, nil)
	suite.optlyCtx = &OptlyContext{mockCache}
}

func (suite *OptlyMiddlewareTestSuite) TestGetError() {
	handler := suite.optlyCtx.ClientCtx(ErrorHandler(suite))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add(OptlySDKHeader, "ERROR")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	suite.Equal(http.StatusInternalServerError, rec.Code)
}

func (suite *OptlyMiddlewareTestSuite) TestGetDefault() {
	handler := suite.optlyCtx.ClientCtx(AssertHandler(suite, &defaultClient))
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *OptlyMiddlewareTestSuite) TestGetExpected() {
	handler := suite.optlyCtx.ClientCtx(AssertHandler(suite, &expectedClient))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add(OptlySDKHeader, "EXPECTED")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
}

func ErrorHandler(suite *OptlyMiddlewareTestSuite) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		panic("test entered test handler, this should not happen")
	}
	return http.HandlerFunc(fn)
}

func AssertHandler(suite *OptlyMiddlewareTestSuite, expected *optimizely.OptlyClient) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		actual, ok := r.Context().Value(OptlyClientKey).(*optimizely.OptlyClient)
		suite.True(ok)
		suite.Same(expected, actual)

	}
	return http.HandlerFunc(fn)
}

func TestOptlyMiddleware(t *testing.T) {
	suite.Run(t, new(OptlyMiddlewareTestSuite))
}
