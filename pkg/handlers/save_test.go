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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
	userprofileservices "github.com/optimizely/agent/plugins/userprofileservice/services"
	"github.com/optimizely/go-sdk/pkg/decision"
)

type SaveTestSuite struct {
	suite.Suite
	oc   *optimizely.OptlyClient
	tc   *optimizelytest.TestClient
	body []byte
	mux  *chi.Mux
}

func (suite *SaveTestSuite) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, suite.oc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (suite *SaveTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	userProfileService := &userprofileservices.InMemoryUserProfileService{
		ProfilesMap: make(map[string]decision.UserProfile),
	}
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient:   testClient.OptimizelyClient,
		ConfigManager:      nil,
		ForcedVariations:   testClient.ForcedVariations,
		UserProfileService: userProfileService,
	}

	mux := chi.NewMux()
	mux.With(suite.ClientCtx).Post("/save", Save)

	body := saveBody{
		UPSResponseOut: UPSResponseOut{
			UserID: "testUser",
		},
	}
	payload, err := json.Marshal(body)
	suite.NoError(err)

	suite.body = payload
	suite.mux = mux
	suite.tc = testClient
	suite.oc = optlyClient
}

func (suite *SaveTestSuite) TestInvalidPayload() {
	req := httptest.NewRequest("POST", "/save", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.assertError(rec, ErrEmptyUserID.Error(), http.StatusBadRequest)
}

func (suite *SaveTestSuite) TestNoUserProfileService() {
	req := httptest.NewRequest("POST", "/save", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.oc.UserProfileService = nil
	suite.mux.ServeHTTP(rec, req)

	suite.assertError(rec, ErrNoUPS.Error(), http.StatusInternalServerError)
}

func (suite *SaveTestSuite) TestSaveEmptyUserID() {
	body := saveBody{
		UPSResponseOut: UPSResponseOut{
			UserID: "",
			ExperimentBucketMap: map[string]interface{}{
				"1": map[string]interface{}{"variation_id": "2"},
				"2": map[string]interface{}{"variation_id": "2"},
				"3": map[string]interface{}{"variation_id": "3"},
			},
		},
	}
	payload, err := json.Marshal(body)
	suite.NoError(err)

	req := httptest.NewRequest("POST", "/save", bytes.NewBuffer(payload))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.assertError(rec, ErrEmptyUserID.Error(), http.StatusBadRequest)
}

func (suite *SaveTestSuite) TestSaveEmptyBucketMap() {
	req := httptest.NewRequest("POST", "/save", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	body := saveBody{
		UPSResponseOut: UPSResponseOut{
			UserID: "testUser",
		},
	}

	expected, success := convertToUserProfile(body)
	suite.True(success)

	actual := suite.oc.UserProfileService.Lookup("testUser")
	suite.Equal(expected, actual)
}

func (suite *SaveTestSuite) TestSave() {

	body := saveBody{
		UPSResponseOut: UPSResponseOut{
			UserID: "testUser",
			ExperimentBucketMap: map[string]interface{}{
				"1": map[string]interface{}{"variation_id": "2"},
				"2": map[string]interface{}{"variation_id": "2"},
				"3": map[string]interface{}{"variation_id": "3"},
			},
		},
	}

	payload, err := json.Marshal(body)
	suite.NoError(err)

	req := httptest.NewRequest("POST", "/save", bytes.NewBuffer(payload))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	expected, success := convertToUserProfile(body)
	suite.True(success)
	// Check if UPS was updated
	actual := suite.oc.UserProfileService.Lookup("testUser")
	suite.Equal(expected, actual)

	// Check if new profile overwrites older one
	body = saveBody{
		UPSResponseOut: UPSResponseOut{
			UserID: "testUser",
			ExperimentBucketMap: map[string]interface{}{
				"4": map[string]interface{}{"variation_id": "5"},
			},
		},
	}

	payload, err = json.Marshal(body)
	suite.NoError(err)
	req = httptest.NewRequest("POST", "/save", bytes.NewBuffer(payload))
	rec = httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	expected, success = convertToUserProfile(body)
	suite.True(success)

	actual = suite.oc.UserProfileService.Lookup("testUser")
	suite.Equal(expected, actual)
}

func (suite *SaveTestSuite) TestSaveEmptyProfile() {
	suite.oc.UserProfileService.Save(decision.UserProfile{
		ID: "testUser",
		ExperimentBucketMap: map[decision.UserDecisionKey]string{
			decision.NewUserDecisionKey("1"): "2",
			decision.NewUserDecisionKey("2"): "2",
			decision.NewUserDecisionKey("3"): "3",
		},
	})

	// Check if empty experiment map deletes old records
	body := saveBody{
		UPSResponseOut: UPSResponseOut{
			UserID: "testUser",
		},
	}

	payload, err := json.Marshal(body)
	suite.NoError(err)

	req := httptest.NewRequest("POST", "/save", bytes.NewBuffer(payload))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	expected, success := convertToUserProfile(body)
	suite.True(success)
	// Check if UPS was updated
	actual := suite.oc.UserProfileService.Lookup("testUser")
	suite.Equal(expected, actual)
}

func (suite *SaveTestSuite) assertError(rec *httptest.ResponseRecorder, msg string, code int) {
	assertError(suite.T(), rec, msg, code)
}

func TestSaveTestSuite(t *testing.T) {
	suite.Run(t, new(SaveTestSuite))
}
