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
	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
)

// TestClient encapsulates both the ProjectConfig interface and the OptimizelyClient
type TestClient struct {
	EventProcessor   *TestEventProcessor
	ProjectConfig    *TestProjectConfig
	OptimizelyClient *client.OptimizelyClient
}

// NewClient provides an instance of OptimizelyClient backed by a TestProjectConfig
func NewClient() *TestClient {
	projectConfig := NewConfig()
	eventProcessor := new(TestEventProcessor)

	factory := client.OptimizelyFactory{}
	optlyClient, _ := factory.Client(
		client.WithConfigManager(config.NewStaticProjectConfigManager(projectConfig)),
		client.WithEventProcessor(eventProcessor),
	)

	return &TestClient{
		EventProcessor:   eventProcessor,
		ProjectConfig:    projectConfig,
		OptimizelyClient: optlyClient,
	}
}

// AddEvent is a helper method for adding events to the ProjectConfig to fascilitate testing.
func (t *TestClient) AddEvent(e entities.Event) {
	t.ProjectConfig.AddEvent(e)
}

// AddFeature is a helper method for adding features to the ProjectConfig to fascilitate testing.
func (t TestClient) AddFeature(f entities.Feature) {
	t.ProjectConfig.AddFeature(f)
}

// AddFeatureRollout is a helper method for adding feature rollouts to the ProjectConfig to fascilitate testing.
func (t *TestClient) AddFeatureRollout(f entities.Feature) {
	t.ProjectConfig.AddFeatureRollout(f)
}

// GetProcessedEvents returns the UserEvent objects sent to the event processor.
func (t *TestClient) GetProcessedEvents() []event.UserEvent {
	return t.EventProcessor.GetEvents()
}
