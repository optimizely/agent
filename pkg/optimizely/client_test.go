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
	"errors"
	"fmt"
	"testing"

	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/stretchr/testify/assert"

	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"

	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
	optlyClient *OptlyClient
	userContext entities.UserContext
	testClient  *optimizelytest.TestClient
}

func (suite *ClientTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	suite.testClient = testClient
	suite.optlyClient = &OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    &MockConfigManager{config: testClient.ProjectConfig},
		ForcedVariations: testClient.ForcedVariations}
	suite.userContext = entities.UserContext{
		ID:         "userId",
		Attributes: make(map[string]interface{}),
	}
}

func (suite *ClientTestSuite) TearDownTest() {
	suite.optlyClient.Close()
}

func (suite *ClientTestSuite) TestTrackEvent() {
	eventKey := "eventKey"
	suite.testClient.AddEvent(entities.Event{Key: eventKey})
	tags := map[string]interface{}{"tag": "value"}
	err := suite.optlyClient.TrackEvent(eventKey, suite.userContext, tags)
	suite.NoError(err)

	events := suite.testClient.GetProcessedEvents()
	suite.Equal(1, len(events))

	actual := events[0]
	suite.Equal(eventKey, actual.Conversion.Key)
	suite.Equal("userId", actual.VisitorID)
	suite.Equal(tags, actual.Conversion.Tags)
}

func (suite *ClientTestSuite) TestSetForcedVariationSuccess() {
	feature := entities.Feature{Key: "my_feat"}
	suite.testClient.ProjectConfig.AddMultiVariationFeatureTest(feature, "disabled_var", "enabled_var")
	featureExp := suite.testClient.ProjectConfig.FeatureMap["my_feat"].FeatureExperiments[0]
	actual, err := suite.optlyClient.SetForcedVariation(featureExp.Key, "userId", "enabled_var")
	suite.NoError(err)
	suite.Equal(actual.UserID, "userId")
	suite.Equal(actual.ExperimentKey, featureExp.Key)
	suite.Equal(actual.VariationKey, "enabled_var")
	suite.Empty(actual.PrevVariationKey)
	isEnabled, _ := suite.optlyClient.IsFeatureEnabled("my_feat", suite.userContext)
	suite.True(isEnabled)
}

func (suite *ClientTestSuite) TestSetForcedVariationAlreadySet() {
	feature := entities.Feature{Key: "my_feat"}
	suite.testClient.ProjectConfig.AddMultiVariationFeatureTest(feature, "disabled_var", "enabled_var")
	featureExp := suite.testClient.ProjectConfig.FeatureMap["my_feat"].FeatureExperiments[0]
	suite.optlyClient.SetForcedVariation(featureExp.Key, "userId", "enabled_var")
	// Set the same forced variation again
	actual, err := suite.optlyClient.SetForcedVariation(featureExp.Key, "userId", "enabled_var")
	suite.NoError(err)
	suite.Equal(actual.UserID, "userId")
	suite.Equal(actual.ExperimentKey, featureExp.Key)
	suite.Equal(actual.VariationKey, "enabled_var")
	suite.Equal(actual.PrevVariationKey, "enabled_var")
	suite.Equal([]string{"updating previous override"}, actual.Messages)
	isEnabled, _ := suite.optlyClient.IsFeatureEnabled("my_feat", suite.userContext)
	suite.True(isEnabled)
}

func (suite *ClientTestSuite) TestSetForcedVariationDifferentVariation() {
	feature := entities.Feature{Key: "my_feat"}
	suite.testClient.ProjectConfig.AddMultiVariationFeatureTest(feature, "disabled_var", "enabled_var")
	featureExp := suite.testClient.ProjectConfig.FeatureMap["my_feat"].FeatureExperiments[0]
	suite.optlyClient.SetForcedVariation(featureExp.Key, "userId", "disabled_var")
	// Set a different forced variation
	actual, err := suite.optlyClient.SetForcedVariation(featureExp.Key, "userId", "enabled_var")
	suite.NoError(err)
	suite.Equal(actual.UserID, "userId")
	suite.Equal(actual.ExperimentKey, featureExp.Key)
	suite.Equal(actual.VariationKey, "enabled_var")
	suite.Equal(actual.PrevVariationKey, "disabled_var")
	suite.Equal([]string{"updating previous override"}, actual.Messages)
	isEnabled, _ := suite.optlyClient.IsFeatureEnabled("my_feat", suite.userContext)
	suite.True(isEnabled)
}

