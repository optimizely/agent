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
	"net/http"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/go-sdk/pkg/decision"

	"github.com/go-chi/render"
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

	convertedProfile, success := convertToUserProfile(body)
	if !success {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}
	optlyClient.UserProfileService.Save(convertedProfile)
	render.Status(r, http.StatusOK)
}

// convertToUserProfile converts map to User Profile object
func convertToUserProfile(body saveBody) (profile decision.UserProfile, success bool) {
	userProfile := decision.UserProfile{ID: body.UserID}
	userProfile.ExperimentBucketMap = make(map[decision.UserDecisionKey]string)
	for k, v := range body.ExperimentBucketMap {
		decisionKey := decision.NewUserDecisionKey(k)
		if bucketMap, ok := v.(map[string]interface{}); ok {
			userProfile.ExperimentBucketMap[decisionKey] = bucketMap[decisionKey.Field].(string)
		}
	}
	return userProfile, true
}
