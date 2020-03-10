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

// Package optimizelytest //
package optimizelytest

import (
	"errors"
	"fmt"
	"strconv"

	optlyconfig "github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/entities"
)

// TestProjectConfig is a project config backed by a datafile
type TestProjectConfig struct {
	AccountID            string
	ProjectID            string
	Revision             string
	ExperimentKeyToIDMap map[string]string
	AudienceMap          map[string]entities.Audience
	AttributeMap         map[string]entities.Attribute
	EventMap             map[string]entities.Event
	AttributeKeyToIDMap  map[string]string
	ExperimentMap        map[string]entities.Experiment
	FeatureMap           map[string]entities.Feature
	GroupMap             map[string]entities.Group
	RolloutMap           map[string]entities.Rollout
	AnonymizeIP          bool
	BotFiltering         bool
	nextID               int
}

// GetProjectID returns projectID
func (c *TestProjectConfig) GetProjectID() string {
	return c.ProjectID
}

// GetRevision returns revision
func (c *TestProjectConfig) GetRevision() string {
	return c.Revision
}

// GetAccountID returns accountID
func (c *TestProjectConfig) GetAccountID() string {
	return c.AccountID
}

// GetAnonymizeIP returns anonymizeIP
func (c *TestProjectConfig) GetAnonymizeIP() bool {
	return c.AnonymizeIP
}

// GetAttributeID returns attributeID
func (c *TestProjectConfig) GetAttributeID(key string) string {
	return c.AttributeKeyToIDMap[key]
}

// GetBotFiltering returns GetBotFiltering
func (c *TestProjectConfig) GetBotFiltering() bool {
	return c.BotFiltering
}

// GetEventByKey returns the event with the given key
func (c *TestProjectConfig) GetEventByKey(eventKey string) (entities.Event, error) {
	if event, ok := c.EventMap[eventKey]; ok {
		return event, nil
	}

	errMessage := fmt.Sprintf("Event with key %s not found", eventKey)
	return entities.Event{}, errors.New(errMessage)
}

// GetFeatureByKey returns the feature with the given key
func (c *TestProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	if feature, ok := c.FeatureMap[featureKey]; ok {
		return feature, nil
	}

	errMessage := fmt.Sprintf("Feature with key %s not found", featureKey)
	return entities.Feature{}, errors.New(errMessage)
}

// GetVariableByKey returns the featureVariable with the given key
func (c *TestProjectConfig) GetVariableByKey(featureKey, variableKey string) (entities.Variable, error) {

	var variable entities.Variable
	var err = fmt.Errorf("variable with key %s not found", featureKey)
	if feature, ok := c.FeatureMap[featureKey]; ok {
		if variable, ok = feature.VariableMap[featureKey]; ok {
			return variable, nil
		}
	}
	return variable, err
}

// GetAttributeByKey returns the attribute with the given key
func (c *TestProjectConfig) GetAttributeByKey(key string) (entities.Attribute, error) {
	if attributeID, ok := c.AttributeKeyToIDMap[key]; ok {
		return c.AttributeMap[attributeID], nil
	}

	errMessage := fmt.Sprintf(`Attribute with key "%s" not found`, key)
	return entities.Attribute{}, errors.New(errMessage)
}

// GetFeatureList returns an array of all the features
func (c *TestProjectConfig) GetFeatureList() (featureList []entities.Feature) {
	for _, feature := range c.FeatureMap {
		featureList = append(featureList, feature)
	}
	return featureList
}

// GetExperimentList returns an array of all the experiments
func (c *TestProjectConfig) GetExperimentList() (experimentList []entities.Experiment) {
	for _, experiment := range c.ExperimentMap {
		experimentList = append(experimentList, experiment)
	}
	return experimentList
}

// GetAudienceByID returns the audience with the given ID
func (c *TestProjectConfig) GetAudienceByID(audienceID string) (entities.Audience, error) {
	if audience, ok := c.AudienceMap[audienceID]; ok {
		return audience, nil
	}

	errMessage := fmt.Sprintf(`Audience with ID "%s" not found`, audienceID)
	return entities.Audience{}, errors.New(errMessage)
}

// GetAudienceMap returns the audience map
func (c *TestProjectConfig) GetAudienceMap() map[string]entities.Audience {
	return c.AudienceMap
}

// GetExperimentByKey returns the experiment with the given key
func (c *TestProjectConfig) GetExperimentByKey(experimentKey string) (entities.Experiment, error) {
	if experimentID, ok := c.ExperimentKeyToIDMap[experimentKey]; ok {
		experiment := c.ExperimentMap[experimentID]
		return experiment, nil
	}

	errMessage := fmt.Sprintf(`Experiment with key "%s" not found`, experimentKey)
	return entities.Experiment{}, errors.New(errMessage)
}

// GetGroupByID returns the group with the given ID
func (c *TestProjectConfig) GetGroupByID(groupID string) (entities.Group, error) {
	if group, ok := c.GroupMap[groupID]; ok {
		return group, nil
	}

	errMessage := fmt.Sprintf(`Group with ID "%s" not found`, groupID)
	return entities.Group{}, errors.New(errMessage)
}

