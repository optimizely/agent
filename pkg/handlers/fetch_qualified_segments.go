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
	// "fmt"
	"fmt"
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
	optimizelyUserContext.FetchQualifiedSegments(segmentOptions) // TODO: NEXT: replace []segment.OptimizelySegmentOption{} with CORRECT VALUE/ARG - segmentOptions - dioen't work - look how it's done in decide????
	// in decide it's an array of strings that is passed to Decide. What about here - are we also passing array of stings (options?) - check out gitpod example
	// NEEDS TO LOOK LIKE THIS:
	// user.FetchQualifiedSegments([]segment.OptimizelySegmentOption{segment.IgnoreCache, segment.ResetCache})

	// for decide it looks like this:
	// decisions = user.DecideAll([]decide.OptimizelyDecideOptions{decide.EnabledFlagsOnly})

	// in decide.go code:
	// d := optimizelyUserContext.Decide(key, decideOptions)

	segments := optimizelyUserContext.GetQualifiedSegments()
	print("RECEIVED SEGMENTS ", segments)

	if len(segments) > 0 {
		fmt.Printf("  >>> SEGMENTS (exist): %v", segments)
		logger.Info().Msg("Segments")
	} else {
		fmt.Printf("  >>> SEGMENTS (don't exist)): %v", segments)
		logger.Info().Msg("Segments don't exist.")
	}
	fmt.Println()

	render.JSON(w, r, segments)

	return
}

// Go uses options like so:
// decision := user.Decide("feature1", []decide.OptimizelyDecideOptions{decide.IncludeReasons})
// agent...

func getUserContextWithOdpOptions(r *http.Request) (FetchBody, error) { // TODO: - should it say "with options"??? is that from copying from  decide
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

// TODO: NEXT:
// - take care of adding options to fetchQualifiedSegments - ca be simpler, not like for DecideOptions
// - continue completing the fetch_qualified_segments.go file
// some unit tests complain when I upgrade go-sdk to the master - I can temporarily disable them just to test if fetch works
//    - notification_test.go, track_test.go
// UNIT TESTS FAILING - cause of the added ODP
// ACCEPTANCE TESTS FAILING - keeps saying port 8080 already in use!!! only agent uses it, so what is going on?

// TODO: NOW I'M TRYING TO RUN FETCH SEGMENTS TO SEE THE OUTPUT !!!!!!
