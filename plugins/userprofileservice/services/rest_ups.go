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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	Requester    *utils.HTTPRequester
	Host         string            `json:"host"`
	Headers      map[string]string `json:"headers"`
	LookupPath   string            `json:"lookupPath"`
	LookupMethod string            `json:"lookupMethod"`
	SavePath     string            `json:"savePath"`
	SaveMethod   string            `json:"saveMethod"`
	UserIDKey    string            `json:"userIDKey"`
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

	userIDKey := r.getUserIDKey()
	// Check if profile exists
	parameters := map[string]interface{}{userIDKey: userID}
	success, response := r.performRequest(requestURL, r.LookupMethod, parameters)
	if !success {
		return
	}

	userProfileMap := map[string]interface{}{}
	if err = json.Unmarshal(response, &userProfileMap); err != nil {
		log.Error().Msg(err.Error())
		return
	}

	return convertToUserProfile(userProfileMap, userIDKey)
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
	go r.performRequest(requestURL, r.SaveMethod, convertUserProfileToMap(profile, r.getUserIDKey()))
}

func (r *RestUserProfileService) getURL(endpointPath string) (string, error) {
	u, err := url.Parse(r.Host + endpointPath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return u.String(), nil
	}
	return "", fmt.Errorf("invalid url components")
}

func (r *RestUserProfileService) getUserIDKey() string {
	if r.UserIDKey == "" {
		return userIDKey
	}
	return r.UserIDKey
}

func (r *RestUserProfileService) performRequest(requestURL, method string, parameters map[string]interface{}) (success bool, response []byte) {
	fHeaders := []utils.Header{}
	for n, v := range r.Headers {
		fHeaders = append(fHeaders, utils.Header{Name: n, Value: v})
	}

	var body io.Reader
	fURL, err := url.Parse(requestURL)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	restAPIMethod := method
	// User Post method in case none is provided
	if method == "" {
		restAPIMethod = "POST"
	}

	switch method {
	case "GET":
		// add parameter to query string
		queryString := fURL.Query()
		for k, v := range parameters {
			switch v := v.(type) {
			// For experimentBucketMap
			case map[string]interface{}:
				if jsonStr, tmpErr := json.Marshal(v); tmpErr == nil {
					queryString.Set(k, string(jsonStr))
				}
			// For userID
			case string:
				queryString.Set(k, v)
			default:
				log.Error().Msgf("incompatible value type %T found for key %v in rest ups request parameters.", v, k)
			}
		}
		// add query to url
		fURL.RawQuery = queryString.Encode()
	default:
		jsonStr, tmpErr := json.Marshal(parameters)
		if tmpErr != nil {
			log.Error().Msg(tmpErr.Error())
			return
		}
		body = bytes.NewBuffer(jsonStr)
	}

	response, _, code, err := r.Requester.Do(fURL.String(), restAPIMethod, body, fHeaders)
	if err != nil {
		return
	}
	return code == http.StatusOK, response
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
