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

	"github.com/go-chi/chi/v5"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
)

type ActivateTestSuite struct {
	suite.Suite
	oc   *optimizely.OptlyClient
	tc   *optimizelytest.TestClient
	body []byte
	mux  *chi.Mux
}

func (suite *ActivateTestSuite) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, suite.oc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (suite *ActivateTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    nil,
		ForcedVariations: testClient.ForcedVariations,
	}

	mux := chi.NewMux()
	mux.With(suite.ClientCtx).Post("/activate", Activate)

	ab := ActivateBody{
		UserID:         "testUser",
		UserAttributes: nil,
	}

	payload, err := json.Marshal(ab)
	suite.NoError(err)

	suite.body = payload
	suite.mux = mux
	suite.tc = testClient
	suite.oc = optlyClient
}

func (suite *ActivateTestSuite) TestInvalidPayload() {
	req := httptest.NewRequest("POST", "/activate", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.assertError(rec, `missing "userId" in request payload`, http.StatusBadRequest)
}

func (suite *ActivateTestSuite) TestGetFeatureWithFeatureTest() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureTest(feature)

	req := httptest.NewRequest("POST", "/activate?featureKey=one&disableTracking=true", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual []optimizely.Decision
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := optimizely.Decision{
		UserID:        "testUser",
		FeatureKey:    "one",
		Type:          "feature",
		Enabled:       true,
		ExperimentKey: "1",
		VariationKey:  "2",
	}

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
	suite.Equal(expected, actual[0])
}

func (suite *ActivateTestSuite) TestTrackFeatureWithFeatureRollout() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureRollout(feature)

	req := httptest.NewRequest("POST", "/activate?featureKey=one", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual []optimizely.Decision
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := optimizely.Decision{
		UserID:     "testUser",
		FeatureKey: "one",
		Enabled:    true,
		Type:       "feature",
	}

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
	suite.Equal(expected, actual[0])
}

func (suite *ActivateTestSuite) TestTrackFeatureWithFeatureTest() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureTest(feature)

	req := httptest.NewRequest("POST", "/activate?featureKey=one", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual []optimizely.Decision
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := optimizely.Decision{
		UserID:        "testUser",
		FeatureKey:    "one",
		Type:          "feature",
		Enabled:       true,
		ExperimentKey: "1",
		VariationKey:  "2",
	}
	suite.Equal(expected, actual[0])

	events := suite.tc.GetProcessedEvents()
	suite.Equal(1, len(events))

	impression := events[0]
	suite.Equal("campaign_activated", impression.Impression.Key)
	suite.Equal("testUser", impression.VisitorID)
}

func (suite *ActivateTestSuite) TestGetFeatureMissingFeature() {
	req := httptest.NewRequest("POST", "/activate?featureKey=feature-missing", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual []optimizely.Decision
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := optimizely.Decision{
		UserID:     "testUser",
		FeatureKey: "feature-missing",
		Error:      "featureKey not found",
	}

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
	suite.Equal(expected, actual[0])
}

func (suite *ActivateTestSuite) TestGetVariationMissingExperiment() {
	req := httptest.NewRequest("POST", "/activate?experimentKey=experiment-missing", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual []optimizely.Decision
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := optimizely.Decision{
		UserID:        "testUser",
		ExperimentKey: "experiment-missing",
		Error:         "experimentKey not found",
	}

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
	suite.Equal(expected, actual[0])
}

func (suite *ActivateTestSuite) TestActivateExperiment() {
	testVariation := suite.tc.ProjectConfig.CreateVariation("variation_a")
	suite.tc.AddExperiment("one", []entities.Variation{testVariation})

	req := httptest.NewRequest("POST", "/activate?experimentKey=one", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual []optimizely.Decision
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := optimizely.Decision{
		UserID:        "testUser",
		ExperimentKey: "one",
		VariationKey:  testVariation.Key,
		Type:          "experiment",
		Enabled:       true,
	}

	suite.Equal(1, len(suite.tc.GetProcessedEvents()))
	suite.Equal(expected, actual[0])
}

func (suite *ActivateTestSuite) TestActivateFeatures() {
	// 100% enabled rollout
	feature := entities.Feature{Key: "featureA"}
	suite.tc.AddFeatureRollout(feature)

	// 100% enabled feature test
	featureB := entities.Feature{Key: "featureB"}
	suite.tc.AddFeatureTest(featureB)

	// Feature test 100% enabled variation 100% with variation variable value
	variable := entities.Variable{DefaultValue: "default", ID: "123", Key: "strvar", Type: "string"}
	featureC := entities.Feature{Key: "featureC", VariableMap: map[string]entities.Variable{"strvar": variable}}
	suite.tc.AddFeatureTestWithCustomVariableValue(featureC, variable, "abc_notdef")

	expected := []optimizely.Decision{
		{
			UserID:     "testUser",
			Enabled:    true,
			FeatureKey: "featureA",
			Type:       "feature",
		},
		{
			UserID:        "testUser",
			Enabled:       true,
			FeatureKey:    "featureB",
			Type:          "feature",
			ExperimentKey: "5",
			VariationKey:  "6",
		},
		{
			UserID:     "testUser",
			Enabled:    true,
			FeatureKey: "featureC",
			Type:       "feature",
			Variables: map[string]interface{}{
				"strvar": "abc_notdef",
			},
			ExperimentKey: "12",
			VariationKey:  "13",
		},
	}

	// Toggle between tracking and no tracking.
	for _, flag := range []string{"true", "false"} {
		req := httptest.NewRequest("POST", "/activate?type=feature&disableTracking="+flag, bytes.NewBuffer(suite.body))
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)

		suite.Equal(http.StatusOK, rec.Code)

		// Unmarshal response
		var actual []optimizely.Decision
		err := json.Unmarshal(rec.Body.Bytes(), &actual)
		suite.NoError(err)
		suite.ElementsMatch(expected, actual)
	}

	// 2 for the 2 feature tests
	suite.Equal(2, len(suite.tc.GetProcessedEvents()))
}

func (suite *ActivateTestSuite) TestActivateExperiments() {
	testVariationA := suite.tc.ProjectConfig.CreateVariation("variation_a")
	suite.tc.AddExperiment("one", []entities.Variation{testVariationA})

	testVariationB := suite.tc.ProjectConfig.CreateVariation("variation_b")
	suite.tc.AddExperiment("two", []entities.Variation{testVariationB})

	testVariationC := suite.tc.ProjectConfig.CreateVariation("variation_c")
	suite.tc.AddExperiment("three", []entities.Variation{testVariationC})

	expected := []optimizely.Decision{
		{
			UserID:        "testUser",
			ExperimentKey: "one",
			VariationKey:  testVariationA.Key,
			Type:          "experiment",
			Enabled:       true,
		},
		{
			UserID:        "testUser",
			ExperimentKey: "two",
			VariationKey:  testVariationB.Key,
			Type:          "experiment",
			Enabled:       true,
		},
		{
			UserID:        "testUser",
			ExperimentKey: "three",
			VariationKey:  testVariationC.Key,
			Type:          "experiment",
			Enabled:       true,
		},
	}

	// Toggle between tracking and no tracking.
	for _, flag := range []string{"true", "false"} {
		req := httptest.NewRequest("POST", "/activate?type=experiment&disableTracking="+flag, bytes.NewBuffer(suite.body))
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)

		suite.Equal(http.StatusOK, rec.Code)

		// Unmarshal response
		var actual []optimizely.Decision
		err := json.Unmarshal(rec.Body.Bytes(), &actual)
		suite.NoError(err)

		suite.ElementsMatch(expected, actual)
	}

	suite.Equal(3, len(suite.tc.GetProcessedEvents()))
}

func (suite *ActivateTestSuite) TestEnabledFilter() {
	// 100% enabled rollout
	feature := entities.Feature{Key: "featureA"}
	suite.tc.AddFeatureRollout(feature)

	// 100% disabled rollout
	featureB := entities.Feature{Key: "featureB"}
	suite.tc.AddDisabledFeatureRollout(featureB)

	// Feature test 100% enabled variation 100% with variation variable value
	variable := entities.Variable{DefaultValue: "default", ID: "123", Key: "strvar", Type: "string"}
	featureC := entities.Feature{Key: "featureC", VariableMap: map[string]entities.Variable{"strvar": variable}}
	suite.tc.AddFeatureTestWithCustomVariableValue(featureC, variable, "abc_notdef")

	expected := []optimizely.Decision{
		{
			UserID:     "testUser",
			Enabled:    true,
			FeatureKey: "featureA",
			Type:       "feature",
		},
		{
			UserID:     "testUser",
			Enabled:    true,
			FeatureKey: "featureC",
			Type:       "feature",
			Variables: map[string]interface{}{
				"strvar": "abc_notdef",
			},
			ExperimentKey: "13",
			VariationKey:  "14",
		},
		{
			UserID:     "testUser",
			Enabled:    false,
			FeatureKey: "featureB",
			Type:       "feature",
		},
	}

	scenarios := []struct {
		param    string
		expected []optimizely.Decision
	}{
		{
			"",
			expected,
		},
		{
			"&enabled=true",
			expected[0:2],
		},
		{
			"&enabled=false",
			expected[2:],
		},
	}

	for _, scenario := range scenarios {
		req := httptest.NewRequest("POST", "/activate?type=feature"+scenario.param, bytes.NewBuffer(suite.body))
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)

		suite.Equal(http.StatusOK, rec.Code)

		// Unmarshal response
		var actual []optimizely.Decision
		err := json.Unmarshal(rec.Body.Bytes(), &actual)
		suite.NoError(err)
		suite.ElementsMatch(scenario.expected, actual)
	}
}

func (suite *ActivateTestSuite) TestInvalidFilter() {
	req := httptest.NewRequest("POST", "/activate?type=invalid", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.assertError(rec, `type "invalid" not supported`, http.StatusBadRequest)
}

func (suite *ActivateTestSuite) assertError(rec *httptest.ResponseRecorder, msg string, code int) {
	assertError(suite.T(), rec, msg, code)
}

func TestActivateTestSuite(t *testing.T) {
	suite.Run(t, new(ActivateTestSuite))
}

func TestGetUserContext(t *testing.T) {
	dc := ActivateBody{
		UserID: "test name",
		UserAttributes: map[string]interface{}{
			"str":    "val",
			"bool":   true,
			"double": 1.01,
			"int":    float64(10), // might be can be problematic
		},
	}

	jsonEntity, err := json.Marshal(dc)
	assert.NoError(t, err)
	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonEntity))

	actual, err := getUserContext(req)
	assert.NoError(t, err)

	expected := entities.UserContext{
		ID:         dc.UserID,
		Attributes: dc.UserAttributes,
	}

	assert.Equal(t, expected, actual)
}

func TestActivateMissingOptlyCtx(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()
	http.HandlerFunc(Activate).ServeHTTP(rec, req)
	assertError(t, rec, "optlyClient not available", http.StatusInternalServerError)
}
