/****************************************************************************
 * Copyright 2023, Optimizely, Inc. and contributors                        *
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

	"github.com/go-chi/chi"
	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/odp/event"
	"github.com/optimizely/go-sdk/pkg/registry"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
)

type SendOdpEventTestSuite struct {
	suite.Suite
	oc  *optimizely.OptlyClient
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

func (suite *SendOdpEventTestSuite) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, suite.oc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type ErrorConfigManagerForSendOdpEvent struct{}

func (e ErrorConfigManagerForSendOdpEvent) RemoveOnProjectConfigUpdateSendOdpEvent(id int) error {
	panic("implement me")
}

func (e ErrorConfigManagerForSendOdpEvent) OnProjectConfigUpdateSendOdpEvent(callback func(notification.ProjectConfigUpdateNotification)) (int, error) {
	return 0, fmt.Errorf("config error")
}

func (e ErrorConfigManagerForSendOdpEvent) GetConfigSendOdpEvent() (config.ProjectConfig, error) {
	return nil, fmt.Errorf("config error")
}

func (e ErrorConfigManagerForSendOdpEvent) GetOptimizelyConfigSendOdpEvent() *config.OptimizelyConfig {
	panic("implement me")
}

func (e ErrorConfigManagerForSendOdpEvent) SyncConfigSendOdpEvent() {
	panic("implement me")
}

type MockConfigManagerForSendOdpEvent struct {
	config config.ProjectConfig
	sdkKey string
}

func (m MockConfigManagerForSendOdpEvent) RemoveOnProjectConfigUpdateSendOdpEvent(int) error {
	panic("implement me")
}

func (m MockConfigManagerForSendOdpEvent) OnProjectConfigUpdateSendOdpEvent(callback func(notification.ProjectConfigUpdateNotification)) (int, error) {
	notificationCenter := registry.GetNotificationCenter(m.sdkKey)
	handler := func(payload interface{}) {
		if projectConfigUpdateNotification, ok := payload.(notification.ProjectConfigUpdateNotification); ok {
			callback(projectConfigUpdateNotification)
		}
	}
	return notificationCenter.AddHandler(notification.ProjectConfigUpdate, handler)
}

func (m MockConfigManagerForSendOdpEvent) GetConfigSendOdpEvent() (config.ProjectConfig, error) {
	return m.config, nil
}

func (m MockConfigManagerForSendOdpEvent) GetOptimizelyConfigSendOdpEvent() *config.OptimizelyConfig {
	panic("implement me")
}

func (m MockConfigManagerForSendOdpEvent) SyncConfigSendOdpEvent() {
	panic("implement me")
}

// Setup Mux
func (suite *SendOdpEventTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    MockConfigManager{config: testClient.ProjectConfig},
		ForcedVariations: testClient.ForcedVariations,
	}

	mux := chi.NewMux()
	mux.With(suite.ClientCtx).Post("/send-odp-event", SendOdpEvent)

	suite.oc = optlyClient
	suite.tc = testClient
	suite.mux = mux
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSendOdpEventTestSuite(t *testing.T) {
	suite.Run(t, new(SendOdpEventTestSuite))
}

func (suite *SendOdpEventTestSuite) assertError(rec *httptest.ResponseRecorder, msg string, code int) {
	assertError(suite.T(), rec, msg, code)
}

func (suite *SendOdpEventTestSuite) TestSendOdpEvent() {
	sb := event.Event{
		Action:      "any",
		Type:        "any",
		Identifiers: map[string]string{"fs-user-id": "test-user", "email": "test@email.com"},
		Data:        nil,
	}

	body, err := json.Marshal(sb)
	suite.NoError(err)

	suite.tc.EventAPIManager.SetExpectedNumberEvents(1)

	req := httptest.NewRequest("POST", "/send-odp-event", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	var actual optimizely.SendOdpEventResponseModel
	suite.NoError(json.Unmarshal(rec.Body.Bytes(), &actual))

	expected := optimizely.SendOdpEventResponseModel{
		Success: true,
	}

	suite.Equal(expected, actual)

	events := suite.tc.EventAPIManager.GetEvents()
	suite.Equal(1, len(events))

	actualEvent := events[0]
	suite.Equal("any", actualEvent.Action)
	suite.Equal("any", actualEvent.Type)
	suite.Equal(map[string]string{"email": "test@email.com", "fs_user_id": "test-user"}, actualEvent.Identifiers)
	suite.Equal("go-sdk", actualEvent.Data["data_source"])
	suite.Equal("sdk", actualEvent.Data["data_source_type"])
	_, exists := actualEvent.Data["data_source_version"]
	suite.True(exists)
	_, exists = actualEvent.Data["idempotence_id"]
	suite.True(exists)
}

func (suite *SendOdpEventTestSuite) TestSendOdpEventMissingAction() {
	db := event.Event{
		Type:        "any",
		Identifiers: map[string]string{"fs-user-id": "testUser", "email": "test@email.com"},
		Data:        nil,
	}

	body, err := json.Marshal(db)
	suite.NoError(err)

	req := httptest.NewRequest("POST", "/send-odp-event", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusBadRequest, rec.Code)
	suite.assertError(rec, `missing "action" in request payload`, http.StatusBadRequest)
}

func (suite *SendOdpEventTestSuite) TestSendOdpEventEmptyAction() {
	db := event.Event{
		Action:      "",
		Type:        "any",
		Identifiers: map[string]string{"fs-user-id": "testUser", "email": "test@email.com"},
		Data:        nil,
	}

	body, err := json.Marshal(db)
	suite.NoError(err)

	req := httptest.NewRequest("POST", "/send-odp-event", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusBadRequest, rec.Code)
	suite.assertError(rec, `missing "action" in request payload`, http.StatusBadRequest)
}

func (suite *SendOdpEventTestSuite) TestSendOdpEventMissingIdentifiers() {
	db := event.Event{
		Action: "any",
		Type:   "any",
		Data:   nil,
	}

	body, err := json.Marshal(db)
	suite.NoError(err)

	req := httptest.NewRequest("POST", "/send-odp-event", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusBadRequest, rec.Code)
	suite.assertError(rec, `missing or empty "identifiers" in request payload`, http.StatusBadRequest)
}

func (suite *SendOdpEventTestSuite) TestSendOdpEventEmptyIdentifiers() {
	db := event.Event{
		Action:      "any",
		Type:        "any",
		Identifiers: map[string]string{},
		Data:        nil,
	}

	body, err := json.Marshal(db)
	suite.NoError(err)

	req := httptest.NewRequest("POST", "/send-odp-event", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusBadRequest, rec.Code)
	suite.assertError(rec, `missing or empty "identifiers" in request payload`, http.StatusBadRequest)
}

func (suite *SendOdpEventTestSuite) TestInvalidPayload() {
	req := httptest.NewRequest("POST", "/send-odp-event", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.assertError(rec, `missing "action" in request payload`, http.StatusBadRequest)
}
