package optimizely

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/optimizely/go-sdk/optimizely"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/stretchr/testify/assert"
)

// TODO: Figure out how to share test code

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

func RandomString(len int) string {
	bytes := make([]byte, len)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	return string(bytes)
}

var userID = RandomString(10)
var userContext = entities.UserContext{
	ID:         userID,
	Attributes: make(map[string]interface{}),
}

func TestProcessEvent(t *testing.T) {

	config := TestConfig{}

	wasCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(""))
		wasCalled = true

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
