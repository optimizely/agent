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

// Package handlers //
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
)

// TestCache implements the Cache interface and is used in testing.
type TestCache struct {
	testClient *optimizelytest.TestClient
	updateConfigsCalled bool
}

// NewCache returns a new implementation of TestCache
func NewCache() *TestCache {
	testClient := optimizelytest.NewClient()
	return &TestCache{
		testClient: testClient,
		updateConfigsCalled: false,
	}
}

// GetClient returns a default OptlyClient for testing
func (tc *TestCache) GetClient(sdkKey string) (*optimizely.OptlyClient, error) {
	return &optimizely.OptlyClient{
		OptimizelyClient: tc.testClient.OptimizelyClient,
		ConfigManager:    nil,
	}, nil
}

// UpdateConfigs sets called boolean to true for testing
func (m *TestCache) UpdateConfigs(_ string) {
	m.updateConfigsCalled = true
}

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
	webhookMsg := OptlyMessage{
		ProjectID: 43,
		Timestamp: 43434343,
		Event:     "project.datafile_updated",
		Data: DatafileUpdateData{
			Revision:    101,
			OriginURL:   "origin.optimizely.com/datafiles/myDatafile",
			CDNUrl:      "cdn.optimizely.com/datafiles/myDatafile",
			Environment: "Production",
		},
	}

	validWebhookMessage, _ := json.Marshal(webhookMsg)

	req := httptest.NewRequest("POST", "/webhooks/optimizely", bytes.NewBuffer(validWebhookMessage))

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(optlyHandler.HandleWebhook)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandleWebhookValidMessageInvalidSignature(t *testing.T) {
	var testWebhookConfigs = map[int64]config.WebhookProject{
		42: {
			SDKKeys: []string{"myDatafile"},
			Secret:  "I am secret",
		},
	}
	optlyHandler := NewWebhookHandler(nil, testWebhookConfigs)
	webhookMsg := OptlyMessage{
		ProjectID: 42,
		Timestamp: 42424242,
		Event:     "project.datafile_updated",
		Data: DatafileUpdateData{
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
	handler := http.HandlerFunc(optlyHandler.HandleWebhook)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Regexp(t, "Computed signature does not match signature in request. Ignoring message.", rec.Body.String())
}

func TestHandleWebhookSkippedCheckInvalidSignature(t *testing.T) {
	testCache := NewCache()
	var testWebhookConfigs = map[int64]config.WebhookProject{
		42: {
			SDKKeys:            []string{"myDatafile"},
			Secret:             "I am secret",
			SkipSignatureCheck: true,
		},
	}
	optlyHandler := NewWebhookHandler(testCache, testWebhookConfigs)
	webhookMsg := OptlyMessage{
		ProjectID: 42,
		Timestamp: 42424242,
		Event:     "project.datafile_updated",
		Data: DatafileUpdateData{
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
	handler := http.HandlerFunc(optlyHandler.HandleWebhook)
	handler.ServeHTTP(rec, req)

	// Message is processed as usual with invalid signature as check is skipped
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, true, testCache.updateConfigsCalled)
}

func TestHandleWebhookValidMessage(t *testing.T) {
	testCache := NewCache()
	var testWebhookConfigs = map[int64]config.WebhookProject{
		42: {
			SDKKeys: []string{"myDatafile"},
			Secret:  "I am secret",
		},
	}
	optlyHandler := NewWebhookHandler(testCache, testWebhookConfigs)
	webhookMsg := OptlyMessage{
		ProjectID: 42,
		Timestamp: 42424242,
		Event:     "project.datafile_updated",
		Data: DatafileUpdateData{
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
	handler := http.HandlerFunc(optlyHandler.HandleWebhook)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, true, testCache.updateConfigsCalled)
}
