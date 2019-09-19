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
	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/config"
	"github.com/optimizely/go-sdk/optimizely/entities"
)

type TestClient struct {
	ProjectConfig    *TestProjectConfig
	OptimizelyClient *client.OptimizelyClient
}

func NewClient() (*TestClient) {
	projectConfig := NewConfig()

	options := client.Options{
		ProjectConfigManager: config.NewStaticProjectConfigManager(projectConfig),
	}

	factory := client.OptimizelyFactory{}
	optlyClient, _ := factory.ClientWithOptions(options)

	return &TestClient{
		ProjectConfig:		projectConfig, 
		OptimizelyClient:	optlyClient,
	}
}

func (t TestClient) AddFeature(feature entities.Feature) {
	t.ProjectConfig.AddFeature(feature)
}

func (t *TestClient) AddFeatureRollout(feature entities.Feature) {
	t.ProjectConfig.AddFeatureRollout(feature)
}
