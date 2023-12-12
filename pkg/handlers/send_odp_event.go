/****************************************************************************
 * Copyright 2023, Optimizely, Inc. and contributors                        *
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
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/go-sdk/pkg/odp/event"
)

// SendOdpEvent sends event to ODP platform
func SendOdpEvent(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	logger := middleware.GetLogger(r)

	body, err := getRequestOdpEvent(r)
	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	err = optlyClient.SendOdpEvent(body.Type, body.Action, body.Identifiers, body.Data)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	returnResult := optimizely.SendOdpEventResponseModel{
		Success: true,
	}

	logger.Info().Msg("Successfully sent event to ODP platform")
	render.JSON(w, r, returnResult)
}

func getRequestOdpEvent(r *http.Request) (event.Event, error) {
	var body event.Event
	err := ParseRequestBody(r, &body)
	if err != nil {
		return event.Event{}, err
	}

	if body.Action == "" {
		return event.Event{}, errors.New(`missing "action" in request payload`)
	}

	if body.Identifiers == nil || len(body.Identifiers) == 0 {
		return event.Event{}, errors.New(`missing or empty "identifiers" in request payload`)
	}

	return body, nil
}
