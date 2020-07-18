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
	"strconv"
	"strings"

	"github.com/optimizely/agent/pkg/optimizely"

	"github.com/optimizely/go-sdk/pkg/config"

	"github.com/go-chi/render"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ErrorResponse Model
type ErrorResponse struct {
	Error string `json:"error"`
}

// GetOptlyClient is a utility to extract the OptlyClient from the http request context.
func GetOptlyClient(r *http.Request) (*optimizely.OptlyClient, error) {
	optlyClient, ok := r.Context().Value(OptlyClientKey).(*optimizely.OptlyClient)
	if !ok || optlyClient == nil {
		return nil, fmt.Errorf("optlyClient not available")
	}

	return optlyClient, nil
}

// GetLogger gets the logger with some info coming from http request
func GetLogger(r *http.Request) *zerolog.Logger {
	sdkKey := r.Header.Get(OptlySDKHeader)
	reqID := r.Header.Get(OptlyRequestHeader)

	sdkKeySplit := strings.Split(sdkKey, ":")
	logger := log.With().Str("sdkKey", sdkKeySplit[0]).Str("requestId", reqID).Logger()
	return &logger
}

// RenderError sets the request status and renders the error message.
func RenderError(err error, status int, w http.ResponseWriter, r *http.Request) {
	render.Status(r, status)
	render.JSON(w, r, ErrorResponse{Error: err.Error()})
}

// GetFeature returns an OptimizelyFeature from the request context
func GetFeature(r *http.Request) (*config.OptimizelyFeature, error) {
	feature, ok := r.Context().Value(OptlyFeatureKey).(*config.OptimizelyFeature)
	if !ok {
		return nil, fmt.Errorf("feature not available")
	}
	return feature, nil
}

// GetExperiment returns an OptimizelyExperiment from the request context
func GetExperiment(r *http.Request) (*config.OptimizelyExperiment, error) {
	experiment, ok := r.Context().Value(OptlyExperimentKey).(*config.OptimizelyExperiment)
	if !ok {
		return nil, fmt.Errorf("experiment not available")
	}
	return experiment, nil
}

// CoerceType coerces typed value from string
func CoerceType(s string) interface{} {

	if u, err := strconv.Unquote(s); err == nil {
		if u != s {
			return u
		}
	}

	if i, err := strconv.ParseInt(s, 10, 0); err == nil {
		return i
	}

	if d, err := strconv.ParseFloat(s, 64); err == nil {
		return d
	}

	// Not using ParseBool since is support too many variants (e.g. 0, 1, FALSE, TRUE)
	if s == "false" {
		return false
	}

	if s == "true" {
		return true
	}

	if s == "" {
		return nil
	}

	return s
}
