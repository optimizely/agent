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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"

	"github.com/go-chi/chi"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/suite"
)

type EventStreamTestSuite struct {
	suite.Suite
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

type EventStreamMW struct {
	optlyClient *optimizely.OptlyClient
}

func (o *EventStreamMW) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, o.optlyClient)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (o *EventStreamMW) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		optlyContext := optimizely.NewContext("testUser", make(map[string]interface{}))
		ctx := context.WithValue(r.Context(), middleware.OptlyContextKey, optlyContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Setup Mux
func (suite *EventStreamTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    nil,
		ForcedVariations: testClient.ForcedVariations,
	}

	mux := chi.NewMux()
	eventsAPI := NewEventStreamHandler()
	EventStreamMW := &EventStreamMW{optlyClient}

	mux.Use(EventStreamMW.ClientCtx, EventStreamMW.UserCtx)
	mux.Get("/notifications/event-stream", eventsAPI.HandleEventSteam)

	suite.mux = mux
	suite.tc = testClient
}

func (suite *EventStreamTestSuite) TestFeatureTestStream() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureTest(feature)

	req := httptest.NewRequest("GET", "/notifications/event-stream", nil)
	rec := httptest.NewRecorder()

	go func() {
		time.Sleep(100)
		suite.tc.OptimizelyClient.IsFeatureEnabled("one", entities.UserContext{"testUser", make(map[string]interface{})})
	}()

	// start the mux with a go routine because the event stream will hang
	go suite.mux.ServeHTTP(rec, req)

	expected := "data: {\"Type\":\"feature\",\"UserContext\":{\"ID\":\"testUser\",\"Attributes\":{}},\"DecisionInfo\":{\"feature\":{\"featureEnabled\":true,\"featureKey\":\"one\",\"source\":\"feature-test\",\"sourceInfo\":{\"experimentKey\":\"1\",\"variationKey\":\"2\"}}}}\n\n"

	for {
		// wait for the body to come in.
		if rec.Body.Len() >= len(expected) {
			break
		}
	}
	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	response := string(rec.Body.Bytes())
	suite.Equal(expected,response)
}

func (suite *EventStreamTestSuite) TestActivateExperiment() {
	testVariation := suite.tc.ProjectConfig.CreateVariation("variation_a")
	suite.tc.AddExperiment("one", []entities.Variation{testVariation})

	req := httptest.NewRequest("GET", "/notifications/event-stream", nil)
	rec := httptest.NewRecorder()

	go func() {
		time.Sleep(100)
		suite.tc.OptimizelyClient.Activate("one", entities.UserContext{"testUser", make(map[string]interface{})})
	}()

	// start the mux with a go routine because the event stream will hang
	go suite.mux.ServeHTTP(rec, req)

	expected := "data: {\"Type\":\"ab-test\",\"UserContext\":{\"ID\":\"testUser\",\"Attributes\":{}},\"DecisionInfo\":{\"experimentKey\":\"one\",\"variationKey\":\"variation_a\"}}\n\n"

	for {
		// wait for the body to come in.
		if rec.Body.Len() >= len(expected) {
			break
		}
	}
	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	response := string(rec.Body.Bytes())
	suite.Equal(expected,response)
}

func (suite *EventStreamTestSuite) assertError(rec *httptest.ResponseRecorder, msg string, code int) {
	assertError(suite.T(), rec, msg, code)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEventStreamTestSuite(t *testing.T) {
	suite.Run(t, new(EventStreamTestSuite))
}

func TestEventStreamMissingOptlyCtx(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req := httptest.NewRequest("POST", "/", nil)
	mw := new(EventStreamMW)

	handler := NewEventStreamHandler()
	handlers := []func(w http.ResponseWriter, r *http.Request){
		handler.HandleEventSteam,
	}

	for _, handler := range handlers {
		rec := httptest.NewRecorder()
		mw.ClientCtx(http.HandlerFunc(handler)).ServeHTTP(rec, req)
		assertError(t, rec, "optlyContext not available", http.StatusUnprocessableEntity)
	}
}

