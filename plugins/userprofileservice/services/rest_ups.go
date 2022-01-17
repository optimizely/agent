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

// Package services //
package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/optimizely/agent/plugins/userprofileservice"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/utils"
	"github.com/rs/zerolog/log"
)

// RestUserProfileService represents the rest API implementation of UserProfileService interface
type RestUserProfileService struct {
	Requester  *utils.HTTPRequester
	Host       string            `json:"host"`
	Headers    map[string]string `json:"headers"`
	LookupPath string            `json:"lookupPath"`
	SavePath   string            `json:"savePath"`
}

// Lookup is used to retrieve past bucketing decisions for users
func (r *RestUserProfileService) Lookup(userID string) (profile decision.UserProfile) {
	if userID == "" {
		return
	}

	requestURL, err := r.getURL(r.LookupPath)
	if err != nil {
		return
	}

	// Check if profile exists
	parameters := map[string]interface{}{"user_id": userID}
	success, response := r.performRequest(requestURL, parameters)
	if !success {
		return
	}

	userProfileMap := map[string]interface{}{}
	if err = json.Unmarshal(response, &userProfileMap); err != nil {
		log.Error().Msg(err.Error())
		return
	}

	return convertToUserProfile(userProfileMap)
}

// Save is used to save bucketing decisions for users
func (r *RestUserProfileService) Save(profile decision.UserProfile) {
	if profile.ID == "" {
		return
	}
	requestURL, err := r.getURL(r.SavePath)
	if err != nil {
		return
	}
	r.performRequest(requestURL, convertUserProfileToMap(profile))
}

func (r *RestUserProfileService) getURL(endpointPath string) (string, error) {
	u, err := url.Parse(r.Host + endpointPath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return u.String(), nil
	}
	return "", fmt.Errorf("invalid url components")
}

func (r *RestUserProfileService) performRequest(requestURL string, parameters map[string]interface{}) (success bool, response []byte) {
	fHeaders := []utils.Header{}
	for n, v := range r.Headers {
		fHeaders = append(fHeaders, utils.Header{Name: n, Value: v})
	}

	response, _, code, err := r.Requester.Post(requestURL, parameters, fHeaders...)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	if code == http.StatusOK {
		return true, response
	}
	return false, response
}

func init() {
	restUPSCreator := func() decision.UserProfileService {
		return &RestUserProfileService{
			Requester: utils.NewHTTPRequester(logging.GetLogger("", "RestUserProfileService")),
			Headers:   map[string]string{},
		}
	}
	userprofileservice.Add("rest", restUPSCreator)
}
