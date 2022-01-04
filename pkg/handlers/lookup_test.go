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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
	userprofileservices "github.com/optimizely/agent/plugins/userprofileservice/services"
)

type LookupTestSuite struct {
	suite.Suite
	oc   *optimizely.OptlyClient
	tc   *optimizelytest.TestClient
	body []byte
	mux  *chi.Mux
}

func (suite *LookupTestSuite) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, suite.oc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (suite *LookupTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	userProfileService := &userprofileservices.InMemoryUserProfileService{
		ProfilesMap: make(map[string]decision.UserProfile),
	}
	userProfileService.Save(decision.UserProfile{
		ID: "testUser",
		ExperimentBucketMap: map[decision.UserDecisionKey]string{
			decision.NewUserDecisionKey("1"): "2",
			decision.NewUserDecisionKey("2"): "2",
			decision.NewUserDecisionKey("3"): "3",
		},
	})
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient:   testClient.OptimizelyClient,
		ConfigManager:      nil,
		ForcedVariations:   testClient.ForcedVariations,
		UserProfileService: userProfileService,
	}

	mux := chi.NewMux()
	mux.With(suite.ClientCtx).Post("/lookup", Lookup)

	body := lookupBody{
		UserID: "testUser",
	}
	payload, err := json.Marshal(body)
	suite.NoError(err)

	suite.body = payload
	suite.mux = mux
	suite.tc = testClient
	suite.oc = optlyClient
}

func (suite *LookupTestSuite) TestInvalidPayload() {
	req := httptest.NewRequest("POST", "/lookup", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.assertError(rec, ErrEmptyUserID.Error(), http.StatusBadRequest)
}

func (suite *LookupTestSuite) TestNoUserProfileService() {
	req := httptest.NewRequest("POST", "/lookup", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.oc.UserProfileService = nil
	suite.mux.ServeHTTP(rec, req)

	suite.assertError(rec, ErrNoUPS.Error(), http.StatusInternalServerError)
}

func (suite *LookupTestSuite) TestNoProfileForUser() {
	body := lookupBody{
		UserID: "testUser1",
	}
	payload, err := json.Marshal(body)
	suite.NoError(err)

	req := httptest.NewRequest("POST", "/lookup", bytes.NewBuffer(payload))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := map[string]interface{}{
		"userId":              "testUser1",
		"experimentBucketMap": map[string]interface{}{},
	}
	suite.Equal(expected, actual)
}

func (suite *LookupTestSuite) TestLookupSavedProfile() {
	// lookup already saved profiles
	req := httptest.NewRequest("POST", "/lookup", bytes.NewBuffer(suite.body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := map[string]interface{}{
		"userId": "testUser",
		"experimentBucketMap": map[string]interface{}{
			"1": map[string]interface{}{"variation_id": "2"},
			"2": map[string]interface{}{"variation_id": "2"},
			"3": map[string]interface{}{"variation_id": "3"},
		},
	}
	suite.Equal(expected, actual)
}

func (suite *LookupTestSuite) assertError(rec *httptest.ResponseRecorder, msg string, code int) {
	assertError(suite.T(), rec, msg, code)
}

func TestLookupTestSuite(t *testing.T) {
	suite.Run(t, new(LookupTestSuite))
}
