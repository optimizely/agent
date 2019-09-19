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

// Package optimizely //
package optimizely

import (
	"testing"

	"github.com/optimizely/sidedoor/pkg/optimizelytest"

	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
	client *OptlyClient
	testClient *optimizelytest.TestClient
}

func (suite *ClientTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	suite.testClient = testClient
	suite.client = &OptlyClient{testClient.OptimizelyClient}
}

func (suite *ClientTestSuite) TestListFeatures() {
	suite.testClient.AddFeature(entities.Feature{Key: "k1"})
	suite.testClient.AddFeature(entities.Feature{Key: "k2"})
	features, err := suite.client.ListFeatures()
	suite.NoError(err)
	suite.Equal(2, len(features))
}

func (suite *ClientTestSuite) TestGetFeature() {
	suite.testClient.AddFeature(entities.Feature{Key: "k1"})
	actual, err := suite.client.GetFeature("k1")
	suite.NoError(err)
	suite.Equal(actual, entities.Feature{Key: "k1"})
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}