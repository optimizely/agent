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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
)

type TrackTestSuite struct {
	suite.Suite
	oc  *optimizely.OptlyClient
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

func (suite *TrackTestSuite) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, suite.oc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Setup Mux
func (suite *TrackTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    nil,
		ForcedVariations: testClient.ForcedVariations,
	}

	mux := chi.NewMux()
	mux.With(suite.ClientCtx).Post("/track", TrackEvent)

	suite.oc = optlyClient
	suite.tc = testClient
	suite.mux = mux
}

func (suite *TrackTestSuite) TestTrackEventNoTags() {
	eventKey := "test-event"
	event := entities.Event{Key: eventKey}
	suite.tc.AddEvent(event)

	tb := trackBody{
		UserID:         "testUser",
		UserAttributes: map[string]interface{}{"test": "attr"},
		EventTags:      nil,
	}

	body, err := json.Marshal(tb)
	suite.NoError(err)

	req := httptest.NewRequest("POST", "/track?eventKey="+eventKey, bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusNoContent, rec.Code)

	events := suite.tc.GetProcessedEvents()
	suite.Equal(1, len(events))

	actual := events[0]
	suite.Equal(eventKey, actual.Conversion.Key)
	suite.Equal("testUser", actual.VisitorID)
}

func (suite *TrackTestSuite) TestTrackEventWithTags() {
	eventKey := "test-event"
	event := entities.Event{Key: eventKey}
	suite.tc.AddEvent(event)

	tb := trackBody{
		UserID:         "testUser",
		UserAttributes: nil,
		EventTags:      map[string]interface{}{"key1": "val"},
	}

	body, err := json.Marshal(tb)
	suite.NoError(err)

	req := httptest.NewRequest("POST", "/track?eventKey="+eventKey, bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusNoContent, rec.Code)

	events := suite.tc.GetProcessedEvents()
	suite.Equal(1, len(events))

	actual := events[0]
	suite.Equal(eventKey, actual.Conversion.Key)
	suite.Equal(tb.UserID, actual.VisitorID)
	suite.Equal(tb.EventTags, actual.Conversion.Tags)
}

func (suite *TrackTestSuite) TestTrackEventWithInvalidTags() {
	eventKey := "test-event"
	event := entities.Event{Key: eventKey}
	suite.tc.AddEvent(event)

	req := httptest.NewRequest("POST", "/track?eventKey="+eventKey, bytes.NewBufferString("invalid"))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.assertError(rec, "error parsing request body", http.StatusBadRequest)
}

func (suite *TrackTestSuite) TestTrackEventError() {
	req := httptest.NewRequest("POST", "/track?eventKey=invalid", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusNoContent, rec.Code) // TODO Should this 404?
}

func (suite *TrackTestSuite) TestTrackEventEmptyKey() {
	req := httptest.NewRequest("POST", "/track", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.assertError(rec, "missing required path parameter: eventKey", http.StatusBadRequest)
}

func (suite *TrackTestSuite) assertError(rec *httptest.ResponseRecorder, msg string, code int) {
	assertError(suite.T(), rec, msg, code)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestTrackTestSuite(t *testing.T) {
	suite.Run(t, new(TrackTestSuite))
}

func TestTrackMissingOptlyCtx(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()
	http.HandlerFunc(TrackEvent).ServeHTTP(rec, req)
	assertError(t, rec, "optlyClient not available", http.StatusInternalServerError)
}
