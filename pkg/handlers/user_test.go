/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                        *
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

	"github.com/go-chi/chi"
	"github.com/optimizely/go-sdk/pkg/config"
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
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, o.optlyClient)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (o *UserMW) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		optlyContext := optimizely.NewContext("testUser", make(map[string]interface{}))
		ctx := context.WithValue(r.Context(), middleware.OptlyContextKey, optlyContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (o *UserMW) FeatureCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		featureKey := chi.URLParam(r, "featureKey")
		if featureKey == "feature-missing" {
			next.ServeHTTP(w, r)
		} else {
			feature := config.OptimizelyFeature{Key: featureKey}
			ctx := context.WithValue(r.Context(), middleware.OptlyFeatureKey, &feature)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})

}

func (o *UserMW) ExperimentCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		experimentKey := chi.URLParam(r, "experimentKey")
		if experimentKey == "experiment-missing" {
			next.ServeHTTP(w, r)
		} else {
			experiment := config.OptimizelyExperiment{Key: experimentKey}
			ctx := context.WithValue(r.Context(), middleware.OptlyExperimentKey, &experiment)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
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
	mux.With(userMW.FeatureCtx).Get("/features/{featureKey}", userAPI.GetFeature)
	mux.Post("/features", userAPI.TrackFeatures)
	mux.With(userMW.FeatureCtx).Post("/features/{featureKey}", userAPI.TrackFeature)

	mux.With(userMW.ExperimentCtx).Get("/experiments/{experimentKey}", userAPI.GetVariation)
	mux.With(userMW.ExperimentCtx).Post("/experiments/{experimentKey}", userAPI.ActivateExperiment)

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

	strvar := entities.Variable{DefaultValue: "default", ID: "123", Key: "strvar", Type: "string"}
	intvar := entities.Variable{DefaultValue: "123", ID: "124", Key: "intvar", Type: "integer"}
	doublevar := entities.Variable{DefaultValue: "123.99", ID: "125", Key: "doublevar", Type: "double"}
	boolvar := entities.Variable{DefaultValue: "true", ID: "126", Key: "boolvar", Type: "boolean"}
	feature := entities.Feature{Key: "one", VariableMap: map[string]entities.Variable{
		"strvar":    strvar,
		"intvar":    intvar,
		"doublevar": doublevar,
		"boolvar":   boolvar,
	}}
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
		Variables: map[string]interface{}{
			"strvar":    "default",
			"intvar":    "123",
			"doublevar": "123.99",
			"boolvar":   "true",
		},
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

func (suite *UserTestSuite) TestGetFeatureMissingFeature() {
	req := httptest.NewRequest("POST", "/features/feature-missing", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusInternalServerError, rec.Code)
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
	req := httptest.NewRequest("GET", "/experiments/experiment-missing", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusInternalServerError, rec.Code)

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
			Variables: map[string]interface{}{
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
			Variables: map[string]interface{}{
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
		assertError(t, rec, "optlyClient not available", http.StatusInternalServerError)
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
	}

	for _, handler := range handlers {
		rec := httptest.NewRecorder()
		mw.ClientCtx(http.HandlerFunc(handler)).ServeHTTP(rec, req)
		assertError(t, rec, "optlyClient not available", http.StatusInternalServerError)
	}
}

func assertError(t *testing.T, rec *httptest.ResponseRecorder, msg string, code int) {
	assert.Equal(t, code, rec.Code)

	var actual ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	assert.NoError(t, err)

	assert.Equal(t, ErrorResponse{Error: msg}, actual)
}
