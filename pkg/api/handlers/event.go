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
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"

	"github.com/go-chi/render"

	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/rs/zerolog/log"
)

// UserEvent - Process a user event
func UserEvent(w http.ResponseWriter, r *http.Request) {
	reqContentType := r.Header.Get("Content-Type")
	reqMediaType, _, err := mime.ParseMediaType(reqContentType)
	if err != nil || reqMediaType != "application/json" {
		log.Error().Err(err).Str("Content-Type", reqContentType).Str("parsed media type", reqMediaType).Msg("Invalid Content-Type")
		render.JSON(w, r, render.M{
			"error": "Invalid content-type",
		})
		render.Status(r, http.StatusUnsupportedMediaType)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading request body")
		render.JSON(w, r, render.M{
			"error": "Error reading request body",
		})
		render.Status(r, http.StatusInternalServerError)
		return
	}

	// TODO: Should we decode the body into interface{} and do validation
	// of that? And then only create a UserEvent after validation?
	// Or implement UnmarshalJSON for event.UserEvent, and do it all in there?

	var userEvent event.UserEvent
	err = json.Unmarshal(body, &userEvent)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshaling request body")
		render.Status(r, http.StatusBadRequest)
		return
	}

	// TODO: Do something with userEvent

	render.Status(r, http.StatusNoContent)
	render.JSON(w, r, render.M{})
}
