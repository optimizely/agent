/****************************************************************************
 * Copyright 2021, Optimizely, Inc. and contributors                        *
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

	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/entities"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DecideTestSuite struct {
	suite.Suite
	oc        *optimizely.OptlyClient
	tc        *optimizelytest.TestClient
	body      []byte
	bodyEvent []byte
	mux       *chi.Mux
}

func (suite *DecideTestSuite) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, suite.oc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (suite *DecideTestSuite) assertError(rec *httptest.ResponseRecorder, msg string, code int) {
	assertError(suite.T(), rec, msg, code)
}

func (suite *DecideTestSuite) TestInvalidPayload() {
	req := httptest.NewRequest("POST", "/decide", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.assertError(rec, `missing "userId" in request payload`, http.StatusBadRequest)
}

func (suite *DecideTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    nil,
		ForcedVariations: testClient.ForcedVariations,
	}

	mux := chi.NewMux()
	mux.With(suite.ClientCtx).Post("/decide", Decide)

	db := DecideBody{
		UserID:         "testUser",
		UserAttributes: nil,
		DecideOptions:  []string{"DISABLE_DECISION_EVENT"},
	}

	payload, err := json.Marshal(db)
	suite.NoError(err)

	suite.body = payload

	db = DecideBody{
		UserID:         "testUser",
		UserAttributes: nil,
		DecideOptions:  []string{},
	}
	payload, err = json.Marshal(db)
	suite.NoError(err)

	suite.bodyEvent = payload

	suite.mux = mux
	suite.tc = testClient
	suite.oc = optlyClient
}

func (suite *DecideTestSuite) TestDecideWithFeatureTest() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureTest(feature)

	req := httptest.NewRequest("POST", "/decide?keys=one", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual client.OptimizelyDecision
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := client.OptimizelyDecision{
		UserContext:  client.OptimizelyUserContext{UserID: "testUser", Attributes: map[string]interface{}{}},
		FlagKey:      "one",
		RuleKey:      "1",
		Enabled:      true,
		VariationKey: "2",
		Reasons:      []string{},
	}

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
	suite.Equal(expected, actual)
}

func (suite *DecideTestSuite) TestTrackWithFeatureRollout() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureRollout(feature)

	req := httptest.NewRequest("POST", "/decide?keys=one", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual client.OptimizelyDecision
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := client.OptimizelyDecision{
		UserContext:  client.OptimizelyUserContext{UserID: "testUser", Attributes: map[string]interface{}{}},
		FlagKey:      "one",
		RuleKey:      "1",
		Enabled:      true,
		VariationKey: "3",
		Reasons:      []string{},
	}

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
	suite.Equal(expected, actual)
}

func (suite *DecideTestSuite) TestTrackWithFeatureTest() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureTest(feature)

	req := httptest.NewRequest("POST", "/decide?keys=one", bytes.NewBuffer(suite.bodyEvent))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual client.OptimizelyDecision
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := client.OptimizelyDecision{
		UserContext:  client.OptimizelyUserContext{UserID: "testUser", Attributes: map[string]interface{}{}},
		FlagKey:      "one",
		RuleKey:      "1",
		Enabled:      true,
		VariationKey: "2",
		Reasons:      []string{},
	}
	suite.Equal(expected, actual)

	events := suite.tc.GetProcessedEvents()
	suite.Equal(1, len(events))

	impression := events[0]
	suite.Equal("campaign_activated", impression.Impression.Key)
	suite.Equal("testUser", impression.VisitorID)
}

func (suite *DecideTestSuite) TestDecideMissingFlag() {
	req := httptest.NewRequest("POST", "/decide?keys=feature-missing", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual client.OptimizelyDecision
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := client.OptimizelyDecision{
		UserContext:  client.OptimizelyUserContext{UserID: "testUser", Attributes: map[string]interface{}{}},
		FlagKey:      "feature-missing",
		RuleKey:      "",
		Enabled:      false,
		VariationKey: "",
		Reasons:      []string{"No flag was found for key \"feature-missing\"."},
	}

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
	suite.Equal(expected, actual)
}

func (suite *DecideTestSuite) TestDecideMultipleFlags() {
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

	expected := []DecideOut{
		{
			OptimizelyDecision: client.OptimizelyDecision{
				UserContext:  client.OptimizelyUserContext{UserID: "testUser", Attributes: map[string]interface{}{}},
				FlagKey:      "featureA",
				RuleKey:      "1",
				Enabled:      true,
				VariationKey: "3",
				Reasons:      []string{},
			},
			Variables: nil,
		},
		{
			OptimizelyDecision: client.OptimizelyDecision{
				UserContext:  client.OptimizelyUserContext{UserID: "testUser", Attributes: map[string]interface{}{}},
				FlagKey:      "featureB",
				RuleKey:      "5",
				Enabled:      true,
				VariationKey: "6",
				Reasons:      []string{},
			},
			Variables: nil,
		},
	}

	// Toggle between tracking and no tracking.
	for _, body := range [][]byte{suite.body, suite.bodyEvent} {
		req := httptest.NewRequest("POST", "/decide?keys=featureA&keys=featureB", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)

		suite.Equal(http.StatusOK, rec.Code)

		// Unmarshal response
		var actual []DecideOut
		err := json.Unmarshal(rec.Body.Bytes(), &actual)

		suite.NoError(err)
		suite.ElementsMatch(expected, actual)
	}

	// 1 for the feature test
	suite.Equal(1, len(suite.tc.GetProcessedEvents()))
}

func (suite *DecideTestSuite) TestDecideAllFlags() {
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

	expected := []DecideOut{
		{
			OptimizelyDecision: client.OptimizelyDecision{
				UserContext:  client.OptimizelyUserContext{UserID: "testUser", Attributes: map[string]interface{}{}},
				FlagKey:      "featureA",
				RuleKey:      "1",
				Enabled:      true,
				VariationKey: "3",
				Reasons:      []string{},
			},
			Variables: nil,
		},
		{
			OptimizelyDecision: client.OptimizelyDecision{
				UserContext:  client.OptimizelyUserContext{UserID: "testUser", Attributes: map[string]interface{}{}},
				FlagKey:      "featureB",
				RuleKey:      "5",
				Enabled:      true,
				VariationKey: "6",
				Reasons:      []string{},
			},
			Variables: nil,
		},
		{
			OptimizelyDecision: client.OptimizelyDecision{
				UserContext:  client.OptimizelyUserContext{UserID: "testUser", Attributes: map[string]interface{}{}},
				FlagKey:      "featureC",
				RuleKey:      "12",
				Enabled:      true,
				VariationKey: "13",
				Reasons:      []string{},
			},
			Variables: map[string]interface{}{"strvar": "abc_notdef"},
		},
	}

	// Toggle between tracking and no tracking.
	for _, body := range [][]byte{suite.body, suite.bodyEvent} {
		req := httptest.NewRequest("POST", "/decide", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		suite.mux.ServeHTTP(rec, req)

		suite.Equal(http.StatusOK, rec.Code)

		// Unmarshal response
		var actual []DecideOut
		err := json.Unmarshal(rec.Body.Bytes(), &actual)

		suite.NoError(err)
		suite.ElementsMatch(expected, actual)
	}

	// 2 for the 2 feature tests
	suite.Equal(2, len(suite.tc.GetProcessedEvents()))
}

func TestDecideTestSuite(t *testing.T) {
	suite.Run(t, new(DecideTestSuite))
}

func TestGetDecideUserContext(t *testing.T) {
	dc := DecideBody{
		UserID: "test name",
		UserAttributes: map[string]interface{}{
			"str":    "val",
			"bool":   true,
			"double": 1.01,
			"int":    float64(10), // might be can be problematic
		},
		DecideOptions: []string{"one", "two"},
	}

	jsonEntity, err := json.Marshal(dc)
	assert.NoError(t, err)
	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonEntity))

	actual, err := getUserContextWithOptions(req)
	assert.NoError(t, err)

	expected := DecideBody{
		UserID:         dc.UserID,
		UserAttributes: dc.UserAttributes,
		DecideOptions:  []string{"one", "two"},
	}

	assert.Equal(t, expected, actual)
}

func TestTranslateOptions(t *testing.T) {
	options := []string{"DISABLE_DECISION_EVENT", "ENABLED_FLAGS_ONLY", "IGNORE_USER_PROFILE_SERVICE",
		"EXCLUDE_VARIABLES", "INCLUDE_REASONS"}

	decideOptions, err := decide.TranslateOptions(options)
	expected := []decide.OptimizelyDecideOptions{decide.DisableDecisionEvent, decide.EnabledFlagsOnly, decide.IgnoreUserProfileService,
		decide.ExcludeVariables, decide.IncludeReasons}
	assert.NoError(t, err)
	assert.Equal(t, expected, decideOptions)

	options = append(options, "invalid")

	decideOptions, err = decide.TranslateOptions(options)
	assert.Error(t, err)
	assert.Equal(t, "invalid option: invalid", err.Error())
	assert.Equal(t, []decide.OptimizelyDecideOptions{}, decideOptions)
}
