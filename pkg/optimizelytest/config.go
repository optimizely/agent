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

	"github.com/optimizely/go-sdk/optimizely/entities"
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
func (c TestProjectConfig) GetProjectID() string {
	return c.ProjectID
}

// GetRevision returns revision
func (c TestProjectConfig) GetRevision() string {
	return c.Revision
}

// GetAccountID returns accountID
func (c TestProjectConfig) GetAccountID() string {
	return c.AccountID
}

// GetAnonymizeIP returns anonymizeIP
func (c TestProjectConfig) GetAnonymizeIP() bool {
	return c.AnonymizeIP
}

// GetAttributeID returns attributeID
func (c TestProjectConfig) GetAttributeID(key string) string {
	return c.AttributeKeyToIDMap[key]
}

// GetBotFiltering returns GetBotFiltering
func (c TestProjectConfig) GetBotFiltering() bool {
	return c.BotFiltering
}

// GetEventByKey returns the event with the given key
func (c TestProjectConfig) GetEventByKey(eventKey string) (entities.Event, error) {
	if event, ok := c.EventMap[eventKey]; ok {
		return event, nil
	}

	errMessage := fmt.Sprintf("Event with key %s not found", eventKey)
	return entities.Event{}, errors.New(errMessage)
}

// GetFeatureByKey returns the feature with the given key
func (c TestProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	if feature, ok := c.FeatureMap[featureKey]; ok {
		return feature, nil
	}

	errMessage := fmt.Sprintf("Feature with key %s not found", featureKey)
	return entities.Feature{}, errors.New(errMessage)
}

// GetVariableByKey returns the featureVariable with the given key
func (c TestProjectConfig) GetVariableByKey(featureKey, variableKey string) (entities.Variable, error) {

	var variable entities.Variable
	var err = fmt.Errorf("variable with key %s not found", featureKey)
	if feature, ok := c.FeatureMap[featureKey]; ok {
		for _, v := range feature.Variables {
			if v.Key == variableKey {
				variable = v
				err = nil
				break
			}
		}
	}
	return variable, err
}

// GetAttributeByKey returns the attribute with the given key
func (c TestProjectConfig) GetAttributeByKey(key string) (entities.Attribute, error) {
	if attributeID, ok := c.AttributeKeyToIDMap[key]; ok {
		return c.AttributeMap[attributeID], nil
	}

	errMessage := fmt.Sprintf(`Attribute with key "%s" not found`, key)
	return entities.Attribute{}, errors.New(errMessage)
}

// GetFeatureList returns an array of all the features
func (c TestProjectConfig) GetFeatureList() (featureList []entities.Feature) {
	for _, feature := range c.FeatureMap {
		featureList = append(featureList, feature)
	}
	return featureList
}

// GetAudienceByID returns the audience with the given ID
func (c TestProjectConfig) GetAudienceByID(audienceID string) (entities.Audience, error) {
	if audience, ok := c.AudienceMap[audienceID]; ok {
		return audience, nil
	}

	errMessage := fmt.Sprintf(`Audience with ID "%s" not found`, audienceID)
	return entities.Audience{}, errors.New(errMessage)
}

// GetAudienceMap returns the audience map
func (c TestProjectConfig) GetAudienceMap() map[string]entities.Audience {
	return c.AudienceMap
}

// GetExperimentByKey returns the experiment with the given key
func (c TestProjectConfig) GetExperimentByKey(experimentKey string) (entities.Experiment, error) {
	if experimentID, ok := c.ExperimentKeyToIDMap[experimentKey]; ok {
		experiment := c.ExperimentMap[experimentID]
		return experiment, nil
	}

	errMessage := fmt.Sprintf(`Experiment with key "%s" not found`, experimentKey)
	return entities.Experiment{}, errors.New(errMessage)
}

// GetGroupByID returns the group with the given ID
func (c TestProjectConfig) GetGroupByID(groupID string) (entities.Group, error) {
	if group, ok := c.GroupMap[groupID]; ok {
		return group, nil
	}

	errMessage := fmt.Sprintf(`Group with ID "%s" not found`, groupID)
	return entities.Group{}, errors.New(errMessage)
}

func (c TestProjectConfig) AddFeature(feature entities.Feature) *TestProjectConfig {
	c.FeatureMap[feature.Key] = feature
	return &c
}

func (c TestProjectConfig) AddFeatureRollout(feature entities.Feature) *TestProjectConfig {
	experimentID := c.getNextID()
	rolloutID := c.getNextID()
	variationID := c.getNextID()
	layerID := c.getNextID()

	variation := entities.Variation{
		Key:            "rollout_var",
		ID:             variationID,
		FeatureEnabled: true,
	}

	experiment := entities.Experiment{
		Key:        "background_experiment",
		LayerID:    layerID,
		ID:         experimentID,
		Variations: map[string]entities.Variation{variationID: variation},
		TrafficAllocation: []entities.Range{
			entities.Range{EntityID: variationID, EndOfRange: 10000},
		},
	}

	rollout := entities.Rollout{
		ID:          rolloutID,
		Experiments: []entities.Experiment{experiment},
	}

	c.RolloutMap[rolloutID] = rollout

	feature.Rollout = rollout
	c.FeatureMap[feature.Key] = feature
	return &c
}

func (c TestProjectConfig) getNextID() (nextID string) {
	c.nextID++
	return strconv.Itoa(c.nextID)
}

// NewTestProjectConfig initializes a new datafile from a json byte array using the default JSON datafile parser
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
