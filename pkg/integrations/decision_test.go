package integrations

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	requester := utils.NewHTTPRequester()
	dl := NewTemplateListener(requester, "../../templates/slack_decision_body.tmpl", "not-needed")

	message := notification.DecisionNotification{
		Type: notification.FeatureTest,
		UserContext: entities.UserContext{
			ID: "testing",
		},
		DecisionInfo: nil,
	}

	expected := "{\n  " + `"text": "Decision Made of type: feature-test for user: testing"` + "\n}"

	buf, err1 := dl.parse(message)
	assert.NoError(t, err1)

	actual, err2 := ioutil.ReadAll(buf)
	assert.NoError(t, err2)
	assert.Equal(t, expected, bytes.NewBuffer(actual).String())
}
