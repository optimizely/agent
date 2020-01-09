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
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/render"

	middleware2 "github.com/optimizely/sidedoor/pkg/middleware"
)

// ErrorResponse Model
type ErrorResponse struct {
	Error string `json:"error"`
}

// RenderError sets the request status and renders the error message.
func RenderError(err error, status int, w http.ResponseWriter, r *http.Request) {
	render.Status(r, status)
	render.JSON(w, r, ErrorResponse{Error: err.Error()})
}

// ParseRequestBody reads the request body from the request and unmarshals it
// into the provided interface. Note that we're sanitizing the returned error
// so that it is not leaked back to the requestor.
func ParseRequestBody(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := "error reading request body"
		middleware2.GetLogger(r).Error().Err(err).Msg(msg)
		return fmt.Errorf(msg)
	}

	if len(body) == 0 {
		middleware2.GetLogger(r).Debug().Msg("body was empty skip JSON unmarshal")
		return nil
	}

	err = json.Unmarshal(body, &v)
	if err != nil {
		msg := "error parsing request body"
		middleware2.GetLogger(r).Error().Err(err).Msg(msg)
		return fmt.Errorf(msg)
	}

	return nil
}