// AddEvent adds the event to the EventMap
func (c *TestProjectConfig) AddEvent(e entities.Event) *TestProjectConfig {
	c.EventMap[e.Key] = e
	return c
}

// AddFeature adds the feature to the FeatureMap
func (c *TestProjectConfig) AddFeature(f entities.Feature) *TestProjectConfig {
	c.FeatureMap[f.Key] = f
	return c
}

// AddFeatureTest adds the feature and supporting entities to complete the feature test modeling
func (c *TestProjectConfig) AddFeatureTest(f entities.Feature) *TestProjectConfig {
	experimentID := c.getNextID()
	variationID := c.getNextID()
	layerID := c.getNextID()

	variation := entities.Variation{
		Key:            variationID,
		ID:             variationID,
		FeatureEnabled: true,
	}

	experiment := entities.Experiment{
		Key:                 experimentID,
		LayerID:             layerID,
		ID:                  experimentID,
		Variations:          map[string]entities.Variation{variationID: variation},
		VariationKeyToIDMap: map[string]string{variation.Key: variationID},
		TrafficAllocation: []entities.Range{
			entities.Range{EntityID: variationID, EndOfRange: 10000},
		},
	}

	f.FeatureExperiments = []entities.Experiment{experiment}
	c.FeatureMap[f.Key] = f
	c.ExperimentMap[experiment.Key] = experiment
	return c
}

// AddFeatureRollout adds the feature and supporting entities to complete the rollout modeling
func (c *TestProjectConfig) AddFeatureRollout(f entities.Feature) *TestProjectConfig {
	experimentID := c.getNextID()
	rolloutID := c.getNextID()
	variationID := c.getNextID()
	layerID := c.getNextID()

	variation := entities.Variation{
		Key:            variationID,
		ID:             variationID,
		FeatureEnabled: true,
	}

	experiment := entities.Experiment{
		Key:                 experimentID,
		LayerID:             layerID,
		ID:                  experimentID,
		Variations:          map[string]entities.Variation{variationID: variation},
		VariationKeyToIDMap: map[string]string{variation.Key: variationID},
		TrafficAllocation: []entities.Range{
			entities.Range{EntityID: variationID, EndOfRange: 10000},
		},
	}

	rollout := entities.Rollout{
		ID:          rolloutID,
		Experiments: []entities.Experiment{experiment},
	}

	c.RolloutMap[rolloutID] = rollout

	f.Rollout = rollout
	c.FeatureMap[f.Key] = f
	return c
}

// AddDisabledFeatureRollout adds modeling of a disabled feature rollout (variation's FeatureEnabled is false)
func (c *TestProjectConfig) AddDisabledFeatureRollout(f entities.Feature) *TestProjectConfig {
	c.AddFeatureRollout(f)
	variations := c.FeatureMap[f.Key].Rollout.Experiments[0].Variations
	for _, variation := range variations {
		variation.FeatureEnabled = false
		variations[variation.ID] = variation
	}
	return c
}

// AddFeatureTestWithCustomVariableValue adds modeling of a 1-variation feature test with a custom variable value
func (c *TestProjectConfig) AddFeatureTestWithCustomVariableValue(feature entities.Feature, variable entities.Variable, customValue string) *TestProjectConfig {
	c.AddFeatureRollout(feature)
	c.AddFeatureTest(feature)
	variations := c.FeatureMap[feature.Key].FeatureExperiments[0].Variations
	customVariableValues := map[string]entities.VariationVariable{
		variable.ID: {
			ID:    variable.ID,
			Value: customValue,
		},
	}
	for _, variation := range variations {
		variation.Variables = customVariableValues
		variations[variation.ID] = variation
	}
	return c
}

// AddExperiment adds the experiment and the supporting entities to complete the experiment modeling
func (c *TestProjectConfig) AddExperiment(experimentKey string, variations []entities.Variation) {
	experimentID := c.getNextID()
	layerID := c.getNextID()

	variationMap := map[string]entities.Variation{}
	variationKeyTOIDMap := make(map[string]string)
	trafficAllocation := []entities.Range{}
	for i, variation := range variations {
		variationMap[variation.ID] = variation
		variationKeyTOIDMap[variation.Key] = variation.ID
		endOfRange := 10000 / len(variations) * (i + 1)
		trafficAllocation = append(trafficAllocation, entities.Range{EntityID: variation.ID, EndOfRange: endOfRange})
	}

	experiment := entities.Experiment{
		Key:                 experimentKey,
		ID:                  experimentID,
		LayerID:             layerID,
		Variations:          variationMap,
		VariationKeyToIDMap: variationKeyTOIDMap,
		TrafficAllocation:   trafficAllocation,
	}

	c.ExperimentKeyToIDMap[experimentKey] = experimentID
	c.ExperimentMap[experimentID] = experiment
}

