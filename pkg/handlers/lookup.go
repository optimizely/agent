/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                        *
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

	"github.com/optimizely/agent/pkg/middleware"

	"github.com/go-chi/render"
)

type lookupBody struct {
	UserID string `json:"userId"`
}

// UPSResponseOut defines the UPS API response
type UPSResponseOut struct {
	UserID              string                 `json:"userId"`
	ExperimentBucketMap map[string]interface{} `json:"experimentBucketMap"`
}

// ErrNoUPS is a constant error if no user profile service was found
var ErrNoUPS = errors.New(`no user profile service found`)

// Lookup searches for user profile for the given userId
func Lookup(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	if optlyClient.UserProfileService == nil {
		RenderError(ErrNoUPS, http.StatusInternalServerError, w, r)
		return
	}

	var body lookupBody
	err = ParseRequestBody(r, &body)

	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	if body.UserID == "" {
		RenderError(ErrEmptyUserID, http.StatusBadRequest, w, r)
		return
	}

	lookupResponse := UPSResponseOut{
		UserID: body.UserID,
	}
	savedProfile := optlyClient.UserProfileService.Lookup(body.UserID)
	// Converting to map
	experimentBucketMap := map[string]interface{}{}
	for k, v := range savedProfile.ExperimentBucketMap {
		experimentBucketMap[k.ExperimentID] = map[string]interface{}{k.Field: v}
	}
	lookupResponse.ExperimentBucketMap = experimentBucketMap
	render.JSON(w, r, lookupResponse)
}
