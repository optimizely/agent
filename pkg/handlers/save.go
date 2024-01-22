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
	"net/http"

	"github.com/go-chi/render"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/go-sdk/v2/pkg/decision"
)

type saveBody struct {
	UPSResponseOut
}

// Save saves the user profile against the given userId
func Save(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	logger := middleware.GetLogger(r)

	if optlyClient.UserProfileService == nil {
		RenderError(ErrNoUPS, http.StatusInternalServerError, w, r)
		return
	}

	var body saveBody
	err = ParseRequestBody(r, &body)

	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	if body.UserID == "" {
		RenderError(ErrEmptyUserID, http.StatusBadRequest, w, r)
		return
	}

	convertedProfile := convertToUserProfile(body)
	optlyClient.UserProfileService.Save(convertedProfile)
	logger.Info().Msgf("Saved user profile for user %s", body.UserID)
	render.Status(r, http.StatusOK)
}

// convertToUserProfile converts map to User Profile object
func convertToUserProfile(body saveBody) decision.UserProfile {
	userProfile := decision.UserProfile{ID: body.UserID}
	userProfile.ExperimentBucketMap = make(map[decision.UserDecisionKey]string)
	for k, v := range body.ExperimentBucketMap {
		decisionKey := decision.NewUserDecisionKey(k)
		if bucketMap, ok := v.(map[string]interface{}); ok {
			userProfile.ExperimentBucketMap[decisionKey] = bucketMap[decisionKey.Field].(string)
		}
	}
	return userProfile
}
