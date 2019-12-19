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

// Package middleware //
package middleware

import (
	"fmt"
	"net/http"

	"github.com/optimizely/sidedoor/pkg/api/models"
	"github.com/optimizely/sidedoor/pkg/optimizely"

	"github.com/go-chi/render"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// GetOptlyClient is a utility to extract the OptlyClient from the http request context.
func GetOptlyClient(r *http.Request) (*optimizely.OptlyClient, error) {
	optlyClient, ok := r.Context().Value(OptlyClientKey).(*optimizely.OptlyClient)
	if !ok {
		return nil, fmt.Errorf("optlyClient not available")
	}

	return optlyClient, nil
}

// GetOptlyContext is a utility to extract the OptlyContext from the http request context.
func GetOptlyContext(r *http.Request) (*optimizely.OptlyContext, error) {
	optlyContext, ok := r.Context().Value(OptlyContextKey).(*optimizely.OptlyContext)
	if !ok {
		return nil, fmt.Errorf("optlyContext not available")
	}

	return optlyContext, nil
}

// GetLogger gets the logger with some info coming from http request
func GetLogger(r *http.Request) *zerolog.Logger {
	sdkKey := r.Header.Get(OptlySDKHeader)
	reqID := r.Header.Get(OptlyRequestHeader)
	logger := log.With().Str("sdkKey", sdkKey).Str("requestId", reqID).Logger()
	return &logger
}

// RenderError sets the request status and renders the error message.
func RenderError(err error, status int, w http.ResponseWriter, r *http.Request) {
	render.Status(r, status)
	render.JSON(w, r, models.ErrorResponse{Error: err.Error()})
}
