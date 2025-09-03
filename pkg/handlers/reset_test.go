/****************************************************************************
 * Copyright 2025, Optimizely, Inc. and contributors                        *
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

// Package handlers //
package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) GetClient(key string) (*optimizely.OptlyClient, error) {
	args := m.Called(key)
	return args.Get(0).(*optimizely.OptlyClient), args.Error(1)
}

func (m *MockCache) UpdateConfigs(_ string) {
}

func (m *MockCache) SetUserProfileService(sdkKey, userProfileService string) {
	m.Called(sdkKey, userProfileService)
}

func (m *MockCache) SetODPCache(sdkKey, odpCache string) {
	m.Called(sdkKey, odpCache)
}

func (m *MockCache) ResetClient(sdkKey string) {
	m.Called(sdkKey)
}

type ResetTestSuite struct {
	suite.Suite
	oc    *optimizely.OptlyClient
	tc    *optimizelytest.TestClient
	mux   *chi.Mux
	cache *MockCache
}

func (suite *ResetTestSuite) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, suite.oc)
		ctx = context.WithValue(ctx, middleware.OptlyCacheKey, suite.cache)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (suite *ResetTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	suite.tc = testClient
	suite.oc = &optimizely.OptlyClient{OptimizelyClient: testClient.OptimizelyClient}

	mockCache := new(MockCache)
	mockCache.On("ResetClient", "test-sdk-key").Return()
	suite.cache = mockCache

	mux := chi.NewMux()
	mux.Use(suite.ClientCtx)
	mux.Post("/reset", ResetClient)
	suite.mux = mux
}

func (suite *ResetTestSuite) TestResetClient() {
	req := httptest.NewRequest("POST", "/reset", nil)
	req.Header.Set("X-Optimizely-SDK-Key", "test-sdk-key")
	recorder := httptest.NewRecorder()

	suite.mux.ServeHTTP(recorder, req)

	suite.Equal(http.StatusOK, recorder.Code)
	suite.Contains(recorder.Header().Get("content-type"), "application/json")
	suite.Contains(recorder.Body.String(), `"result":true`)

	// Verify ResetClient was called with correct SDK key
	suite.cache.AssertCalled(suite.T(), "ResetClient", "test-sdk-key")
}

func (suite *ResetTestSuite) TestResetClientMissingSDKKey() {
	req := httptest.NewRequest("POST", "/reset", nil)
	recorder := httptest.NewRecorder()

	suite.mux.ServeHTTP(recorder, req)

	suite.Equal(http.StatusBadRequest, recorder.Code)
	suite.Contains(recorder.Body.String(), "SDK key required for reset")
}

func (suite *ResetTestSuite) TestResetClientCacheNotAvailable() {
	// Create a context without cache
	req := httptest.NewRequest("POST", "/reset", nil)
	req.Header.Set("X-Optimizely-SDK-Key", "test-sdk-key")
	recorder := httptest.NewRecorder()

	// Use middleware that doesn't include cache
	mux := chi.NewMux()
	mux.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, suite.oc)
			// Note: no cache in context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	mux.Post("/reset", ResetClient)

	mux.ServeHTTP(recorder, req)

	suite.Equal(http.StatusInternalServerError, recorder.Code)
	suite.Contains(recorder.Body.String(), "cache not available")
}

func TestResetTestSuite(t *testing.T) {
	suite.Run(t, new(ResetTestSuite))
}
