/****************************************************************************
 * Copyright 2019,2021 Optimizely, Inc. and contributors                    *
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
	"time"

	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/odp"
	odp_event "github.com/optimizely/go-sdk/pkg/odp/event"
)

// TestClient encapsulates both the ProjectConfig interface and the OptimizelyClient
type TestClient struct {
	EventProcessor   *TestEventProcessor
	ProjectConfig    *TestProjectConfig
	OptimizelyClient *client.OptimizelyClient
	ForcedVariations *decision.MapExperimentOverridesStore
	EventApiManager  *TestEventApiManager
}

// NewClient provides an instance of OptimizelyClient backed by a TestProjectConfig
func NewClient() *TestClient {
	projectConfig := NewConfig()
	eventProcessor := new(TestEventProcessor)
	forcedVariations := decision.NewMapExperimentOverridesStore()

	eventApiManager := new(TestEventApiManager)

	eventOptions := []odp_event.EMOptionFunc{odp_event.WithAPIManager(eventApiManager), odp_event.WithFlushInterval(time.Duration(0))}
	eventManager := odp_event.NewBatchEventManager(eventOptions...)
	odpManagerOptions := []odp.OMOptionFunc{odp.WithEventManager(eventManager)}
	odpManager := odp.NewOdpManager(projectConfig.sdkKey, false, odpManagerOptions...)

	factory := client.OptimizelyFactory{}
	optlyClient, _ := factory.Client(
		client.WithConfigManager(config.NewStaticProjectConfigManager(projectConfig, logging.GetLogger("test", "test"))),
		client.WithEventProcessor(eventProcessor),
		client.WithExperimentOverrides(forcedVariations),
		client.WithOdpManager(odpManager),
	)

	return &TestClient{
		EventProcessor:   eventProcessor,
		ProjectConfig:    projectConfig,
		OptimizelyClient: optlyClient,
		ForcedVariations: forcedVariations,
		EventApiManager:  eventApiManager,
	}
}

// AddEvent is a helper method for adding events to the ProjectConfig to facilitate testing.
func (t *TestClient) AddEvent(e entities.Event) {
	t.ProjectConfig.AddEvent(e)
}

// AddFeature is a helper method for adding features to the ProjectConfig to facilitate testing.
func (t TestClient) AddFeature(f entities.Feature) {
	t.ProjectConfig.AddFeature(f)
}

// AddFeatureRollout is a helper method for adding feature rollouts to the ProjectConfig to facilitate testing.
func (t *TestClient) AddFeatureRollout(f entities.Feature) {
	t.ProjectConfig.AddFeatureRollout(f)
}

// AddFeatureTest is a helper method for adding feature rollouts to the ProjectConfig to facilitate testing.
func (t *TestClient) AddFeatureTest(f entities.Feature) {
	t.ProjectConfig.AddFeatureTest(f)
}

// AddFlagVariation is a helper method for adding flag variation to the ProjectConfig to facilitate testing.
func (t *TestClient) AddFlagVariation(f entities.Feature, v entities.Variation) {
	t.ProjectConfig.AddFlagVariation(f, v)
}

// GetProcessedEvents returns the UserEvent objects sent to the event processor.
func (t *TestClient) GetProcessedEvents() []event.UserEvent {
	return t.EventProcessor.GetEvents()
}

// AddExperimentWithVariations is a helper method for adding and experiment with N variations
func (t TestClient) AddExperimentWithVariations(experimentKey string, variationKeys ...string) {
	variations := make([]entities.Variation, len(variationKeys))
	for i, key := range variationKeys {
		variations[i] = entities.Variation{Key: key}
	}

	t.AddExperiment(experimentKey, variations)
}

// AddExperiment is a helper method for creating experiments in the ProjectConfig to facilitate testing.
func (t *TestClient) AddExperiment(experimentKey string, variations []entities.Variation) {
	t.ProjectConfig.AddExperiment(experimentKey, variations)
}

// AddDisabledFeatureRollout is a helper method for creating a disabled rollout in the ProjectConfig to facilitate testing.
func (t *TestClient) AddDisabledFeatureRollout(f entities.Feature) {
	t.ProjectConfig.AddDisabledFeatureRollout(f)
}

// AddFeatureTestWithCustomVariableValue is a helper method for creating a feature test with a custom variable value in the ProjectConfig to facilitate testing.
func (t *TestClient) AddFeatureTestWithCustomVariableValue(feature entities.Feature, variable entities.Variable, customValue string) {
	t.ProjectConfig.AddFeatureTestWithCustomVariableValue(feature, variable, customValue)
}
