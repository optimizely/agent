/****************************************************************************
 * Copyright 2021,2024 Optimizely, Inc. and contributors                    *
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
	"net/http"

	"github.com/go-chi/render"

	"github.com/optimizely/agent/pkg/middleware"
)

// GetDatafile returns the datafile directly from the SDK
func GetDatafile(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	logger := middleware.GetLogger(r)

	datafile := optlyClient.WithTraceContext(r.Context()).GetOptimizelyConfig().GetDatafile()
	var raw map[string]interface{}
	if err = json.Unmarshal([]byte(datafile), &raw); err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	logger.Info().Msg("Successfully returned datafile")
	render.JSON(w, r, raw)
}
