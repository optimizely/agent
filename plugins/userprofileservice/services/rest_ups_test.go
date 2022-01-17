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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/render"
	"github.com/optimizely/agent/pkg/handlers"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/utils"
	"github.com/stretchr/testify/suite"
)

type RestUPSTestSuite struct {
	suite.Suite
	ups              RestUserProfileService
	server           *httptest.Server
	userProfile      decision.UserProfile
	savedUserProfile decision.UserProfile
}

func (rups *RestUPSTestSuite) SetupTest() {
	strValue := "1"
	rups.userProfile = decision.UserProfile{
		ID: strValue,
		ExperimentBucketMap: map[decision.UserDecisionKey]string{
			decision.NewUserDecisionKey(strValue): strValue,
		},
	}
	rups.savedUserProfile = decision.UserProfile{}
	rups.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case "/ups/save":
			var userProfile map[string]interface{}
			_ = handlers.ParseRequestBody(r, &userProfile)
			rups.savedUserProfile = convertToUserProfile(userProfile)
			w.WriteHeader(http.StatusOK)
		case "/ups/lookup":
			userProfileMap := convertUserProfileToMap(rups.userProfile)
			w.Header().Set("Content-Type", "application/json")
			render.JSON(w, r, userProfileMap)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	rups.ups = RestUserProfileService{
		Host:       rups.server.URL,
		Requester:  utils.NewHTTPRequester(logging.GetLogger("", "")),
		LookupPath: "/ups/lookup",
		SavePath:   "/ups/save",
		Headers:    map[string]string{"Auth-Token": "123"},
	}
}

func (rups *RestUPSTestSuite) TestGetURL() {
	// Incorrect host
	rups.ups.Host = ""
	_, err := rups.ups.getURL("lookup")
	rups.Error(err)

	rups.ups.Host = "http//google.com"
	_, err = rups.ups.getURL("lookup")
	rups.Error(err)

	// correct host
	rups.ups.Host = "http://google.com/"
	urlString, err := rups.ups.getURL("lookup")
	rups.Equal("http://google.com/lookup", urlString)
	rups.NoError(err)
}

func (rups *RestUPSTestSuite) TestEmptyUserID() {
	profile := rups.ups.Lookup("")
	rups.Assert().Empty(profile)

	rups.userProfile.ID = ""
	rups.ups.Save(rups.userProfile)
	rups.Empty(rups.savedUserProfile)
}

func (rups *RestUPSTestSuite) TestLookupInvalidHost() {
	rups.ups.Host = "abccc..as"
	profile := rups.ups.Lookup("a")
	rups.Assert().Empty(profile)

	rups.ups.Host = ""
	profile = rups.ups.Lookup("a")
	rups.Assert().Empty(profile)
}

func (rups *RestUPSTestSuite) TestLookupInvalidPath() {
	// Incorrect lookup path
	rups.ups.LookupPath = "abccc..as"
	profile := rups.ups.Lookup("a")
	rups.Assert().Empty(profile)

	// Empty lookup path
	rups.ups.LookupPath = ""
	profile = rups.ups.Lookup("a")
	rups.Assert().Empty(profile)
}

func (rups *RestUPSTestSuite) TestLookup() {
	profile := rups.ups.Lookup("a")
	rups.Equal(rups.userProfile, profile)
}

func (rups *RestUPSTestSuite) TestSaveInvalidHost() {
	rups.ups.Host = "abccc..as"
	rups.ups.Save(rups.userProfile)
	rups.Empty(rups.savedUserProfile)

	rups.ups.Host = ""
	rups.ups.Save(rups.userProfile)
	rups.Empty(rups.savedUserProfile)
}

func (rups *RestUPSTestSuite) TestSaveInvalidPath() {
	// Incorrect save path
	rups.ups.SavePath = "abccc..as"
	rups.ups.Save(rups.userProfile)
	rups.Empty(rups.savedUserProfile)

	// Empty save path
	rups.ups.SavePath = ""
	rups.ups.Save(rups.userProfile)
	rups.Empty(rups.savedUserProfile)
}

func (rups *RestUPSTestSuite) TestSave() {
	rups.ups.Save(rups.userProfile)
	rups.Equal(rups.userProfile, rups.savedUserProfile)
}

func TestRestUPSTestSuite(t *testing.T) {
	suite.Run(t, new(RestUPSTestSuite))
}
