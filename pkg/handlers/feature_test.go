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

// Package handlers //
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"

	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/entities"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FeatureTestSuite struct {
	suite.Suite
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

type OptlyMWFeature struct {
	optlyClient *optimizely.OptlyClient
}

func (o *OptlyMWFeature) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, o.optlyClient)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (o *OptlyMWFeature) FeatureCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		featureKey := chi.URLParam(r, "featureKey")
		if featureKey == "one" {
			feature := config.OptimizelyFeature{Key: featureKey}
			ctx := context.WithValue(r.Context(), middleware.OptlyFeatureKey, &feature)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// Setup Mux
func (suite *FeatureTestSuite) SetupTest() {

	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    nil,
		ForcedVariations: testClient.ForcedVariations,
	}

	mux := chi.NewMux()
	featureAPI := new(FeatureHandler)
	optlyMW := &OptlyMWFeature{optlyClient}

	mux.Use(optlyMW.ClientCtx)
	mux.Get("/features", featureAPI.ListFeatures)
	mux.With(optlyMW.FeatureCtx).Get("/features/{featureKey}", featureAPI.GetFeature)

	suite.mux = mux
	suite.tc = testClient
}

func (suite *FeatureTestSuite) TestListFeatures() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeature(feature)

	// Add appropriate context
	req := httptest.NewRequest("GET", "/features", nil)
	rec := httptest.NewRecorder()

	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual []entities.Feature
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	suite.Equal(1, len(actual))
	suite.Equal(feature, actual[0])
}

func (suite *FeatureTestSuite) TestGetFeature() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeature(feature)

	req := httptest.NewRequest("GET", "/features/one", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual entities.Feature
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	suite.Equal(feature, actual)
}

func (suite *FeatureTestSuite) TestGetFeaturesMissingFeature() {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/features/feature-404", nil)
	suite.Nil(err)

	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusInternalServerError, rec.Code)
	// Unmarshal response
	var actual ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	suite.Equal(ErrorResponse{Error: "feature not available"}, actual)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFeatureTestSuite(t *testing.T) {
	suite.Run(t, new(FeatureTestSuite))
}

func TestListFeatureMissingClientCtx(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req := httptest.NewRequest("GET", "/", nil)
	featureHander := new(FeatureHandler)
	rec := httptest.NewRecorder()
	http.HandlerFunc(featureHander.ListFeatures).ServeHTTP(rec, req)

	// Unmarshal response
	var actual ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Equal(t, ErrorResponse{Error: "optlyClient not available"}, actual)
}
