/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/optimizely/sidedoor/pkg/api/middleware"
	"github.com/optimizely/sidedoor/pkg/api/models"

	"github.com/optimizely/sidedoor/pkg/optimizely"
	"github.com/optimizely/sidedoor/pkg/optimizelytest"

	"github.com/go-chi/chi"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
	suite.Suite
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

type UserMW struct {
	optlyClient *optimizely.OptlyClient
}

func (o *UserMW) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, o.optlyClient)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (o *UserMW) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		optlyContext := optimizely.NewContext("testUser", make(map[string]interface{}))
		ctx := context.WithValue(r.Context(), middleware.OptlyContextKey, optlyContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Setup Mux
func (suite *UserTestSuite) SetupTest() {

	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{testClient.OptimizelyClient, nil}

	mux := chi.NewMux()
	userAPI := new(UserHandler)
	userMW := &UserMW{optlyClient}

	mux.Use(userMW.ClientCtx)
	mux.With(userMW.UserCtx).Post("/features/{featureKey}", userAPI.ActivateFeature)

	suite.mux = mux
	suite.tc = testClient
}

func (suite *UserTestSuite) TestActivateFeature() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureRollout(feature)

	req, err := http.NewRequest("POST", "/features/one", nil)
	suite.Nil(err)

	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual models.Feature
	err = json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	expected := models.Feature{
		Key:     "one",
		Enabled: true,
	}

	suite.Equal(expected, actual)
}

func (suite *UserTestSuite) TestGetFeaturesMissingFeature() {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/features/feature-404", nil)
	suite.Nil(err)

	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusInternalServerError, rec.Code)
	// Unmarshal response
	var actual models.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	suite.Equal(models.ErrorResponse{Error: `Feature with key feature-404 not found`}, actual)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

func TestUserMissingClientCtx(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req := httptest.NewRequest("POST", "/", nil)

	userHandler := new(UserHandler)
	handlers := []func(w http.ResponseWriter, r *http.Request){
		userHandler.ActivateFeature,
		userHandler.TrackEvent,
	}

	for _, handler := range handlers {
		rec := httptest.NewRecorder()
		http.HandlerFunc(handler).ServeHTTP(rec, req)

		// Unmarshal response
		var actual models.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &actual)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Equal(t, models.ErrorResponse{Error: "optlyClient not available"}, actual)
	}
}

func TestUserMissingOptlyCtx(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req := httptest.NewRequest("POST", "/", nil)
	mw := new(UserMW)

	userHandler := new(UserHandler)
	handlers := []func(w http.ResponseWriter, r *http.Request){
		userHandler.ActivateFeature,
		userHandler.TrackEvent,
	}

	for _, handler := range handlers {
		rec := httptest.NewRecorder()
		mw.ClientCtx(http.HandlerFunc(handler)).ServeHTTP(rec, req)

		// Unmarshal response
		var actual models.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &actual)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Equal(t, models.ErrorResponse{Error: "optlyContext not available"}, actual)
	}
}
