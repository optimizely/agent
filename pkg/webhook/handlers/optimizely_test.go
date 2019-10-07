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
 * WITHOUT WArecANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/optimizely/sidedoor/pkg/optlytest"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/optimizely/sidedoor/pkg/webhook/models"
)

func TestHandleWebhookInvalidMessage(t *testing.T) {
	jsonValue, _ := json.Marshal("Invalid message")
	req := httptest.NewRequest("POST", "/optimizely/webhook", bytes.NewBuffer(jsonValue))

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc((&OptlyWebhookHandler{}).HandleWebhook)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Regexp(t, "Unable to parse webhook message.", rec.Body.String())
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

	req := httptest.NewRequest("POST", "/webhooks/optimizely", bytes.NewBuffer(validWebhookMessage))

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc((&optlyHandler).HandleWebhook)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
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

	req := httptest.NewRequest("POST", "/webhooks/optimizely", bytes.NewBuffer(validWebhookMessage))
	req.Header.Set(signatureHeader, "sha1=some_random_signature_in_header")

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc((&optlyHandler).HandleWebhook)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Regexp(t, "Computed signature does not match signature in request. Ignoring message.", rec.Body.String())
}


func TestHandleWebhookValidMessage(t *testing.T) {
	testCache := optlytest.NewCache()
	optlyHandler := OptlyWebhookHandler{
		optlyCache: testCache,
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

	req := httptest.NewRequest("POST", "/webhooks/optimizely", bytes.NewBuffer(validWebhookMessage))

	// This sha1 has been computed from the Optimizely application
	req.Header.Set(signatureHeader, "sha1=e0199de63fb7192634f52136d4ceb7dc6f191da3")

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc((&optlyHandler).HandleWebhook)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}
