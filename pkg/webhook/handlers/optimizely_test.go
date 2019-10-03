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

package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/optimizely/sidedoor/pkg/optimizely"
	"github.com/optimizely/sidedoor/pkg/webhook/models"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleWebhookInvalidMessage(t *testing.T) {
	jsonValue, _ := json.Marshal("Invalid message")
	req, err := http.NewRequest("POST", "/optimizely/webhook", bytes.NewBuffer(jsonValue))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc((&OptlyWebhookHandler{}).HandleWebhook)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Regexp(t, "Unable to parse webhook message.", rr.Body.String())
}

func TestHandleWebhookNoWebhookForProject(t *testing.T) {
	optlyHandler := OptlyWebhookHandler{}
	optlyHandler.Init()
	webhookMsg := models.OptlyMessage{
		ProjectID: 43,
		Timestamp: 43434343,
		Event:     "project.datafile_updated",
		Data:      models.DatafileUpdateData{
			Revision:    101,
			OriginURL:   "origin.optimizely.com/datafiles/myDatafile",
			CDNUrl:      "cdn.optimizely.com/datafiles/myDatafile",
			Environment: "Production",
		},
	}

	validWebhookMessage, _ := json.Marshal(webhookMsg)

	req, err := http.NewRequest("POST", "/webhooks/optimizely", bytes.NewBuffer(validWebhookMessage))
	assert.Nil(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc((&optlyHandler).HandleWebhook)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestHandleWebhookValidMessageInvalidSignature(t *testing.T) {
	optlyHandler := OptlyWebhookHandler{}
	optlyHandler.Init()
	webhookMsg := models.OptlyMessage{
		ProjectID: 42,
		Timestamp: 42424242,
		Event:     "project.datafile_updated",
		Data:      models.DatafileUpdateData{
			Revision:    101,
			OriginURL:   "origin.optimizely.com/datafiles/myDatafile",
			CDNUrl:      "cdn.optimizely.com/datafiles/myDatafile",
			Environment: "Production",
		},
	}

	validWebhookMessage, _ := json.Marshal(webhookMsg)

	req, err := http.NewRequest("POST", "/webhooks/optimizely", bytes.NewBuffer(validWebhookMessage))
	assert.Nil(t, err)
	req.Header.Set(signatureHeader, "sha1=some_random_signature_in_header")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc((&optlyHandler).HandleWebhook)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Regexp(t, "Computed signature does not match signature in request. Ignoring message.", rr.Body.String())
}


func TestHandleWebhookValidMessage(t *testing.T) {
	optlyHandler := OptlyWebhookHandler{
		optlyCache: optimizely.NewCache(),
	}
	optlyHandler.Init()
	webhookMsg := models.OptlyMessage{
		ProjectID: 42,
		Timestamp: 42424242,
		Event:     "project.datafile_updated",
		Data:      models.DatafileUpdateData{
			Revision:    101,
			OriginURL:   "origin.optimizely.com/datafiles/myDatafile",
			CDNUrl:      "cdn.optimizely.com/datafiles/myDatafile",
			Environment: "Production",
		},
	}

	validWebhookMessage, _ := json.Marshal(webhookMsg)

	req, err := http.NewRequest("POST", "/webhooks/optimizely", bytes.NewBuffer(validWebhookMessage))
	assert.Nil(t, err)
	// This sha1 has been computed from the Optimizely application
	req.Header.Set(signatureHeader, "sha1=e0199de63fb7192634f52136d4ceb7dc6f191da3")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc((&optlyHandler).HandleWebhook)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}
