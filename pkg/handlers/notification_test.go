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
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/registry"
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

// Setup Mux
func (suite *NotificationTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    nil,
		ForcedVariations: testClient.ForcedVariations,
	}

	mux := chi.NewMux()
	EventStreamMW := &NotificationMW{optlyClient}

	mux.Use(EventStreamMW.ClientCtx)
	mux.Get("/notifications/event-stream", NotificationEventSteamHandler)

	suite.mux = mux
	suite.tc = testClient
}

func (suite *NotificationTestSuite) TestFeatureTestFilter() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureTest(feature)

	req := httptest.NewRequest("GET", "/notifications/event-stream?filter="+string(notification.Track)+","+string(notification.ProjectConfigUpdate), nil)
	rec := httptest.NewRecorder()

	expected := ""

	// create a cancelable request context
	ctx := req.Context()
	ctx1, _ := context.WithTimeout(ctx, 1*time.Second)

	go func() {
		suite.tc.OptimizelyClient.IsFeatureEnabled("one", entities.UserContext{"testUser", make(map[string]interface{})})
	}()

	suite.mux.ServeHTTP(rec, req.WithContext(ctx1))

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	response := string(rec.Body.Bytes())
	suite.Equal(expected, response)
}

func (suite *NotificationTestSuite) TestFilter() {
	filter := []string{"decision", "track"}

	notifications := getFilter(filter)

	suite.True(len(notifications) == 2)
	suite.EqualValues(notification.Track, notifications["track"])
	suite.EqualValues(notification.Decision, notifications["decision"])

	filter = []string{"decision,track", "track"}

	notifications = getFilter(filter)

	suite.True(len(notifications) == 2)
	suite.EqualValues(notification.Track, notifications["track"])
	suite.EqualValues(notification.Decision, notifications["decision"])
}

func (suite *NotificationTestSuite) TestTrackAndProjectConfig() {
	event := entities.Event{Key: "one"}
	suite.tc.AddEvent(event)

	req := httptest.NewRequest("GET", "/notifications/event-stream", nil)
	rec := httptest.NewRecorder()

	expected := `data: {"test":"value"}` + "\n\n" + `data: {"Type":"project_config_update","Revision":"revision"}` + "\n\n"

	// create a cancelable request context
	ctx := req.Context()
	ctx1, _ := context.WithTimeout(ctx, 3*time.Second)

	nc := registry.GetNotificationCenter("")

	go func() {
		time.Sleep(1 * time.Second)
		_ = nc.Send(notification.Track, map[string]string{"test": "value"})
		projectConfigUpdateNotification := notification.ProjectConfigUpdateNotification{
			Type:     notification.ProjectConfigUpdate,
			Revision: suite.tc.ProjectConfig.GetRevision(),
		}
		_ = nc.Send(notification.ProjectConfigUpdate, projectConfigUpdateNotification)
	}()

	suite.mux.ServeHTTP(rec, req.WithContext(ctx1))

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	response := string(rec.Body.Bytes())
	suite.Equal(expected, response)
}

func (suite *NotificationTestSuite) TestActivateExperimentRaw() {
	testVariation := suite.tc.ProjectConfig.CreateVariation("variation_a")
	suite.tc.AddExperiment("one", []entities.Variation{testVariation})

	req := httptest.NewRequest("GET", "/notifications/event-stream?raw=yes", nil)
	rec := httptest.NewRecorder()

	expected := `{"key":"value"}` + "\n"

	// create a cancelable request context
	ctx := req.Context()
	ctx1, _ := context.WithTimeout(ctx, 2*time.Second)

	nc := registry.GetNotificationCenter("")
	go func() {
		time.Sleep(1 * time.Second)
		nc.Send(notification.Decision, map[string]string{"key": "value"})
	}()

	suite.mux.ServeHTTP(rec, req.WithContext(ctx1))

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	response := string(rec.Body.Bytes())
	suite.Equal(expected, response)
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

	handlers := []func(w http.ResponseWriter, r *http.Request){
		NotificationEventSteamHandler,
	}

	for _, handler := range handlers {
		rec := httptest.NewRecorder()
		mw.ClientCtx(http.HandlerFunc(handler)).ServeHTTP(rec, req)
		assertError(t, rec, "optlyClient not available", http.StatusUnprocessableEntity)
	}
}
