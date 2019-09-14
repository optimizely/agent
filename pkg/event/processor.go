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

// Package event //
package event

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/rs/zerolog/log"
)

const jsonContentType = "application/json"

// SidedoorEventProcessor - sends events to sidedoor API
type SidedoorEventProcessor struct {
	URL string
}

// ProcessEvent - send event to sidedoor API
func (s *SidedoorEventProcessor) ProcessEvent(userEvent event.UserEvent) error {
	jsonValue, err := json.Marshal(userEvent)
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling event")
		return err
	}

	resp, err := http.Post(s.URL, jsonContentType, bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Error().Err(err).Msg("Error sending request")
		return err
	}

	err = resp.Body.Close()
	if err != nil {
		log.Error().Err(err).Msg("Error closing response body")
		return err
	}

	return nil
}
