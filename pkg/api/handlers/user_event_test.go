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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/optimizely/go-sdk/optimizely"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	optimizely.ProjectConfig
}

func (TestConfig) GetEventByKey(string) (entities.Event, error) {
	return entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, nil
}
func (TestConfig) GetFeatureByKey(string) (entities.Feature, error) {
	return entities.Feature{}, nil
}
func (TestConfig) GetProjectID() string {
	return "15389410617"
}
func (TestConfig) GetRevision() string {
	return "7"
}
func (TestConfig) GetAccountID() string {
	return "8362480420"
}
func (TestConfig) GetAnonymizeIP() bool {
	return true
}
func (TestConfig) GetAttributeID(key string) string { // returns "" if there is no id
	return ""
}
func (TestConfig) GetBotFiltering() bool {
	return false
}
func (TestConfig) GetClientName() string {
	return "go-sdk"
}
func (TestConfig) GetClientVersion() string {
	return "1.0.0"
}

var userID = "user1"
var userContext = entities.UserContext{
	ID:         userID,
	Attributes: make(map[string]interface{}),
}

func TestHandleUserEvent(t *testing.T) {
	config := TestConfig{}
	experiment := entities.Experiment{
		Key:     "background_experiment",
		LayerID: "15399420423",
		ID:      "15402980349",
	}
	variation := entities.Variation{
		Key: "variation_a",
		ID:  "15410990633",
	}
	userEvent := event.CreateImpressionUserEvent(config, experiment, variation, userContext)
	jsonValue, _ := json.Marshal(userEvent)
	req, err := http.NewRequest("POST", "/user-event", bytes.NewBuffer(jsonValue))
	if err != nil {
		t.Fatal(err)
	}
	req.Header["Content-Type"] = []string{"application/json"}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(new(UserEventHandler).AddUserEvent)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "", rr.Body.String())
}
