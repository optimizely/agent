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

	// "github.com/optimizely/sidedoor/pkg/api/middleware"
	"github.com/optimizely/sidedoor/pkg/optimizely"
	"github.com/optimizely/sidedoor/pkg/optimizelytest"

	"github.com/go-chi/chi"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FeatureTestSuite struct {
	suite.Suite
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

// Setup Mux
func (suite *FeatureTestSuite) SetupTest() {

	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{testClient.OptimizelyClient, nil}

	optimizelymw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO figure out why using the middleware constant doesn't work :/
			ctx := context.WithValue(r.Context(), "optlyClient", optlyClient)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	mux := chi.NewMux()
	featureAPI := new(FeatureHandler)

	mux.Use(optimizelymw)
	mux.Get("/features", featureAPI.ListFeatures)
	mux.Get("/features/{featureKey}", featureAPI.GetFeature)

	suite.mux = mux
	suite.tc = testClient
}

func (suite *FeatureTestSuite) TestListFeatures() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeature(feature)

	// Add appropriate context
	req, err := http.NewRequest("GET", "/features", nil)
	suite.Nil(err)

	rr := httptest.NewRecorder()
	suite.mux.ServeHTTP(rr, req)

	suite.Equal(http.StatusOK, rr.Code)

	suite.NoError(err)

	// Unmarshal response
	var actual []entities.Feature
	err = json.Unmarshal(rr.Body.Bytes(), &actual)
	suite.NoError(err)

	suite.Equal(1, len(actual))
	suite.Equal(feature, actual[0])
}

func (suite *FeatureTestSuite) TestGetFeature() {
	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeature(feature)

	req, err := http.NewRequest("GET", "/features/one", nil)
	suite.Nil(err)

	rr := httptest.NewRecorder()
	suite.mux.ServeHTTP(rr, req)

	suite.Equal(http.StatusOK, rr.Code)

	// Unmarshal response
	var actual entities.Feature
	err = json.Unmarshal(rr.Body.Bytes(), &actual)
	suite.NoError(err)

	suite.Equal(feature, actual)
}

func (suite *FeatureTestSuite) TestGetFeaturesMissingFeature() {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/features/dne", nil)
	suite.Nil(err)

	rr := httptest.NewRecorder()
	suite.mux.ServeHTTP(rr, req)

	suite.Equal(http.StatusInternalServerError, rr.Code)
	// TODO create an actual error response model
	//suite.Equal("{\"error\":\"Feature with key one not found\"}\n", rr.Body.String())
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFeatureTestSuite(t *testing.T) {
	suite.Run(t, new(FeatureTestSuite))
}

func TestListFeaturesMissingCtx(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/features", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	http.HandlerFunc(new(FeatureHandler).ListFeatures).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Equal(t, "OptlyClient not available\n", rr.Body.String())
}

func TestGetFeatureMissingCtx(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/features/one", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	http.HandlerFunc(new(FeatureHandler).GetFeature).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Equal(t, "OptlyClient not available\n", rr.Body.String())
}
