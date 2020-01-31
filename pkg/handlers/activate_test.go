package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/assert"
)

func TestGetUserContext(t *testing.T) {
	dc := ActivateBody{
		UserID: "test name",
		UserAttributes: map[string]interface{}{
			"str":    "val",
			"bool":   true,
			"double": 1.01,
			"int":    float64(10), // might be can be problematic
		},
	}

	jsonEntity, err := json.Marshal(dc)
	assert.NoError(t, err)
	req := httptest.NewRequest("GET", "/", bytes.NewBuffer(jsonEntity))

	actual, err := getUserContext(req)
	assert.NoError(t, err)

	expected := entities.UserContext{
		ID:         dc.UserID,
		Attributes: dc.UserAttributes,
	}

	assert.Equal(t, expected, actual)
}
