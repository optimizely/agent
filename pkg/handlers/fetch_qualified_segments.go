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
	// "fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/go-sdk/pkg/odp/segment"
)

// FetchBody defines the request body for decide API
type FetchBody struct {
	UserID         string                 `json:"userId"`
	UserAttributes map[string]interface{} `json:"userAttributes"`
	SegmentOptions []string               `json:"segmentOptions"`
}

// FetchQualifiedSegments fetches qualified segments from ODP platform
func FetchQualifiedSegments(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	logger := middleware.GetLogger(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	db, err := getUserContextWithOdpOptions(r)
	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	segmentOptions, err := segment.TranslateOptions(db.SegmentOptions)
	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	// Fetch qualified segments
	optimizelyUserContext := optlyClient.CreateUserContext(db.UserID, db.UserAttributes)
	optimizelyUserContext.FetchQualifiedSegments(segmentOptions)
	segments := optimizelyUserContext.GetQualifiedSegments()
	logger.Debug().Msg("Fetching ODP segments")
	render.JSON(w, r, segments)
}

func getUserContextWithOdpOptions(r *http.Request) (FetchBody, error) {
	var body FetchBody
	err := ParseRequestBody(r, &body)
	if err != nil {
		return FetchBody{}, err
	}

	if body.UserID == "" {
		return FetchBody{}, ErrEmptyUserID
	}

	return body, nil
}
