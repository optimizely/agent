/****************************************************************************
 * Copyright 2025, Optimizely, Inc. and contributors                        *
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

// ResetClient handles the /v1/reset endpoint from FSC tests
// This clears the client cache to ensure clean state between test scenarios,
// particularly important for CMAB cache testing
func ResetClient(w http.ResponseWriter, r *http.Request) {
	// Get SDK key from header
	sdkKey := r.Header.Get("X-Optimizely-SDK-Key")
	if sdkKey == "" {
		RenderError(errors.New("SDK key required for reset"), http.StatusBadRequest, w, r)
		return
	}

	// Get the cache from context
	cache, err := middleware.GetOptlyCache(r)
	if err != nil {
		RenderError(errors.New("cache not available"), http.StatusInternalServerError, w, r)
		return
	}

	// Get logger for debugging
	logger := middleware.GetLogger(r)
	logger.Debug().Str("sdkKey", sdkKey).Msg("Resetting client for FSC test")

	// Reset the client using the cache interface
	if optlyCache, ok := cache.(interface{ ResetClient(string) }); ok {
		optlyCache.ResetClient(sdkKey)
	} else {
		RenderError(errors.New("cache reset not supported"), http.StatusInternalServerError, w, r)
		return
	}

	// Return success
	render.JSON(w, r, map[string]interface{}{"result": true})
}