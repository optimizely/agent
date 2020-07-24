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
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"github.com/optimizely/agent/config"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"

	"github.com/optimizely/agent/pkg/optimizely"
)

const signatureHeader = "X-Hub-Signature"
const signaturePrefix = "sha1="

// DatafileUpdateData model which represents data specific to datafile update
type DatafileUpdateData struct {
	Revision    int32  `json:"revision"`
	OriginURL   string `json:"origin_url"`
	CDNUrl      string `json:"cdn_url"`
	Environment string `json:"environment"`
}

// OptlyMessage model which represents any message received from Optimizely
type OptlyMessage struct {
	ProjectID int64              `json:"project_id"`
	Timestamp int64              `json:"timestamp"`
	Event     string             `json:"event"`
	Data      DatafileUpdateData `json:"data"`
}

// OptlyWebhookHandler handles incoming messages from Optimizely
type OptlyWebhookHandler struct {
	optlyCache optimizely.Cache
	ProjectMap map[int64]config.WebhookProject
}

// NewWebhookHandler returns a new instance of OptlyWebhookHandler
func NewWebhookHandler(optlyCache optimizely.Cache, projectMap map[int64]config.WebhookProject) *OptlyWebhookHandler {
	return &OptlyWebhookHandler{
		optlyCache: optlyCache,
		ProjectMap: projectMap,
	}
}

// computeSignature computes signature based on payload
func (h *OptlyWebhookHandler) computeSignature(payload []byte, secretKey string) string {
	mac := hmac.New(sha1.New, []byte(secretKey))
	_, err := mac.Write(payload)

	if err != nil {
		log.Error().Msg("Unable to compute signature.")
		return ""
	}

	return signaturePrefix + hex.EncodeToString(mac.Sum(nil))
}

// validateSignature computes and compares message digest
func (h *OptlyWebhookHandler) validateSignature(requestSignature string, payload []byte, projectID int64) bool {
	webhookConfig, ok := h.ProjectMap[projectID]
	if !ok {
		log.Error().Str("Project ID", strconv.FormatInt(projectID, 10)).Msg("No webhook configuration found for project ID.")
		return false
	}

	computedSignature := h.computeSignature(payload, webhookConfig.Secret)
	return subtle.ConstantTimeCompare([]byte(computedSignature), []byte(requestSignature)) == 1
}

// HandleWebhook handles incoming webhook messages from Optimizely application
func (h *OptlyWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Msg("Unable to read webhook message body.")
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, render.M{
			"error": "Unable to read webhook message body.",
		})
		return
	}

	var webhookMsg OptlyMessage
	err = json.Unmarshal(body, &webhookMsg)
	if err != nil {
		log.Error().Msg("Unable to parse webhook message.")
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, render.M{
			"error": "Unable to parse webhook message.",
		})
		return
	}

	// Check if there is configuration corresponding to the project
	webhookConfig, ok := h.ProjectMap[webhookMsg.ProjectID]
	if !ok {
		log.Error().Str("Project ID", strconv.FormatInt(webhookMsg.ProjectID, 10)).Msg("No webhook configured for Project ID.")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Check signature if check is not skipped
	if !webhookConfig.SkipSignatureCheck {
		requestSignature := r.Header.Get(signatureHeader)
		isValid := h.validateSignature(requestSignature, body, webhookMsg.ProjectID)
		if !isValid {
			log.Error().Msg("Computed signature does not match signature in request. Ignoring message.")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, render.M{
				"error": "Computed signature does not match signature in request. Ignoring message.",
			})
			return
		}
	}

	// Iterate through all SDK keys and update config
	for _, sdkKey := range webhookConfig.SDKKeys {
		h.optlyCache.UpdateConfigs(sdkKey)
	}
	w.WriteHeader(http.StatusNoContent)
}
