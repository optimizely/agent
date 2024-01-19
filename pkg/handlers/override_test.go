/****************************************************************************
 * Copyright 2020,2023, Optimizely, Inc. and contributors                   *
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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"

	"github.com/optimizely/go-sdk/v2/pkg/decision"
	"github.com/optimizely/go-sdk/v2/pkg/entities"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"
)

type OverrideTestSuite struct {
	suite.Suite
	oc            *optimizely.OptlyClient
	tc            *optimizelytest.TestClient
	mux           *chi.Mux
	experimentKey string
}

func (suite *OverrideTestSuite) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, suite.oc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Setup Mux
func (suite *OverrideTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    nil,
		ForcedVariations: testClient.ForcedVariations,
	}

	mux := chi.NewMux()
	mux.With(suite.ClientCtx).Post("/override", Override)

	feature := entities.Feature{Key: "my_feat"}
	testClient.ProjectConfig.AddMultiVariationFeatureTest(feature, "variation_disabled", "variation_enabled")
	featureExp := testClient.ProjectConfig.FeatureMap["my_feat"].FeatureExperiments[0]

	testClient.AddExperimentWithVariations("valid", "valid")

	suite.experimentKey = featureExp.Key
	suite.mux = mux
	suite.tc = testClient
	suite.oc = optlyClient
}

func (suite *OverrideTestSuite) TestValidOverrides() {
	scenarios := []struct {
		variationKey string
		previousKey  string
		messages     []string
	}{
		{
			variationKey: "variation_enabled",
			previousKey:  "",
			messages:     nil,
		},
		{
			variationKey: "variation_enabled",
			previousKey:  "variation_enabled",
			messages:     []string{"updating previous override"},
		},
		{
			variationKey: "",
			previousKey:  "variation_enabled",
			messages:     []string{"removing previous override"},
		},
		{
			variationKey: "",
			previousKey:  "",
			messages:     []string{"no pre-existing override"},
		},
	}

	for _, scenario := range scenarios {
		ob := OverrideBody{
			UserID:        "testUser",
			ExperimentKey: suite.experimentKey,
			VariationKey:  scenario.variationKey,
		}

		body, err := json.Marshal(ob)
		suite.NoError(err)

		req := httptest.NewRequest("POST", "/override", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)
		suite.Equal(http.StatusOK, rec.Code)

		// Unmarshal response
		var actual optimizely.Override
		suite.NoError(json.Unmarshal(rec.Body.Bytes(), &actual))

		expected := optimizely.Override{
			UserID:           "testUser",
			ExperimentKey:    suite.experimentKey,
			VariationKey:     scenario.variationKey,
			PrevVariationKey: scenario.previousKey,
			Messages:         scenario.messages,
		}

		suite.Equal(expected, actual)

		key := decision.ExperimentOverrideKey{
			ExperimentKey: suite.experimentKey,
			UserID:        "testUser",
		}

		actVar, _ := suite.tc.ForcedVariations.GetVariation(key)
		suite.Equal(scenario.variationKey, actVar)
	}
}

func (suite *OverrideTestSuite) TestInvalidOverrides() {
	invalid := []map[string]interface{}{
		{
			"userID":        "",
			"experimentKey": "valid",
			"variationKey":  "valid",
		},
		{
			"userID":        "valid",
			"experimentKey": "",
			"variationKey":  "valid",
		},
		{
			"userId": true,
		},
	}

	for _, payload := range invalid {
		body, err := json.Marshal(payload)
		suite.NoError(err)

		req := httptest.NewRequest("POST", "/override", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)
		suite.Equal(http.StatusBadRequest, rec.Code)
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestOverrideTestSuite(t *testing.T) {
	suite.Run(t, new(OverrideTestSuite))
}

func TestOverrideMissingClientCtx(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()
	http.HandlerFunc(Override).ServeHTTP(rec, req)
	assertError(t, rec, "optlyClient not available", http.StatusInternalServerError)
}
