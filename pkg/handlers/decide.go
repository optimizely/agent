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

	"github.com/optimizely/agent/pkg/middleware"

	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/decide"

	"github.com/go-chi/render"
)

// DecideBody defines the request body for decide API
type DecideBody struct {
	UserID         string                 `json:"userId"`
	UserAttributes map[string]interface{} `json:"userAttributes"`
	DecideOptions  []string               `json:"decideOptions"`
}

// DecideOut defines the response
type DecideOut struct {
	client.OptimizelyDecision
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// Decide makes feature decisions for the selected query parameters
func Decide(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	logger := middleware.GetLogger(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	db, err := getUserContextWithOptions(r)
	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	decideOptions, err := translateOptions(db.DecideOptions)
	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	optimizelyUserContext := optlyClient.CreateUserContext(db.UserID, db.UserAttributes)

	r.ParseForm()
	keys := r.Form["keys"]

	decides := map[string]client.OptimizelyDecision{}
	switch len(keys) {
	case 0:
		// Decide All
		decides = optimizelyUserContext.DecideAll(decideOptions)
	case 1:
		// Decide
		key := keys[0]
		logger.Debug().Str("featureKey", key).Msg("fetching feature decision")
		d := optimizelyUserContext.Decide(key, decideOptions)
		decideOut := DecideOut{d, d.Variables.ToMap()}
		render.JSON(w, r, decideOut)
		return
	default:
		// Decide for Keys
		decides = optimizelyUserContext.DecideForKeys(keys, decideOptions)
	}

	decideOuts := []DecideOut{}
	for _, d := range decides {
		decideOut := DecideOut{d, d.Variables.ToMap()}
		decideOuts = append(decideOuts, decideOut)
	}
	render.JSON(w, r, decideOuts)
}

func getUserContextWithOptions(r *http.Request) (DecideBody, error) {
	var body DecideBody
	err := ParseRequestBody(r, &body)
	if err != nil {
		return DecideBody{}, err
	}

	if body.UserID == "" {
		return DecideBody{}, ErrEmptyUserID
	}

	return body, nil
}

func translateOptions(options []string) ([]decide.OptimizelyDecideOptions, error) {
	decideOptions := []decide.OptimizelyDecideOptions{}
	for _, val := range options {
		switch val {
		case "DISABLE_DECISION_EVENT":
			decideOptions = append(decideOptions, decide.DisableDecisionEvent)
		case "ENABLED_FLAGS_ONLY":
			decideOptions = append(decideOptions, decide.EnabledFlagsOnly)
		case "IGNORE_USER_PROFILE_SERVICE":
			decideOptions = append(decideOptions, decide.IgnoreUserProfileService)
		case "EXCLUDE_VARIABLES":
			decideOptions = append(decideOptions, decide.ExcludeVariables)
		case "INCLUDE_REASONS":
			decideOptions = append(decideOptions, decide.IncludeReasons)
		default:
			return []decide.OptimizelyDecideOptions{}, errors.New("invalid option: " + val)
		}
	}
	return decideOptions, nil
}
