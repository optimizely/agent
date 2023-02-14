/****************************************************************************
 * Copyright 2021, Optimizely, Inc. and contributors                        *
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
)

// SendBody defines the request body for decide API
type SendBody struct {
	Action      string                 `json:"action"`
	Type        string                 `json:"type"`
	Identifiers map[string]string      `json:"identifiers"`
	Data        map[string]interface{} `json:"data"`
}

// SendOdpEvent sends event to ODP platform
func SendOdpEvent(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	logger := middleware.GetLogger(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	db, err := getResponseBody(r)
	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	success := optlyClient.SendOdpEvent(db.Action, db.Type, db.Identifiers, db.Data)
	logger.Info().Bool("Success", success).Msg("Success. ODP event sent.")

	render.JSON(w, r, success)

}

var ErrIdentifiers = errors.New(`missing "identifiers" in request payload`)

func getResponseBody(r *http.Request) (SendBody, error) {
	var body SendBody
	err := ParseRequestBody(r, &body)
	if err != nil {
		return SendBody{}, err
	}

	if body.Identifiers == nil { // TODO: REPLACE NIL WITH EMPTY VALUE???
		return SendBody{}, ErrIdentifiers
	}

	// TODO: add any other error catching (for missing action, type, data?)

	return body, nil
}
