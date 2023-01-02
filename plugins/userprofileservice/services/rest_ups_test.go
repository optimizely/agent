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
	"net/http"
	"net/http/httptest"
	"sync"
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
	MethodUsed       string
	savedUserProfile decision.UserProfile
	wg               *sync.WaitGroup
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
	rups.ups.UserIDKey = "custom_user_id"
	rups.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rups.MethodUsed = r.Method
		switch r.URL.Path {
		case "/ups/save":
			userProfile := map[string]interface{}{}
			switch r.Method {
			case "GET":
				userID := r.URL.Query().Get(rups.ups.getUserIDKey())
				bucketMap := map[string]interface{}{}
				if err := json.Unmarshal([]byte(r.URL.Query().Get(experimentBucketMapKey)), &bucketMap); err != nil {
					panic(err)
				}
				userProfile[rups.ups.getUserIDKey()] = userID
				userProfile[experimentBucketMapKey] = bucketMap
			default:
				_ = handlers.ParseRequestBody(r, &userProfile)
			}
			rups.savedUserProfile = convertToUserProfile(userProfile, rups.ups.getUserIDKey())
			w.WriteHeader(http.StatusOK)
			if rups.wg != nil {
				rups.wg.Done()
			}
		case "/ups/lookup":
			userProfileMap := convertUserProfileToMap(rups.userProfile, rups.ups.getUserIDKey())
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

func (rups *RestUPSTestSuite) TestLookupDefaultPost() {
	profile := rups.ups.Lookup("a")
	rups.Equal(rups.userProfile, profile)
	rups.Equal("POST", rups.MethodUsed)
}

func (rups *RestUPSTestSuite) TestLookupGet() {
	rups.ups.LookupMethod = "GET"
	profile := rups.ups.Lookup("a")
	rups.Equal(rups.userProfile, profile)
	rups.Equal("GET", rups.MethodUsed)
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

func (rups *RestUPSTestSuite) TestSaveDefaultPOST() {
	rups.ups.Save(rups.userProfile)
	rups.Equal(rups.userProfile, rups.savedUserProfile)
	rups.Equal("POST", rups.MethodUsed)
}

func (rups *RestUPSTestSuite) TestSaveDefaultPOSTAsync() {
	rups.ups.Async = true
	rups.wg = &sync.WaitGroup{}
	rups.wg.Add(1)
	rups.ups.Save(rups.userProfile)
	rups.wg.Wait()
	rups.wg = nil
	rups.Equal(rups.userProfile, rups.savedUserProfile)
	rups.Equal("POST", rups.MethodUsed)
}

func (rups *RestUPSTestSuite) TestSaveWithGetMethod() {
	rups.ups.SaveMethod = "GET"
	rups.ups.Save(rups.userProfile)
	rups.Equal(rups.userProfile, rups.savedUserProfile)
	rups.Equal("GET", rups.MethodUsed)
}

func (rups *RestUPSTestSuite) TestSaveWithGetMethodAsync() {
	rups.ups.Async = true
	rups.wg = &sync.WaitGroup{}
	rups.ups.SaveMethod = "GET"
	rups.wg.Add(1)
	rups.ups.Save(rups.userProfile)
	rups.wg.Wait()
	rups.wg = nil
	rups.Equal(rups.userProfile, rups.savedUserProfile)
	rups.Equal("GET", rups.MethodUsed)
}

func (rups *RestUPSTestSuite) TestCustomUserID() {
	ups := RestUserProfileService{
		UserIDKey: "custom",
	}
	rups.Equal(ups.UserIDKey, ups.getUserIDKey())

	// Test default value
	ups.UserIDKey = ""
	rups.Equal(userIDKey, ups.getUserIDKey())
}

func TestRestUPSTestSuite(t *testing.T) {
	suite.Run(t, new(RestUPSTestSuite))
}
