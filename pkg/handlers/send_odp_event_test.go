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
	"net/http"
	"net/http/httptest"
	"strings"

	"testing"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/suite"
)

type SendOdpEventTestSuite struct {
	suite.Suite
	oc   *optimizely.OptlyClient
	tc   *optimizelytest.TestClient
	body []byte
	mux  *chi.Mux
}

func (suite *SendOdpEventTestSuite) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, suite.oc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (suite *SendOdpEventTestSuite) assertError(rec *httptest.ResponseRecorder, msg string, code int) {
	assertError(suite.T(), rec, msg, code)
}

func TestSendOdpEventTestSuite(t *testing.T) {
	suite.Run(t, new(SendOdpEventTestSuite))
}

func (suite *SendOdpEventTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    nil,
	}

	mux := chi.NewMux()
	mux.With(suite.ClientCtx).Post("/send-odp-event", SendOdpEvent)

	suite.mux = mux
	suite.tc = testClient
	suite.oc = optlyClient
}

func (suite *SendOdpEventTestSuite) TestInvalidPayload() {
	req := httptest.NewRequest("POST", "/send-odp-event", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.assertError(rec, `missing "action" in request payload`, http.StatusBadRequest)
}

func (suite *SendOdpEventTestSuite) TestSendOdpEvent() {
	db := SendBody{
		Action:      "any",
		Type:        "any",
		Identifiers: map[string]string{"fs-user-id": "test-user", "email": "test@email.com"},
		Data:        nil,
	}

	payload, err := json.Marshal(db)
	suite.NoError(err)

	suite.body = payload

	req := httptest.NewRequest("POST", "/send-odp-event", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
	expected := strings.TrimSuffix(rec.Body.String(), "\n")
	suite.Equal("true", expected)

}

func (suite *SendOdpEventTestSuite) TestSendOdpEventMissingTypeAndData() {
	db := SendBody{
		Action:      "any",
		Identifiers: map[string]string{"fs-user-id": "testUser", "email": "test@email.com"},
	}

	payload, err := json.Marshal(db)
	suite.NoError(err)

	suite.body = payload

	req := httptest.NewRequest("POST", "/send-odp-event", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	suite.Equal(0, len(suite.tc.GetProcessedEvents()))
	expected := strings.TrimSuffix(rec.Body.String(), "\n")
	suite.Equal("true", expected)
}

func (suite *SendOdpEventTestSuite) TestSendOdpEventMissingAction() {
	db := SendBody{
		Action:      "",
		Type:        "any",
		Identifiers: map[string]string{"fs-user-id": "testUser", "email": "test@email.com"},
		Data:        nil,
	}

	payload, err := json.Marshal(db)
	suite.NoError(err)

	suite.body = payload

	req := httptest.NewRequest("POST", "/send-odp-event", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusBadRequest, rec.Code)
	suite.assertError(rec, `missing "action" in request payload`, http.StatusBadRequest)
}

func (suite *SendOdpEventTestSuite) TestSendOdpEventMissingIdentifiers() {
	db := SendBody{
		Action: "any",
		Type:   "any",
		Data:   nil,
	}

	payload, err := json.Marshal(db)
	suite.NoError(err)

	suite.body = payload

	req := httptest.NewRequest("POST", "/send-odp-event", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusBadRequest, rec.Code)
	suite.assertError(rec, `missing or empty "identifiers" in request payload`, http.StatusBadRequest)
}

func (suite *SendOdpEventTestSuite) TestSendOdpEventEmptyIdentifiers() {
	db := SendBody{
		Action:      "any",
		Type:        "any",
		Identifiers: map[string]string{},
		Data:        nil,
	}

	payload, err := json.Marshal(db)
	suite.NoError(err)

	suite.body = payload

	req := httptest.NewRequest("POST", "/send-odp-event", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusBadRequest, rec.Code)
	suite.assertError(rec, `missing or empty "identifiers" in request payload`, http.StatusBadRequest)
}

// somehw test when successis false - but how to make it false??? (in the test I can mock!)
