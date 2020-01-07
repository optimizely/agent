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
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"

	"github.com/optimizely/sidedoor/pkg/optimizely"
	"github.com/optimizely/sidedoor/pkg/webhook/models"
)

const signatureHeader = "X-Hub-Signature"
const signaturePrefix = "sha1="

// OptlyWebhookHandler handles incoming messages from Optimizely
type OptlyWebhookHandler struct {
	optlyCache       optimizely.Cache
	webhookConfigMap map[int64]models.OptlyWebhookConfig
}

// NewWebhookHandler returns a new instance of OptlyWebhookHandler
func NewWebhookHandler(optlyCache optimizely.Cache, webhookConfigs []models.OptlyWebhookConfig) *OptlyWebhookHandler {
	configMap := make(map[int64]models.OptlyWebhookConfig)
	for _, config := range webhookConfigs {
		configMap[config.ProjectID] = config
	}

	return &OptlyWebhookHandler{
		optlyCache:       optlyCache,
		webhookConfigMap: configMap,
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
	webhookConfig, ok := h.webhookConfigMap[projectID]
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

	var webhookMsg models.OptlyMessage
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
	webhookConfig, ok := h.webhookConfigMap[webhookMsg.ProjectID]
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
		optlyClient, err := h.optlyCache.GetClient(sdkKey)
		if err != nil {
			log.Error().Str("SDK key", sdkKey).Msg("No client found for SDK key.")
			continue
		}
		optlyClient.UpdateConfig()
	}
	w.WriteHeader(http.StatusNoContent)
}