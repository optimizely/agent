// handlers_test.go
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
	experiment := entities.Experiment{}
	experiment.Key = "background_experiment"
	experiment.LayerID = "15399420423"
	experiment.ID = "15402980349"
	variation := entities.Variation{}
	variation.Key = "variation_a"
	variation.ID = "15410990633"
	userEvent := event.CreateImpressionUserEvent(config, experiment, variation, userContext)
	jsonValue, _ := json.Marshal(userEvent)
	req, err := http.NewRequest("POST", "/events/userEvent", bytes.NewBuffer(jsonValue))
	req.Header["Content-Type"] = []string{"application/json"}
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(UserEvent)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "", rr.Body.String())
}
