/****************************************************************************
 * Copyright 2021,2023, Optimizely, Inc. and contributors                   *
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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/odp/segment"
)

type DecideTestSuite struct {
	suite.Suite
	oc        *optimizely.OptlyClient
	tc        *optimizelytest.TestClient
	body      []byte
	bodyEvent []byte
	mux       *chi.Mux
}

type TestDecideBody struct {
	UserID               string                 `json:"userId"`
	UserAttributes       map[string]interface{} `json:"userAttributes"`
	DecideOptions        []string               `json:"decideOptions"`
	ForcedDecisions      []ForcedDecision       `json:"forcedDecisions,omitempty"`
	FetchSegments        bool                   `json:"fetchSegments"`
	FetchSegmentsOptions json.RawMessage        `json:"fetchSegmentsOptions,omitempty"`
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

func (suite *DecideTestSuite) TestInvalidForcedDecisions() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureTest(feature)

	// Adding Forced Decision
	suite.tc.AddFlagVariation(feature, entities.Variation{Key: "4", FeatureEnabled: true})
	db := DecideBody{
		UserID:         "testUser",
		UserAttributes: nil,
		DecideOptions:  []string{"DISABLE_DECISION_EVENT"},
		ForcedDecisions: []ForcedDecision{
			{
				VariationKey: "2",
			},
			{
				RuleKey:      "1",
				VariationKey: "3",
			},
			{
				FlagKey: "one",
				RuleKey: "1",
			},
		},
	}

	payload, err := json.Marshal(db)
	suite.NoError(err)

	suite.body = payload

	req := httptest.NewRequest("POST", "/decide?keys=one", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual client.OptimizelyDecision
	err = json.Unmarshal(rec.Body.Bytes(), &actual)
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

func (suite *DecideTestSuite) TestForcedDecisionWithFeatureTest() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureTest(feature)

	// Adding Forced Decision
	suite.tc.AddFlagVariation(feature, entities.Variation{Key: "4", FeatureEnabled: true})
	db := DecideBody{
		UserID:         "testUser",
		UserAttributes: nil,
		DecideOptions:  []string{"DISABLE_DECISION_EVENT"},
		ForcedDecisions: []ForcedDecision{
			{
				FlagKey:      "two",
				RuleKey:      "1",
				VariationKey: "2",
			},
			{
				FlagKey:      "one",
				RuleKey:      "1",
				VariationKey: "3",
			},
			{
				// Checking for last forced-decision to be considered if 2 similar rule and flagkeys's are given
				FlagKey:      "one",
				RuleKey:      "1",
				VariationKey: "4",
			},
		},
	}

	payload, err := json.Marshal(db)
	suite.NoError(err)

	suite.body = payload

	req := httptest.NewRequest("POST", "/decide?keys=one", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual client.OptimizelyDecision
	err = json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := client.OptimizelyDecision{
		UserContext:  client.OptimizelyUserContext{UserID: "testUser", Attributes: map[string]interface{}{}},
		FlagKey:      "one",
		RuleKey:      "1",
		Enabled:      true,
		VariationKey: "4",
		Reasons:      []string{},
	}

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
	suite.Equal(expected, actual)
}

func (suite *DecideTestSuite) TestForcedDecisionFeatureRollout() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureRollout(feature)

	// Adding Forced Decision
	suite.tc.AddFlagVariation(feature, entities.Variation{Key: "4", FeatureEnabled: true})
	db := DecideBody{
		UserID:         "testUser",
		UserAttributes: nil,
		DecideOptions:  []string{"DISABLE_DECISION_EVENT"},
		ForcedDecisions: []ForcedDecision{
			{
				FlagKey:      "two",
				RuleKey:      "1",
				VariationKey: "2",
			},
			{
				FlagKey:      "one",
				RuleKey:      "1",
				VariationKey: "3",
			},
			{
				// Checking for last forced-decision to be considered if 2 similar rule and flagkeys's are given
				FlagKey:      "one",
				RuleKey:      "1",
				VariationKey: "4",
			},
		},
	}

	payload, err := json.Marshal(db)
	suite.NoError(err)
	suite.body = payload

	req := httptest.NewRequest("POST", "/decide?keys=one", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual client.OptimizelyDecision
	err = json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := client.OptimizelyDecision{
		UserContext:  client.OptimizelyUserContext{UserID: "testUser", Attributes: map[string]interface{}{}},
		FlagKey:      "one",
		RuleKey:      "1",
		Enabled:      true,
		VariationKey: "4",
		Reasons:      []string{},
	}

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
	suite.Equal(expected, actual)
}

func (suite *DecideTestSuite) TestForcedDecisionWithInvalidVariationKey() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureRollout(feature)

	// Adding Forced Decision
	db := DecideBody{
		UserID:         "testUser",
		UserAttributes: nil,
		DecideOptions:  []string{"DISABLE_DECISION_EVENT"},
		ForcedDecisions: []ForcedDecision{
			{
				FlagKey:      "one",
				RuleKey:      "1",
				VariationKey: "4",
			},
		},
	}

	payload, err := json.Marshal(db)
	suite.NoError(err)
	suite.body = payload

	req := httptest.NewRequest("POST", "/decide?keys=one", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual client.OptimizelyDecision
	err = json.Unmarshal(rec.Body.Bytes(), &actual)
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

func (suite *DecideTestSuite) TestForcedDecisionWithEmptyRuleKey() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureRollout(feature)

	// Adding Forced Decision
	suite.tc.AddFlagVariation(feature, entities.Variation{Key: "4", FeatureEnabled: true})
	db := DecideBody{
		UserID:         "testUser",
		UserAttributes: nil,
		DecideOptions:  []string{"DISABLE_DECISION_EVENT"},
		ForcedDecisions: []ForcedDecision{
			{
				FlagKey:      "one",
				RuleKey:      "",
				VariationKey: "4",
			},
		},
	}

	payload, err := json.Marshal(db)
	suite.NoError(err)
	suite.body = payload

	req := httptest.NewRequest("POST", "/decide?keys=one", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual client.OptimizelyDecision
	err = json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := client.OptimizelyDecision{
		UserContext:  client.OptimizelyUserContext{UserID: "testUser", Attributes: map[string]interface{}{}},
		FlagKey:      "one",
		RuleKey:      "",
		Enabled:      true,
		VariationKey: "4",
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

func DecideWithFetchSegments(suite *DecideTestSuite, userID string, fetchSegmentsOptions json.RawMessage) {
	audienceID := "odp-audience-1"
	variationKey := "variation-a"
	featureKey := "flag-segment"
	experimentKey := "experiment-segment"

	experiment := entities.Experiment{
		ID:  experimentKey,
		Key: experimentKey,
		TrafficAllocation: []entities.Range{
			{
				EntityID:   variationKey,
				EndOfRange: 10000,
			},
		},
		Variations: map[string]entities.Variation{
			variationKey: {
				ID:             variationKey,
				Key:            variationKey,
				FeatureEnabled: true,
			}},

		AudienceConditionTree: &entities.TreeNode{
			Operator: "or",
			Nodes:    []*entities.TreeNode{{Item: audienceID}},
		},
	}

	feature := entities.Feature{
		Key:                featureKey,
		FeatureExperiments: []entities.Experiment{experiment},
	}
	suite.tc.AddFeature(feature)

	audience := entities.Audience{
		ID:   audienceID,
		Name: audienceID,
		ConditionTree: &entities.TreeNode{
			Operator: "and",
			Nodes: []*entities.TreeNode{{
				Operator: "or",
				Nodes: []*entities.TreeNode{{
					Operator: "or",
					Nodes: []*entities.TreeNode{
						{
							Item: entities.Condition{
								Name:  "odp.audiences",
								Match: "qualified",
								Type:  "third_party_dimension",
								Value: "odp-segment-1",
							},
						},
					},
				}},
			}},
		},
	}
	suite.tc.AddAudience(audience)

	suite.tc.AddSegments([]string{"odp-segment-1", "odp-segment-2", "odp-segment-3"})

	db := TestDecideBody{
		UserID:               userID,
		UserAttributes:       nil,
		DecideOptions:        []string{},
		FetchSegments:        true,
		FetchSegmentsOptions: fetchSegmentsOptions,
	}
	if fetchSegmentsOptions != nil {
		db.FetchSegmentsOptions = fetchSegmentsOptions
	}

	payload, err := json.Marshal(db)
	suite.NoError(err)

	suite.body = payload

	req := httptest.NewRequest("POST", "/decide?keys=flag-segment", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual client.OptimizelyDecision
	err = json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := client.OptimizelyDecision{
		UserContext:  client.OptimizelyUserContext{UserID: userID, Attributes: map[string]interface{}{}},
		FlagKey:      featureKey,
		RuleKey:      experimentKey,
		Enabled:      true,
		VariationKey: variationKey,
		Reasons:      []string{},
	}

	suite.Equal(expected, actual)
}

func (suite *DecideTestSuite) TestDecideFetchQualifiedSegments() {
	DecideWithFetchSegments(suite, "testUser", nil)
}

func (suite *DecideTestSuite) TestFetchQualifiedSegmentsUtilizesCache() {
	DecideWithFetchSegments(suite, "testUser", nil)
	// second call should utilize cache
	DecideWithFetchSegments(suite, "testUser", nil)

	// api manager should not have been used on the second call
	assert.Equal(suite.T(), suite.tc.SegmentAPIManager.GetCallCount(), 1)
}

func (suite *DecideTestSuite) TestFetchQualifiedSegmentsIgnoresCache() {
	DecideWithFetchSegments(suite, "testUser", nil)
	DecideWithFetchSegments(suite, "testUser", json.RawMessage(fmt.Sprintf(`["%s"]`, segment.IgnoreCache)))

	// api manager should have been used on both calls
	assert.Equal(suite.T(), suite.tc.SegmentAPIManager.GetCallCount(), 2)
}

func (suite *DecideTestSuite) TestFetchQualifiedSegmentsResetsCache() {
	DecideWithFetchSegments(suite, "testUser", nil)
	DecideWithFetchSegments(suite, "secondUser", nil)
	DecideWithFetchSegments(suite, "testUser", json.RawMessage(fmt.Sprintf(`["%s"]`, segment.ResetCache)))
	DecideWithFetchSegments(suite, "secondUser", nil)
	// api manager should have been used on all calls
	assert.Equal(suite.T(), suite.tc.SegmentAPIManager.GetCallCount(), 4)
}

func (suite *DecideTestSuite) TestFetchQualifiedSegmentsIgnoreAndResetsCache() {
	DecideWithFetchSegments(suite, "testUser", nil)
	DecideWithFetchSegments(suite, "secondUser", nil)
	DecideWithFetchSegments(suite, "testUser", json.RawMessage(fmt.Sprintf(`["%s","%s"]`, segment.ResetCache, segment.IgnoreCache)))
	DecideWithFetchSegments(suite, "secondUser", nil)
	// api manager should have been used on all calls
	assert.Equal(suite.T(), suite.tc.SegmentAPIManager.GetCallCount(), 4)
}

func (suite *DecideTestSuite) TestDecideFetchQualifiedSegmentsWithInvalidOption() {
	DecideWithFetchSegments(suite, "testUser", json.RawMessage(`["INVALID_OPTION"]`))
}

func (suite *DecideTestSuite) TestDecideFetchQualifiedSegmentsWithEmptyArray() {
	DecideWithFetchSegments(suite, "testUser", json.RawMessage(`[]`))
}

func (suite *DecideTestSuite) TestDecideFetchQualifiedSegmentsWithNull() {
	DecideWithFetchSegments(suite, "testUser", json.RawMessage(`null`))
}

func (suite *DecideTestSuite) TestFetchQualifiedSegmentsInvalidMix() {
	DecideWithFetchSegments(suite, "testUser", nil)
	DecideWithFetchSegments(suite, "testUser", json.RawMessage(fmt.Sprintf(`["%s","INVALID_OPTION"]`, segment.IgnoreCache)))
	// api manager should have been used in both calls
	assert.Equal(suite.T(), suite.tc.SegmentAPIManager.GetCallCount(), 2)
}

func (suite *DecideTestSuite) TestFetchQualifiedSegmentsFailure() {
	suite.tc.AddSegments([]string{"odp-segment-1", "odp-segment-2", "odp-segment-3"})
	suite.tc.SetSegmentAPIErrorMode(true)

	db := DecideBody{
		UserID:         "testUser",
		UserAttributes: nil,
		DecideOptions:  []string{},
		FetchSegments:  true,
	}

	payload, err := json.Marshal(db)
	suite.NoError(err)

	suite.body = payload

	req := httptest.NewRequest("POST", "/decide?keys=flag-segment", bytes.NewBuffer(suite.body))

	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.assertError(rec, `failed to fetch qualified segments`, http.StatusInternalServerError)
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
