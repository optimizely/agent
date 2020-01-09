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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/optimizely/go-sdk/pkg/decision"

	middleware2 "github.com/optimizely/sidedoor/pkg/middleware"

	"github.com/optimizely/sidedoor/pkg/optimizely"
	"github.com/optimizely/sidedoor/pkg/optimizelytest"

	"github.com/go-chi/chi"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
	suite.Suite
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

type UserMW struct {
	optlyClient *optimizely.OptlyClient
}

func (o *UserMW) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware2.OptlyClientKey, o.optlyClient)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (o *UserMW) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		optlyContext := optimizely.NewContext("testUser", make(map[string]interface{}))
		ctx := context.WithValue(r.Context(), middleware2.OptlyContextKey, optlyContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Setup Mux
func (suite *UserTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    nil,
		ForcedVariations: testClient.ForcedVariations,
	}

	mux := chi.NewMux()
	userAPI := new(UserHandler)
	userMW := &UserMW{optlyClient}

	mux.Use(userMW.ClientCtx, userMW.UserCtx)
	mux.Post("/events/{eventKey}", userAPI.TrackEvent)
	mux.Post("/events/{eventKey}/", userAPI.TrackEvent) // Needed to assert non-empty eventKey

	mux.Get("/features", userAPI.ListFeatures)
	mux.Get("/features/{featureKey}", userAPI.GetFeature)
	mux.Post("/features", userAPI.TrackFeatures)
	mux.Post("/features/{featureKey}", userAPI.TrackFeature)

	mux.Get("/experiments/{experimentKey}", userAPI.GetVariation)
	mux.Post("/experiments/{experimentKey}", userAPI.ActivateExperiment)
	mux.Put("/experiments/{experimentKey}/variations/{variationKey}", userAPI.SetForcedVariation)
	mux.Delete("/experiments/{experimentKey}/variations", userAPI.RemoveForcedVariation)

	suite.mux = mux
	suite.tc = testClient
}

func (suite *UserTestSuite) TestGetFeatureWithFeatureTest() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureTest(feature)

	req := httptest.NewRequest("GET", "/features/one", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual Feature
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := Feature{
		Key:     "one",
		Enabled: true,
	}

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
	suite.Equal(expected, actual)
}

func (suite *UserTestSuite) TestTrackFeatureWithFeatureRollout() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureRollout(feature)

	req := httptest.NewRequest("POST", "/features/one", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual Feature
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := Feature{
		Key:     "one",
		Enabled: true,
	}

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
	suite.Equal(expected, actual)
}

func (suite *UserTestSuite) TestTrackFeatureWithFeatureTest() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureTest(feature)

	req := httptest.NewRequest("POST", "/features/one", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual Feature
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := Feature{
		Key:     "one",
		Enabled: true,
	}
	suite.Equal(expected, actual)

	events := suite.tc.GetProcessedEvents()
	suite.Equal(1, len(events))

	impression := events[0]
	suite.Equal("campaign_activated", impression.Impression.Key)
	suite.Equal("testUser", impression.VisitorID)
}

func (suite *UserTestSuite) TestGetFeaturesMissingFeature() {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req := httptest.NewRequest("POST", "/features/feature-missing", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code) // TODO should this 404
}

func (suite *UserTestSuite) TestTrackEventNoTags() {
	eventKey := "test-event"
	event := entities.Event{Key: eventKey}
	suite.tc.AddEvent(event)
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req := httptest.NewRequest("POST", "/events/"+eventKey, nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusNoContent, rec.Code)

	events := suite.tc.GetProcessedEvents()
	suite.Equal(1, len(events))

	actual := events[0]
	suite.Equal(eventKey, actual.Conversion.Key)
	suite.Equal("testUser", actual.VisitorID)
}

func (suite *UserTestSuite) TestTrackEventWithTags() {
	eventKey := "test-event"
	event := entities.Event{Key: eventKey}
	suite.tc.AddEvent(event)

	tags := make(map[string]interface{})
	tags["key1"] = "val"

	body, err := json.Marshal(tags)
	suite.NoError(err)

	req := httptest.NewRequest("POST", "/events/"+eventKey, bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusNoContent, rec.Code)

	events := suite.tc.GetProcessedEvents()
	suite.Equal(1, len(events))

	actual := events[0]
	suite.Equal(eventKey, actual.Conversion.Key)
	suite.Equal("testUser", actual.VisitorID)
	suite.Equal(tags, actual.Conversion.Tags)
}

func (suite *UserTestSuite) TestTrackEventWithInvalidTags() {
	eventKey := "test-event"
	event := entities.Event{Key: eventKey}
	suite.tc.AddEvent(event)

	req := httptest.NewRequest("POST", "/events/"+eventKey, bytes.NewBufferString("invalid"))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.assertError(rec, "error parsing request body", http.StatusBadRequest)
}

func (suite *UserTestSuite) TestTrackEventError() {
	req := httptest.NewRequest("POST", "/events/missing-event", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusNoContent, rec.Code) // TODO Should this 404?
}

