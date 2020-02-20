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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/suite"
)

type NotificationDisabledTestSuite struct {
	suite.Suite
	mux *chi.Mux
}

// Setup Mux
func (suite *NotificationDisabledTestSuite) SetupTest() {

	mux := chi.NewMux()
	eventsAPI := NewDisableNotificationHandler()

	mux.Get("/notifications/event-stream", eventsAPI.HandleEventSteam)

	suite.mux = mux

}

func (suite *NotificationDisabledTestSuite) TestEventStream() {
	req := httptest.NewRequest("GET", "/notifications/event-stream?", nil)
	rec := httptest.NewRecorder()

	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusForbidden, rec.Code)

	// Unmarshal response
	response := string(rec.Body.Bytes())
	suite.Equal("Streaming not enabled\n", response)
}

func TestDisabledEventStreamTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationDisabledTestSuite))
}
