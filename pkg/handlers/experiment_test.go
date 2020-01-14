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

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"

	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/entities"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ExperimentTestSuite struct {
	suite.Suite
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

type OptlyMWExperiment struct {
	optlyClient *optimizely.OptlyClient
}

func (o *OptlyMWExperiment) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, o.optlyClient)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (o *OptlyMWExperiment) ExperimentCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		experimentKey := chi.URLParam(r, "experimentKey")
		experiment := config.OptimizelyExperiment{
			Key: experimentKey,
			VariationsMap: map[string]config.OptimizelyVariation{},
		}
		ctx := context.WithValue(r.Context(), middleware.OptlyExperimentKey, &experiment)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Setup Mux
func (suite *ExperimentTestSuite) SetupTest() {

	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    nil,
		ForcedVariations: testClient.ForcedVariations,
	}

	mux := chi.NewMux()
	experimentAPI := new(ExperimentHandler)
	optlyMW := &OptlyMWExperiment{optlyClient}

	mux.Use(optlyMW.ClientCtx)
	mux.Get("/experiments", experimentAPI.ListExperiments)
	mux.With(optlyMW.ExperimentCtx).Get("/experiments/{experimentKey}", experimentAPI.GetExperiment)

	suite.mux = mux
	suite.tc = testClient
}

func (suite *ExperimentTestSuite) TestListExperiments() {
	experiment := config.OptimizelyExperiment{Key: "one", ID: "1", VariationsMap: map[string]config.OptimizelyVariation{}}

	suite.tc.AddExperiment("one", []entities.Variation{})

	// Add appropriate context
	req := httptest.NewRequest("GET", "/experiments", nil)
	rec := httptest.NewRecorder()

	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual []config.OptimizelyExperiment
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	suite.Equal(1, len(actual))
	suite.Equal(experiment, actual[0])
}

func (suite *ExperimentTestSuite) TestGetExperiment() {
	experiment := config.OptimizelyExperiment{Key: "one", VariationsMap: map[string]config.OptimizelyVariation{}}

	suite.tc.AddExperiment("one", []entities.Variation{})

	req := httptest.NewRequest("GET", "/experiments/one", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual config.OptimizelyExperiment
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	suite.Equal(experiment, actual)
}

func (suite *ExperimentTestSuite) TestGetExperimentsMissingExperiments() {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/experiments/experiment-404", nil)
	suite.Nil(err)

	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)

	suite.Equal(http.StatusInternalServerError, rec.Code)
	// Unmarshal response
	var actual ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	suite.Equal(ErrorResponse{Error: "experiment not available"}, actual)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(ExperimentTestSuite))
}
func TestExperimentMissingClientCtx(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req := httptest.NewRequest("GET", "/", nil)

	experimentHandler := new(ExperimentHandler)
	handlers := []func(w http.ResponseWriter, r *http.Request){
		experimentHandler.ListExperiments,
		experimentHandler.GetExperiment,
	}

	for _, handler := range handlers {
		rec := httptest.NewRecorder()
		http.HandlerFunc(handler).ServeHTTP(rec, req)

		// Unmarshal response
		var actual ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &actual)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Equal(t, ErrorResponse{Error: "optlyClient not available"}, actual)
	}
}

