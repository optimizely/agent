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

// Package event //
package event

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/optimizely/go-sdk/pkg"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/stretchr/testify/assert"
)

// TODO(Matt): Use shared test package when available

type TestConfig struct {
	pkg.ProjectConfig
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

func TestProcessEvent(t *testing.T) {
	config := TestConfig{}

	wasCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Errorf("Error reading request body")
			return
		}

		var sentEvent event.UserEvent
		json.Unmarshal(body, &sentEvent)

		assert.Equal(t, "campaign_activated", sentEvent.Impression.Key)
		assert.Equal(t, config.GetProjectID(), sentEvent.EventContext.ProjectID)
		assert.Equal(t, config.GetRevision(), sentEvent.EventContext.Revision)

		rw.WriteHeader(http.StatusNoContent)
		rw.Write([]byte(""))
		wasCalled = true

	}))
	defer server.Close()

	experiment := entities.Experiment{}
	experiment.Key = "background_experiment"
	experiment.LayerID = "15399420423"
	experiment.ID = "15402980349"
	variation := entities.Variation{}
	variation.Key = "variation_a"
	variation.ID = "15410990633"
	userEvent := event.CreateImpressionUserEvent(config, experiment, variation, userContext)

	processor := SidedoorEventProcessor{
		URL: server.URL,
	}
	processor.ProcessEvent(userEvent)

	if !wasCalled {
		t.Errorf("Server endpoint was not called")
	}
}