func (suite *ClientTestSuite) TestSetForcedVariationMissingExperiment() {
	actual, err := suite.optlyClient.SetForcedVariation("does-not-exist", "userId", "disabled_var")
	suite.NoError(err)
	suite.Equal(actual.UserID, "userId")
	suite.Equal(actual.ExperimentKey, "does-not-exist")
	suite.Equal(actual.VariationKey, "disabled_var")
	suite.Empty(actual.PrevVariationKey)
	suite.Equal([]string{"experimentKey not found in configuration"}, actual.Messages)
}

func (suite *ClientTestSuite) TestSetForcedVariationMissingVariation() {
	suite.testClient.AddExperimentWithVariations("my_exp", "my_var")
	actual, err := suite.optlyClient.SetForcedVariation("my_exp", "userId", "does-not-exist")
	suite.NoError(err)
	suite.Equal(actual.UserID, "userId")
	suite.Equal(actual.ExperimentKey, "my_exp")
	suite.Equal(actual.VariationKey, "does-not-exist")
	suite.Empty(actual.PrevVariationKey)
	suite.Equal([]string{"variationKey not found in configuration"}, actual.Messages)
}

func (suite *ClientTestSuite) TestRemoveForcedVariation() {
	feature := entities.Feature{Key: "my_feat"}
	suite.testClient.ProjectConfig.AddMultiVariationFeatureTest(feature, "disabled_var", "enabled_var")
	featureExp := suite.testClient.ProjectConfig.FeatureMap["my_feat"].FeatureExperiments[0]
	suite.optlyClient.SetForcedVariation(featureExp.Key, "userId", "enabled_var")
	err := suite.optlyClient.RemoveForcedVariation(featureExp.Key, "userId")
	suite.NoError(err)
	isEnabled, _ := suite.optlyClient.IsFeatureEnabled("my_feat", suite.userContext)
	suite.False(isEnabled)
}

func (suite *ClientTestSuite) TestSetForcedVariationABTestSuccess() {
	suite.testClient.ProjectConfig.AddMultiVariationABTest("my_exp", "var_1", "var_2")
	suite.optlyClient.SetForcedVariation("my_exp", "userId", "var_1")
	variation, err := suite.optlyClient.Activate("my_exp", suite.userContext)
	suite.NoError(err)
	suite.Equal("var_1", variation)
}

func (suite *ClientTestSuite) TestRemoveForcedVariationABTest() {
	suite.testClient.ProjectConfig.AddMultiVariationABTest("my_exp", "var_1", "var_2")
	suite.optlyClient.SetForcedVariation("my_exp", "userId", "var_1")
	err := suite.optlyClient.RemoveForcedVariation("my_exp", "userId")
	suite.NoError(err)
	variation, err := suite.optlyClient.Activate("my_exp", suite.userContext)
	suite.NoError(err)
	suite.Equal("var_2", variation)
}

