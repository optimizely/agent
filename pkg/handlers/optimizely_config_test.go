/****************************************************************************
 * Copyright 2024 Optimizely, Inc. and contributors                         *
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

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
	"github.com/optimizely/go-sdk/pkg/config"
)

type OptimizelyConfigTestSuite struct {
	suite.Suite
	oc  *optimizely.OptlyClient
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

func (suite *OptimizelyConfigTestSuite) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, suite.oc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Setup Mux
func (suite *OptimizelyConfigTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    nil,
		ForcedVariations: testClient.ForcedVariations,
	}

	mux := chi.NewMux()
	mux.With(suite.ClientCtx).Get("/config", OptimizelyConfig)

	suite.oc = optlyClient
	suite.tc = testClient
	suite.mux = mux
}

func (suite *OptimizelyConfigTestSuite) TestConfig() {
	req := httptest.NewRequest("GET", "/config", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual config.OptimizelyConfig
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	suite.Equal(*suite.oc.GetOptimizelyConfig(), actual)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestOptimizelyConfigTestSuite(t *testing.T) {
	suite.Run(t, new(OptimizelyConfigTestSuite))
}

func TestOptimizelyConfigMissingOptlyCtx(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()
	http.HandlerFunc(OptimizelyConfig).ServeHTTP(rec, req)
	assertError(t, rec, "optlyClient not available", http.StatusInternalServerError)
}
