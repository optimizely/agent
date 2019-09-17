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
}

func (suite *ContextTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	testClient.ProjectConfig.FeatureMap["basic"] = entities.Feature{
		Key: "basic",
		
	}
	testClient.ProjectConfig.FeatureMap["withVars"] = entities.Feature{Key: "withVars"}

	suite.context =  NewContextWithOptimizely("userId", make(map[string]interface{}), testClient.OptimizelyClient)
}

func (suite *ContextTestSuite) TestGetNonExistentFeature() {
	_, _, err := suite.context.GetFeature("DNE")
	if !suite.Error(err) {
		suite.Equal(fmt.Errorf("Feature with key DNE not found"), err)
	}
}

func (suite *ContextTestSuite) TestGetBasicFeature() {
	enabled, variableMap, err := suite.context.GetFeature("basic")
	suite.Nil(err)
	suite.True(enabled)
	suite.Equal(0, len(variableMap))
}

func (suite *ContextTestSuite) TestGetFeatureWithVariables() {
	enabled, variableMap, err := suite.context.GetFeature("withVars")
	suite.Nil(err)
	suite.True(enabled)
	suite.Equal(2, len(variableMap))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestContextTestSuite(t *testing.T) {
	suite.Run(t, new(ContextTestSuite))
}
