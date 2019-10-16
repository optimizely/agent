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
	"net/http"

	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/sidedoor/pkg/api/middleware"
)

// UserEventHandler implements the UserEventAPI interface for sending and receiving user event payloads.
type UserEventHandler struct{}

// AddUserEvent - Process a user event
func (h *UserEventHandler) AddUserEvent(w http.ResponseWriter, r *http.Request) {
	var userEvent event.UserEvent
	err := ParseRequestBody(r, userEvent)
	if err != nil {
		middleware.GetLogger(r).Error().Err(err).Msg("Error reading request body")
		RenderError(err, http.StatusBadRequest, w, r)
	}

	// TODO: Should we decode the body into interface{} and do validation
	// of that? And then only create a UserEvent after validation?
	// Or implement UnmarshalJSON for event.UserEvent, and do it all in there?

	// TODO: Do something with userEvent

	w.WriteHeader(http.StatusNoContent)
}