func (suite *ClientTestSuite) TestActivateFeature() {
	var1 := entities.Variable{Key: "var1", DefaultValue: "val1"}
	var2 := entities.Variable{Key: "var2", DefaultValue: "val2"}
	advancedFeature := entities.Feature{
		Key: "advanced",
		VariableMap: map[string]entities.Variable{
			"var1": var1,
			"var2": var2,
		},
	}

	suite.testClient.AddFeatureTest(advancedFeature)
	feature := suite.testClient.OptimizelyClient.GetOptimizelyConfig().FeaturesMap["advanced"]

	expected := &Decision{
		FeatureKey: "advanced",
		Type:       "feature",
		Variables: map[string]interface{}{
			"var1": "val1",
			"var2": "val2",
		},
		Enabled: true,
	}

	// Response should be the same regardless of the flag
	for _, flag := range []bool{true, false} {
		actual, err := suite.optlyClient.ActivateFeature(feature.Key, entities.UserContext{ID: "testUser"}, flag)
		suite.NoError(err)
		suite.Equal(expected, actual)
	}

	// Only one event should have been triggered
	suite.Equal(1, len(suite.testClient.GetProcessedEvents()))
}

func (suite *ClientTestSuite) TestActivateExperiment() {
	testExperimentKey := "testExperiment1"
	testVariation := suite.testClient.ProjectConfig.CreateVariation("variationA")
	suite.testClient.AddExperiment(testExperimentKey, []entities.Variation{testVariation})
	experiment := suite.testClient.OptimizelyClient.GetOptimizelyConfig().ExperimentsMap["testExperiment1"]

	expected := &Decision{
		ExperimentKey: "testExperiment1",
		VariationKey:  "variationA",
		Type:          "experiment",
		Variables:     map[string]interface{}{},
		Enabled:       true,
	}

	// Response should be the same regardless of the flag
	for _, flag := range []bool{true, false} {
		actual, err := suite.optlyClient.ActivateExperiment(experiment.Key, entities.UserContext{ID: "testUser"}, flag)
		suite.NoError(err)
		suite.Equal(expected, actual)
	}

	// Only one event should have been triggered
	suite.Equal(1, len(suite.testClient.GetProcessedEvents()))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

type ErrorConfigManager struct{}

func (e ErrorConfigManager) RemoveOnProjectConfigUpdate(id int) error {
	panic("implement me")
}

func (e ErrorConfigManager) OnProjectConfigUpdate(callback func(notification.ProjectConfigUpdateNotification)) (int, error) {
	panic("implement me")
}

func (e ErrorConfigManager) GetConfig() (config.ProjectConfig, error) {
	return nil, fmt.Errorf("config error")
}

func (e ErrorConfigManager) GetOptimizelyConfig() *config.OptimizelyConfig {
	panic("implement me")
}

func (e ErrorConfigManager) SyncConfig() {
	panic("implement me")
}

type MockConfigManager struct {
	config config.ProjectConfig
}

func (m MockConfigManager) GetConfig() (config.ProjectConfig, error) {
	return m.config, nil
}

func (m MockConfigManager) GetOptimizelyConfig() *config.OptimizelyConfig {
	panic("implement me")
}

func (m MockConfigManager) SyncConfig() {
	panic("implement me")
}

func TestTrackErrorConfigManager(t *testing.T) {
	testClient := optimizelytest.NewClient()
	optlyClient := &OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    ErrorConfigManager{},
		ForcedVariations: testClient.ForcedVariations,
	}

	err := optlyClient.TrackEvent("something", entities.UserContext{}, map[string]interface{}{})
	assert.EqualError(t, err, "config error")
}

func TestTrackErrorClient(t *testing.T) {
	// Construct an OptimizelyClient with an erroring config manager
	factory := client.OptimizelyFactory{}
	oClient, _ := factory.Client(
		client.WithConfigManager(ErrorConfigManager{}),
	)

	// Construct a valid config manager as part of the OptlyClient wrapper
	testConfig := optimizelytest.NewConfig()
	eventKey := "test-event"
	event := entities.Event{Key: eventKey}
	testConfig.AddEvent(event)

	optlyClient := &OptlyClient{
		OptimizelyClient: oClient,
		ConfigManager:    &MockConfigManager{config: testConfig},
		ForcedVariations: nil,
	}

	err := optlyClient.TrackEvent("something", entities.UserContext{}, map[string]interface{}{})
	assert.EqualError(t, err, `eventKey: "something" not found`)
	assert.True(t, errors.Is(err, ErrEntityNotFound))
}