func (suite *UserTestSuite) TestTrackEventEmptyKey() {
	req := httptest.NewRequest("POST", "/events//", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.assertError(rec, "missing required path parameter: eventKey", http.StatusBadRequest)
}

func (suite *UserTestSuite) TestSetForcedVariation() {
	feature := entities.Feature{Key: "my_feat"}
	suite.tc.ProjectConfig.AddMultiVariationFeatureTest(feature, "variation_disabled", "variation_enabled")
	featureExp := suite.tc.ProjectConfig.FeatureMap["my_feat"].FeatureExperiments[0]

	req := httptest.NewRequest("PUT", "/experiments/"+featureExp.Key+"/variations/variation_enabled", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusCreated, rec.Code)

	req = httptest.NewRequest("GET", "/features/my_feat", nil)
	rec = httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	var actual Feature
	json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.True(actual.Enabled)

	req = httptest.NewRequest("PUT", "/experiments/"+featureExp.Key+"/variations/variation_enabled", nil)
	rec = httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusNoContent, rec.Code)

	req = httptest.NewRequest("GET", "/features/my_feat", nil)
	rec = httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	var actualRepeated Feature
	json.Unmarshal(rec.Body.Bytes(), &actualRepeated)
	suite.True(actualRepeated.Enabled)
}

func (suite *UserTestSuite) TestSetForcedVariationEmptyExperimentKey() {
	req := httptest.NewRequest("PUT", "/experiments//variations/variation_enabled", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *UserTestSuite) TestRemoveForcedVariation() {
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

	req = httptest.NewRequest("GET", "/features/my_feat", nil)
	rec = httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
	var actual Feature
	json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.False(actual.Enabled)
}

func (suite *UserTestSuite) TestRemoveForcedVariationEmptyExperimentKey() {
	req := httptest.NewRequest("DELETE", "/experiments//variations", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *UserTestSuite) TestGetVariation() {
	testVariation := suite.tc.ProjectConfig.CreateVariation("variation_a")
	suite.tc.AddExperiment("one", []entities.Variation{testVariation})

	req := httptest.NewRequest("GET", "/experiments/one", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual Variation
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := Variation{
		Key: testVariation.Key,
		ID:  testVariation.ID,
	}

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
	suite.Equal(expected, actual)
}

func (suite *UserTestSuite) TestGetVariationMissingExperiment() {
	req := httptest.NewRequest("GET", "/experiments/one", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual Variation
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := Variation{}

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
	suite.Equal(expected, actual)
}

func (suite *UserTestSuite) TestActivateExperiment() {
	testVariation := suite.tc.ProjectConfig.CreateVariation("variation_a")
	suite.tc.AddExperiment("one", []entities.Variation{testVariation})

	req := httptest.NewRequest("POST", "/experiments/one", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual Variation
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := Variation{
		Key: testVariation.Key,
		ID:  testVariation.ID,
	}

	suite.Equal(1, len(suite.tc.GetProcessedEvents()))
	suite.Equal(expected, actual)
}

func (suite *UserTestSuite) TestListFeatures() {
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

	req := httptest.NewRequest("GET", "/features", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual []Feature
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	suite.ElementsMatch([]Feature{
		Feature{
			Enabled: true,
			Key:     "featureA",
		},
		Feature{
			Enabled: false,
			Key:     "featureB",
		},
		Feature{
			Enabled: true,
			Key:     "featureC",
			Variables: map[string]string{
				"strvar": "abc_notdef",
			},
		},
	}, actual)

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
}
func (suite *UserTestSuite) TestTrackFeatures() {
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

	req := httptest.NewRequest("POST", "/features", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual []Feature
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	suite.ElementsMatch([]Feature{
		Feature{
			Enabled: true,
			Key:     "featureA",
		},
		Feature{
			Enabled: true,
			Key:     "featureB",
		},
		Feature{
			Enabled: true,
			Key:     "featureC",
			Variables: map[string]string{
				"strvar": "abc_notdef",
			},
		},
	}, actual)

	suite.Equal(2, len(suite.tc.GetProcessedEvents()))
}

func (suite *UserTestSuite) assertError(rec *httptest.ResponseRecorder, msg string, code int) {
	assertError(suite.T(), rec, msg, code)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

func TestUserMissingClientCtx(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req := httptest.NewRequest("POST", "/", nil)

	userHandler := new(UserHandler)
	handlers := []func(w http.ResponseWriter, r *http.Request){
		userHandler.ActivateExperiment,
		userHandler.GetFeature,
		userHandler.ListFeatures,
		userHandler.GetVariation,
		userHandler.TrackFeature,
		userHandler.TrackFeatures,
		userHandler.TrackEvent,
	}

	for _, handler := range handlers {
		rec := httptest.NewRecorder()
		http.HandlerFunc(handler).ServeHTTP(rec, req)
		assertError(t, rec, "optlyClient not available", http.StatusUnprocessableEntity)
	}
}

func TestUserMissingOptlyCtx(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req := httptest.NewRequest("POST", "/", nil)
	mw := new(UserMW)

	userHandler := new(UserHandler)
	handlers := []func(w http.ResponseWriter, r *http.Request){
		userHandler.ActivateExperiment,
		userHandler.GetFeature,
		userHandler.ListFeatures,
		userHandler.GetVariation,
		userHandler.TrackFeature,
		userHandler.TrackFeatures,
		userHandler.TrackEvent,
		userHandler.SetForcedVariation,
		userHandler.RemoveForcedVariation,
	}

	for _, handler := range handlers {
		rec := httptest.NewRecorder()
		mw.ClientCtx(http.HandlerFunc(handler)).ServeHTTP(rec, req)
		assertError(t, rec, "optlyContext not available", http.StatusUnprocessableEntity)
	}
}

func assertError(t *testing.T, rec *httptest.ResponseRecorder, msg string, code int) {
	assert.Equal(t, code, rec.Code)

	var actual ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	assert.NoError(t, err)

	assert.Equal(t, ErrorResponse{Error: msg}, actual)
}
