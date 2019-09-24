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

func TestHandleWebhookValidMessage(t *testing.T) {
	webhookMsg := models.OptlyMessage{
		ProjectId: 42,
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
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc((&OptlyWebhookHandler{}).HandleWebhook)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}
