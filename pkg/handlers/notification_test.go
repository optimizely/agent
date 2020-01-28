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

type NotificationTestSuite struct {
	suite.Suite
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

type NotificationMW struct {
	optlyClient *optimizely.OptlyClient
}

func (o *NotificationMW) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, o.optlyClient)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (o *NotificationMW) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		optlyContext := optimizely.NewContext("testUser", make(map[string]interface{}))
		ctx := context.WithValue(r.Context(), middleware.OptlyContextKey, optlyContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Setup Mux
func (suite *NotificationTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    nil,
		ForcedVariations: testClient.ForcedVariations,
	}

	mux := chi.NewMux()
	eventsAPI := NewEventStreamHandler()
	EventStreamMW := &NotificationMW{optlyClient}

	mux.Use(EventStreamMW.ClientCtx, EventStreamMW.UserCtx)
	mux.Get("/notifications/event-stream", eventsAPI.HandleEventSteam)

	suite.mux = mux
	suite.tc = testClient
}

func (suite *NotificationTestSuite) TestFeatureTestStream() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureTest(feature)

	req := httptest.NewRequest("GET", "/notifications/event-stream", nil)
	rec := httptest.NewRecorder()

	go func() {
		// sleep for 100 milliseconds to allow the handler to start
		time.Sleep(100)
		suite.tc.OptimizelyClient.IsFeatureEnabled("one", entities.UserContext{"testUser", make(map[string]interface{})})
	}()

	// create a cancelable request context
	ctx := req.Context()
	ctx1,cancelFun := context.WithCancel(ctx)
	// wait 1 second for the request to be serviced and then cancel the request which closes the request
	// and notifies the handler. This causes the handler to shuts down properly
	// wait 1 second for the request to be serviced and then cancel the request which closes the request
	// notifies the handler, and the handler shuts down properly
	go func() {
		time.Sleep(1 * time.Second)
		cancelFun()
	}()

	expected := "data: {\"Type\":\"feature\",\"UserContext\":{\"ID\":\"testUser\",\"Attributes\":{}},\"DecisionInfo\":{\"feature\":{\"featureEnabled\":true,\"featureKey\":\"one\",\"source\":\"feature-test\",\"sourceInfo\":{\"experimentKey\":\"1\",\"variationKey\":\"2\"}}}}\n\n"

	suite.mux.ServeHTTP(rec, req.WithContext(ctx1))
	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	response := string(rec.Body.Bytes())
	suite.Equal(expected,response)
}

func (suite *NotificationTestSuite) TestActivateExperiment() {
	testVariation := suite.tc.ProjectConfig.CreateVariation("variation_a")
	suite.tc.AddExperiment("one", []entities.Variation{testVariation})

	req := httptest.NewRequest("GET", "/notifications/event-stream", nil)
	rec := httptest.NewRecorder()

	go func() {
		// sleep for 100 milliseconds to allow the handler to start
		time.Sleep(100)
		suite.tc.OptimizelyClient.Activate("one", entities.UserContext{"testUser", make(map[string]interface{})})
	}()

	expected := "data: {\"Type\":\"ab-test\",\"UserContext\":{\"ID\":\"testUser\",\"Attributes\":{}},\"DecisionInfo\":{\"experimentKey\":\"one\",\"variationKey\":\"variation_a\"}}\n\n"

	// create a cancelable request context
	ctx := req.Context()
	ctx1,cancelFun := context.WithCancel(ctx)

	// wait 1 second for the request to be serviced and then cancel the request which closes the request
	// notifies the handler, and the handler shuts down properly
	go func() {
		time.Sleep(1 * time.Second)
		cancelFun()
	}()

	suite.mux.ServeHTTP(rec, req.WithContext(ctx1))
	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	response := string(rec.Body.Bytes())
	suite.Equal(expected,response)
}

func (suite *NotificationTestSuite) TestActivateExperimentRaw() {
	testVariation := suite.tc.ProjectConfig.CreateVariation("variation_a")
	suite.tc.AddExperiment("one", []entities.Variation{testVariation})

	req := httptest.NewRequest("GET", "/notifications/event-stream?raw=yes", nil)
	rec := httptest.NewRecorder()

	go func() {
		time.Sleep(100)
		suite.tc.OptimizelyClient.Activate("one", entities.UserContext{"testUser", make(map[string]interface{})})
	}()

	expected := "{\"Type\":\"ab-test\",\"UserContext\":{\"ID\":\"testUser\",\"Attributes\":{}},\"DecisionInfo\":{\"experimentKey\":\"one\",\"variationKey\":\"variation_a\"}}\n"

	// create a cancelable request context
	ctx := req.Context()
	ctx1,cancelFun := context.WithCancel(ctx)

	// wait 1 second for the request to be serviced and then cancel the request which closes the request
	// and notifies the handler. This causes the handler to shuts down properly
	go func() {
		time.Sleep(1 * time.Second)
		cancelFun()
	}()

	// start the mux
	suite.mux.ServeHTTP(rec, req.WithContext(ctx1))

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	response := string(rec.Body.Bytes())
	suite.Equal(expected,response)
}

func (suite *NotificationTestSuite) assertError(rec *httptest.ResponseRecorder, msg string, code int) {
	assertError(suite.T(), rec, msg, code)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEventStreamTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationTestSuite))
}

func TestEventStreamMissingOptlyCtx(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req := httptest.NewRequest("GET", "/", nil)
	mw := new(NotificationMW)
	mw.optlyClient = nil

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

