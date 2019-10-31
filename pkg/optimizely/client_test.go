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

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
	optlyClient  *OptlyClient
	optlyContext *OptlyContext
	testClient   *optimizelytest.TestClient
}

func (suite *ClientTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	suite.testClient = testClient
	suite.optlyClient = &OptlyClient{testClient.OptimizelyClient, nil}
	suite.optlyContext = NewContext("userId", make(map[string]interface{}))
}

func (suite *ClientTestSuite) TestListFeatures() {
	suite.testClient.AddFeature(entities.Feature{Key: "k1"})
	suite.testClient.AddFeature(entities.Feature{Key: "k2"})
	features, err := suite.optlyClient.ListFeatures()
	suite.NoError(err)
	suite.Equal(2, len(features))
}

func (suite *ClientTestSuite) TestGetFeature() {
	suite.testClient.AddFeature(entities.Feature{Key: "k1"})
	actual, err := suite.optlyClient.GetFeature("k1")
	suite.NoError(err)
	suite.Equal(actual, entities.Feature{Key: "k1"})
}

func (suite *ClientTestSuite) TestGetNonExistentFeature() {
	enabled, variationMap, err := suite.optlyClient.GetFeatureWithContext("DNE", suite.optlyContext)

	suite.False(enabled)
	suite.Equal(0, len(variationMap))
	suite.NoError(err) // TODO should this error?
}

func (suite *ClientTestSuite) TestGetBasicFeature() {
	basicFeature := entities.Feature{Key: "basic"}
	suite.testClient.AddFeatureRollout(basicFeature)
	enabled, variableMap, err := suite.optlyClient.GetFeatureWithContext("basic", suite.optlyContext)

	suite.NoError(err)
	suite.True(enabled)
	suite.Equal(0, len(variableMap))
}

func (suite *ClientTestSuite) TestGetAdvancedFeature() {
	var1 := entities.Variable{Key: "var1", DefaultValue: "val1"}
	var2 := entities.Variable{Key: "var2", DefaultValue: "val2"}
	advancedFeature := entities.Feature{
		Key: "advanced",
		VariableMap: map[string]entities.Variable{
			"var1": var1,
			"var2": var2,
		},
	}

	suite.testClient.AddFeatureRollout(advancedFeature)
	enabled, variableMap, err := suite.optlyClient.GetFeatureWithContext("advanced", suite.optlyContext)

	suite.NoError(err)
	suite.True(enabled)
	suite.Equal(2, len(variableMap))
}

func (suite *ClientTestSuite) TestTrackEventWithContext() {
	eventKey := "eventKey"
	suite.testClient.AddEvent(entities.Event{Key: eventKey})
	tags := map[string]interface{}{"tag": "value"}
	err := suite.optlyClient.TrackEventWithContext(eventKey, suite.optlyContext, tags)
	suite.NoError(err)

	events := suite.testClient.GetProcessedEvents()
	suite.Equal(1, len(events))

	actual := events[0]
	suite.Equal(eventKey, actual.Conversion.Key)
	suite.Equal("userId", actual.VisitorID)
	suite.Equal(tags, actual.Conversion.Tags)
}

func (suite *ClientTestSuite) TestTrackEventWithContextError() {
	err := suite.optlyClient.TrackEventWithContext("missing-key", suite.optlyContext, map[string]interface{}{})
	suite.NoError(err) // TODO Should this error?
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
