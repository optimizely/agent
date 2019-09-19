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
	"fmt"
	"testing"

	"github.com/optimizely/sidedoor/pkg/optimizelytest"

	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/suite"
)

type ContextTestSuite struct {
	suite.Suite
	context *Context
	testClient *optimizelytest.TestClient
}

func (suite *ContextTestSuite) SetupTest() {
	suite.testClient = optimizelytest.NewClient()
	suite.context =  NewContextWithOptimizely("userId", make(map[string]interface{}), suite.testClient.OptimizelyClient)
}

func (suite *ContextTestSuite) TestGetNonExistentFeature() {
	_, _, err := suite.context.GetFeature("DNE")
	if !suite.Error(err) {
		suite.Equal(fmt.Errorf("Feature with key DNE not found"), err)
	}
}

func (suite *ContextTestSuite) TestGetBasicFeature() {
	basicFeature := entities.Feature{Key: "basic"}
	suite.testClient.AddFeatureRollout(basicFeature)
	enabled, variableMap, err := suite.context.GetFeature("basic")

	suite.NoError(err)
	suite.True(enabled)
	suite.Equal(0, len(variableMap))
}

func (suite *ContextTestSuite) TestGetAdvancedFeature() {
	var1 := entities.Variable{Key: "var1", DefaultValue: "val1"}
	var2 := entities.Variable{Key: "var2", DefaultValue: "val2"}
	advancedFeature := entities.Feature{
		Key: "advanced",
		Variables: []entities.Variable{var1, var2},
	}

	suite.testClient.AddFeatureRollout(advancedFeature)
	enabled, variableMap, err := suite.context.GetFeature("advanced")

	suite.NoError(err)
	suite.True(enabled)
	suite.Equal(2, len(variableMap))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestContextTestSuite(t *testing.T) {
	suite.Run(t, new(ContextTestSuite))
}
