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

	"github.com/optimizely/sidedoor/pkg/optimizely"
	"github.com/optimizely/sidedoor/pkg/optimizelytest"

	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/suite"
)

type FeatureTestSuite struct {
	suite.Suite
	optlyClient *optimizely.OptlyClient
	testClient  *optimizelytest.TestClient
}

// Setup Middleware
func (suite *FeatureTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	suite.testClient = testClient
	suite.optlyClient = &optimizely.OptlyClient{testClient.OptimizelyClient}
}

func (suite *FeatureTestSuite) TestListFeatures() {
	feature := entities.Feature{Key: "one"}
	suite.testClient.AddFeature(feature)

	req, err := http.NewRequest("GET", "/features", nil)
	suite.Nil(err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ListFeatures)

	// Add appropriate context
	ctx := context.WithValue(req.Context(), "optlyClient", suite.optlyClient)
	handler.ServeHTTP(rr, req.WithContext(ctx))

	suite.Equal(http.StatusOK, rr.Code)

	// Unmarshal response
	var actual []entities.Feature
	err = json.Unmarshal(rr.Body.Bytes(), &actual)
	suite.NoError(err)

	suite.Equal(1, len(actual))
	suite.Equal(feature, actual[0])
}

func (suite *FeatureTestSuite) TestListFeaturesMissingCtx() {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/features", nil)
	suite.Nil(err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ListFeatures)
	handler.ServeHTTP(rr, req)

	suite.Equal(http.StatusUnprocessableEntity, rr.Code)
	suite.Equal("OptlyClient not available\n", rr.Body.String())
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFeatureTestSuite(t *testing.T) {
	suite.Run(t, new(FeatureTestSuite))
}
