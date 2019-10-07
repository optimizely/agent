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

// Package optlytest //
package optlytest

import (
	"github.com/optimizely/sidedoor/pkg/optimizely"
	"github.com/optimizely/sidedoor/pkg/optimizelytest"
)

// TestCache implements the Cache interface and is used in testing.
type TestCache struct {
	testClient		*optimizelytest.TestClient
}

// NewCache returns a new implementation of TestCache
func NewCache() *TestCache {
	testClient := optimizelytest.NewClient()
	return &TestCache{
		testClient: testClient,
	}
}

// GetDefaultClient returns a default OptlyClient for testing
func (tc *TestCache) GetDefaultClient() (*optimizely.OptlyClient, error) {
	return tc.GetClient("default")
}

// GetClient returns a default OptlyClient for testing
func (tc *TestCache) GetClient(sdkKey string) (*optimizely.OptlyClient, error) {
	return &optimizely.OptlyClient{
		OptimizelyClient: tc.testClient.OptimizelyClient,
		ConfigManager:    nil,
	}, nil
}
