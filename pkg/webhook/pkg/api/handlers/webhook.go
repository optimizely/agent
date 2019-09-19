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
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/optimizely/sidedoor/pkg/webhook/pkg/api/models"
)

func HandleWebhook(w http.ResponseWriter, r *http.Request)  {
	decodedWebhookMsg := json.NewDecoder(r.Body)
	var webhookMsg models.WebhookMessage
	err := decodedWebhookMsg.Decode(&webhookMsg)

	if err != nil {
		log.Error().Msg("Invalid webhook message received.")
		return
	}

	// TODO: Set project config here after creating ability on Go SDK's polling config manager
	return
}
