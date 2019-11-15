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

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

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

type OptlyMiddlewareTestSuite struct {
	suite.Suite
	mw *CachedOptlyMiddleware
}

func (suite *OptlyMiddlewareTestSuite) SetupTest() {
	mockCache := new(MockCache)
	mockCache.On("GetClient", "ERROR").Return(new(optimizely.OptlyClient), fmt.Errorf("Error"))
	mockCache.On("GetClient", "EXPECTED").Return(&expectedClient, nil)
	suite.mw = &CachedOptlyMiddleware{mockCache}
}

func (suite *OptlyMiddlewareTestSuite) TestGetError() {
	handler := suite.mw.ClientCtx(ErrorHandler(suite))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add(OptlySDKHeader, "ERROR")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	suite.Equal(http.StatusInternalServerError, rec.Code)
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
	handler := ErrorHandler(suite)
	mux.With(suite.mw.UserCtx).Get("/{userID}/features", handler)

	req := httptest.NewRequest("GET", "//features?foo=true&bar=yes&baz=100", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

// ErrorHandler will panic if reached.
func ErrorHandler(suite *OptlyMiddlewareTestSuite) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		panic("test entered test handler, this should not happen")
	}
	return http.HandlerFunc(fn)
}

func AssertOptlyClientHandler(suite *OptlyMiddlewareTestSuite, expected *optimizely.OptlyClient) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		actual, err := GetOptlyClient(r)
		suite.NoError(err)
		suite.Same(expected, actual)

	}
	return http.HandlerFunc(fn)
}

func AssertOptlyContextHandler(suite *OptlyMiddlewareTestSuite, expected *optimizely.OptlyContext) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		actual, err := GetOptlyContext(r)
		suite.NoError(err)
		suite.Equal(expected, actual)

	}
	return http.HandlerFunc(fn)
}

func TestOptlyMiddleware(t *testing.T) {
	suite.Run(t, new(OptlyMiddlewareTestSuite))
}