// CreateVariation creates a variation with the given key and a generated ID
func (c *TestProjectConfig) CreateVariation(varKey string) entities.Variation {
	variationID := c.getNextID()
	variation := entities.Variation{
		Key: varKey,
		ID:  variationID,
	}
	return variation
}

// ConvertVariation converts entities variation to optimizely config variation
func (c *TestProjectConfig) ConvertVariation(v entities.Variation) optlyconfig.OptimizelyVariation {

	variation := optlyconfig.OptimizelyVariation{
		Key:          v.Key,
		ID:           v.ID,
		VariablesMap: map[string]optlyconfig.OptimizelyVariable{},
	}
	return variation
}

// AddMultiVariationFeatureTest adds the feature and supporting entities to complete modeling of the following:
// - Feature test with two variations
// - Feature is disabled in first variation, enabled in second variation
// - Traffic allocation is 100% first variation, 0% second variation
func (c *TestProjectConfig) AddMultiVariationFeatureTest(f entities.Feature, disabledVariationKey, enabledVariationKey string) *TestProjectConfig {
	experimentID := c.getNextID()
	disabledVariationID := c.getNextID()
	enabledVariationID := c.getNextID()
	layerID := c.getNextID()

	disabledVariation := entities.Variation{
		Key:            disabledVariationKey,
		ID:             disabledVariationID,
		FeatureEnabled: false,
	}
	enabledVariation := entities.Variation{
		Key:            enabledVariationKey,
		ID:             enabledVariationID,
		FeatureEnabled: true,
	}

	experiment := entities.Experiment{
		// Note: experiment ID and Key are the same
		Key:     experimentID,
		LayerID: layerID,
		ID:      experimentID,
		Variations: map[string]entities.Variation{
			disabledVariationID: disabledVariation,
			enabledVariationID:  enabledVariation,
		},
		VariationKeyToIDMap: map[string]string{
			enabledVariationKey: enabledVariationID,
		},
		TrafficAllocation: []entities.Range{
			// Note: Intentionally using the same variation ID for both ranges.
			// This is a valid representation that can occur in real datafiles.
			// This happens when traffic started out as 50/50, and then was changed to 100/0.
			entities.Range{EntityID: disabledVariationID, EndOfRange: 5000},
			entities.Range{EntityID: disabledVariationID, EndOfRange: 10000},
		},
	}

	// Note: experiment ID and Key are the same
	c.ExperimentKeyToIDMap[experimentID] = experimentID
	c.ExperimentMap[experimentID] = experiment

	f.FeatureExperiments = []entities.Experiment{experiment}
	c.FeatureMap[f.Key] = f
	return c
}

// AddMultiVariationABTest adds an AB test with the following property:
// - Traffic allocation is 100% first variation, 0% second variation
func (c *TestProjectConfig) AddMultiVariationABTest(experimentKey, variationAKey, variationBKey string) *TestProjectConfig {
	variationA := c.CreateVariation(variationAKey)
	variationB := c.CreateVariation(variationBKey)

	variationMap := map[string]entities.Variation{
		variationA.ID: variationA,
		variationB.ID: variationB,
	}
	variationKeyToIDMap := map[string]string{
		variationA.Key: variationA.ID,
		variationB.Key: variationB.ID,
	}
	trafficAllocation := []entities.Range{
		// Note: Intentionally using the same variation ID for both ranges.
		// This is a valid representation that can occur in real datafiles.
		// This happens when traffic started out as 50/50, and then was changed to 100/0.
		entities.Range{EntityID: variationB.ID, EndOfRange: 5000},
		entities.Range{EntityID: variationB.ID, EndOfRange: 10000},
	}

	experimentID := c.getNextID()
	layerID := c.getNextID()
	experiment := entities.Experiment{
		Key:                 experimentKey,
		ID:                  experimentID,
		LayerID:             layerID,
		Variations:          variationMap,
		VariationKeyToIDMap: variationKeyToIDMap,
		TrafficAllocation:   trafficAllocation,
	}

	c.ExperimentKeyToIDMap[experimentKey] = experimentID
	c.ExperimentMap[experimentID] = experiment

	return c
}

func (c *TestProjectConfig) getNextID() (nextID string) {
	c.nextID++
	return strconv.Itoa(c.nextID)
}

// NewConfig initializes a new datafile from a json byte array using the default JSON datafile parser
func NewConfig() *TestProjectConfig {
	config := &TestProjectConfig{
		AccountID:            "accountId",
		AnonymizeIP:          true,
		AttributeKeyToIDMap:  make(map[string]string),
		AudienceMap:          make(map[string]entities.Audience),
		AttributeMap:         make(map[string]entities.Attribute),
		BotFiltering:         true,
		ExperimentKeyToIDMap: make(map[string]string),
		ExperimentMap:        make(map[string]entities.Experiment),
		EventMap:             make(map[string]entities.Event),
		FeatureMap:           make(map[string]entities.Feature),
		ProjectID:            "projectId",
		Revision:             "revision",
		RolloutMap:           make(map[string]entities.Rollout),
	}

	return config
}
