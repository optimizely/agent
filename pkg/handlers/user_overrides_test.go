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

// Package handlers //
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/optimizely/go-sdk/pkg/decision"

	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"

	"github.com/go-chi/chi"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/suite"
)

type UserOverridesTestSuite struct {
	suite.Suite
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

// Setup Mux
func (suite *UserOverridesTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    nil,
		ForcedVariations: testClient.ForcedVariations,
	}

	mux := chi.NewMux()
	userAPI := new(UserOverrideHandler)
	userMW := &UserMW{optlyClient}

	mux.Use(userMW.ClientCtx, userMW.UserCtx)
	mux.Put("/experiments/{experimentKey}/variations/{variationKey}", userAPI.SetForcedVariation)
	mux.Delete("/experiments/{experimentKey}/variations", userAPI.RemoveForcedVariation)

	suite.mux = mux
	suite.tc = testClient
}

func (suite *UserOverridesTestSuite) TestSetForcedVariation() {
	feature := entities.Feature{Key: "my_feat"}
	suite.tc.ProjectConfig.AddMultiVariationFeatureTest(feature, "variation_disabled", "variation_enabled")
	featureExp := suite.tc.ProjectConfig.FeatureMap["my_feat"].FeatureExperiments[0]

	req := httptest.NewRequest("PUT", "/experiments/"+featureExp.Key+"/variations/variation_enabled", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusCreated, rec.Code)

	key := decision.ExperimentOverrideKey{
		ExperimentKey: featureExp.Key,
		UserID:        "testUser",
	}

	actual, ok := suite.tc.ForcedVariations.GetVariation(key)
	suite.True(ok)
	suite.Equal("variation_enabled", actual)

	req = httptest.NewRequest("PUT", "/experiments/"+featureExp.Key+"/variations/variation_enabled", nil)
	rec = httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusNoContent, rec.Code)

	actual, ok = suite.tc.ForcedVariations.GetVariation(key)
	suite.True(ok)
	suite.Equal("variation_enabled", actual)
}

func (suite *UserOverridesTestSuite) TestSetForcedVariationEmptyExperimentKey() {
	req := httptest.NewRequest("PUT", "/experiments//variations/variation_enabled", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *UserOverridesTestSuite) TestRemoveForcedVariation() {
	feature := entities.Feature{Key: "my_feat"}
	suite.tc.ProjectConfig.AddMultiVariationFeatureTest(feature, "variation_disabled", "variation_enabled")
	featureExp := suite.tc.ProjectConfig.FeatureMap["my_feat"].FeatureExperiments[0]

	suite.tc.ForcedVariations.SetVariation(decision.ExperimentOverrideKey{
		ExperimentKey: featureExp.Key,
		UserID:        "testUser",
	}, "variation_enabled")

	req := httptest.NewRequest("DELETE", "/experiments/"+featureExp.Key+"/variations", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusNoContent, rec.Code)

	key := decision.ExperimentOverrideKey{
		ExperimentKey: "my_feat",
		UserID:        "testUser",
	}

	suite.Empty(suite.tc.ForcedVariations.GetVariation(key))
}

func (suite *UserOverridesTestSuite) TestRemoveForcedVariationEmptyExperimentKey() {
	req := httptest.NewRequest("DELETE", "/experiments//variations", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestUserOverrideTestSuite(t *testing.T) {
	suite.Run(t, new(UserOverridesTestSuite))
}

func TestUserOverrideMissingClientCtx(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req := httptest.NewRequest("POST", "/", nil)

	userOverrideHandler := new(UserOverrideHandler)
	handlers := []func(w http.ResponseWriter, r *http.Request){
		userOverrideHandler.SetForcedVariation,
		userOverrideHandler.RemoveForcedVariation,
	}

	for _, handler := range handlers {
		rec := httptest.NewRecorder()
		http.HandlerFunc(handler).ServeHTTP(rec, req)
		assertError(t, rec, "optlyClient not available", http.StatusInternalServerError)
	}
}

func TestUserOverrideMissingOptlyCtx(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req := httptest.NewRequest("POST", "/", nil)
	mw := new(UserMW)

	userOverrideHandler := new(UserOverrideHandler)
	handlers := []func(w http.ResponseWriter, r *http.Request){
		userOverrideHandler.SetForcedVariation,
		userOverrideHandler.RemoveForcedVariation,
	}

	for _, handler := range handlers {
		rec := httptest.NewRecorder()
		mw.ClientCtx(http.HandlerFunc(handler)).ServeHTTP(rec, req)
		assertError(t, rec, "optlyContext not available", http.StatusInternalServerError)
	}
}
