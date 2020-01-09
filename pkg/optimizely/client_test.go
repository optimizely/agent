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

	"github.com/optimizely/sidedoor/pkg/optimizely/optimizelytest"

	"github.com/optimizely/go-sdk/pkg/config"
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
	suite.optlyClient = &OptlyClient{testClient.OptimizelyClient, nil, testClient.ForcedVariations}
	suite.optlyContext = NewContext("userId", make(map[string]interface{}))
}

func (suite *ClientTestSuite) TearDownTest() {
	suite.optlyClient.Close()
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
	suite.Equal(actual, config.OptimizelyFeature{Key: "k1", ExperimentsMap: map[string]config.OptimizelyExperiment{},
		VariablesMap: map[string]config.OptimizelyVariable{}})
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

func (suite *ClientTestSuite) TestGetExperiment() {
	testExperimentKey := "testExperiment1"
	testVariation := suite.testClient.ProjectConfig.CreateVariation("variationA")
	suite.testClient.AddExperiment(testExperimentKey, []entities.Variation{testVariation})
	experiment, err := suite.optlyClient.GetExperiment("testExperiment1")
	suite.Equal(testExperimentKey, experiment.Key)
	suite.NoError(err)
}

func (suite *ClientTestSuite) TestListExperiments() {
	suite.testClient.AddExperiment("k1", []entities.Variation{})
	suite.testClient.AddExperiment("k2", []entities.Variation{})
	experiments, err := suite.optlyClient.ListExperiments()
	suite.NoError(err)
	suite.Equal(2, len(experiments))
}

func (suite *ClientTestSuite) TestGetExperimentVariation() {
	testExperimentKey := "testExperiment1"
	testVariation := suite.testClient.ProjectConfig.CreateVariation("variationA")
	suite.testClient.AddExperiment(testExperimentKey, []entities.Variation{testVariation})
	optiConfigVariation := suite.testClient.ProjectConfig.ConvertVariation(testVariation)
	variation, err := suite.optlyClient.GetExperimentVariation(testExperimentKey, false, suite.optlyContext)
	suite.Equal(optiConfigVariation, variation)
	suite.NoError(err)
}

func (suite *ClientTestSuite) TestGetExperimentVariationWithActivation() {
	testExperimentKey := "testExperiment1"
	testVariation := suite.testClient.ProjectConfig.CreateVariation("variationA")
	suite.testClient.AddExperiment(testExperimentKey, []entities.Variation{testVariation})
	optiConfigVariation := suite.testClient.ProjectConfig.ConvertVariation(testVariation)
	variation, err := suite.optlyClient.GetExperimentVariation(testExperimentKey, true, suite.optlyContext)
	suite.Equal(optiConfigVariation, variation)
	suite.NoError(err)
}

func (suite *ClientTestSuite) TestGetExperimentVariationNonExistentExperiment() {
	variation, err := suite.optlyClient.GetExperimentVariation("non_existent_experiment", false, suite.optlyContext)
	suite.Equal("", variation.ID) // empty variation
	suite.NoError(err)
}

func (suite *ClientTestSuite) TestSetForcedVariationSuccess() {
	feature := entities.Feature{Key: "my_feat"}
	suite.testClient.ProjectConfig.AddMultiVariationFeatureTest(feature, "disabled_var", "enabled_var")
	featureExp := suite.testClient.ProjectConfig.FeatureMap["my_feat"].FeatureExperiments[0]
	wasSet, err := suite.optlyClient.SetForcedVariation(featureExp.Key, "userId", "enabled_var")
	suite.NoError(err)
	suite.True(wasSet)
	isEnabled, _ := suite.optlyClient.IsFeatureEnabled("my_feat", *suite.optlyContext.UserContext)
	suite.True(isEnabled)
}

func (suite *ClientTestSuite) TestSetForcedVariationAlreadySet() {
	feature := entities.Feature{Key: "my_feat"}
	suite.testClient.ProjectConfig.AddMultiVariationFeatureTest(feature, "disabled_var", "enabled_var")
	featureExp := suite.testClient.ProjectConfig.FeatureMap["my_feat"].FeatureExperiments[0]
	suite.optlyClient.SetForcedVariation(featureExp.Key, "userId", "enabled_var")
	// Set the same forced variation again
	wasSet, err := suite.optlyClient.SetForcedVariation(featureExp.Key, "userId", "enabled_var")
	suite.NoError(err)
	suite.False(wasSet)
	isEnabled, _ := suite.optlyClient.IsFeatureEnabled("my_feat", *suite.optlyContext.UserContext)
	suite.True(isEnabled)
}

func (suite *ClientTestSuite) TestSetForcedVariationDifferentVariation() {
	feature := entities.Feature{Key: "my_feat"}
	suite.testClient.ProjectConfig.AddMultiVariationFeatureTest(feature, "disabled_var", "enabled_var")
	featureExp := suite.testClient.ProjectConfig.FeatureMap["my_feat"].FeatureExperiments[0]
	suite.optlyClient.SetForcedVariation(featureExp.Key, "userId", "disabled_var")
	// Set a different forced variation
	wasSet, err := suite.optlyClient.SetForcedVariation(featureExp.Key, "userId", "enabled_var")
	suite.NoError(err)
	suite.True(wasSet)
	isEnabled, _ := suite.optlyClient.IsFeatureEnabled("my_feat", *suite.optlyContext.UserContext)
	suite.True(isEnabled)
}

func (suite *ClientTestSuite) TestRemoveForcedVariation() {
	feature := entities.Feature{Key: "my_feat"}
	suite.testClient.ProjectConfig.AddMultiVariationFeatureTest(feature, "disabled_var", "enabled_var")
	featureExp := suite.testClient.ProjectConfig.FeatureMap["my_feat"].FeatureExperiments[0]
	suite.optlyClient.SetForcedVariation(featureExp.Key, "userId", "enabled_var")
	err := suite.optlyClient.RemoveForcedVariation(featureExp.Key, "userId")
	suite.NoError(err)
	isEnabled, _ := suite.optlyClient.IsFeatureEnabled("my_feat", *suite.optlyContext.UserContext)
	suite.False(isEnabled)
}

func (suite *ClientTestSuite) TestSetForcedVariationABTestSuccess() {
	suite.testClient.ProjectConfig.AddMultiVariationABTest("my_exp", "var_1", "var_2")
	suite.optlyClient.SetForcedVariation("my_exp", "userId", "var_1")
	variation, err := suite.optlyClient.Activate("my_exp", *suite.optlyContext.UserContext)
	suite.NoError(err)
	suite.Equal("var_1", variation)
}

func (suite *ClientTestSuite) TestRemoveForcedVariationABTest() {
	suite.testClient.ProjectConfig.AddMultiVariationABTest("my_exp", "var_1", "var_2")
	suite.optlyClient.SetForcedVariation("my_exp", "userId", "var_1")
	err := suite.optlyClient.RemoveForcedVariation("my_exp", "userId")
	suite.NoError(err)
	variation, err := suite.optlyClient.Activate("my_exp", *suite.optlyContext.UserContext)
	suite.NoError(err)
	suite.Equal("var_2", variation)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
