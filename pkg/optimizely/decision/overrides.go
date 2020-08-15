/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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

// Package decision provides decision logic on top of the SDK
package decision

import (
	"sync"

	"github.com/optimizely/go-sdk/pkg/decision"
)

// MapExperimentOverridesStore is a map-based implementation of ExperimentOverridesStore that is safe to use concurrently
type MapExperimentOverridesStore struct {
	OverridesMap map[decision.ExperimentOverrideKey]string
	mutex        sync.RWMutex
}

// NewMapExperimentOverridesStore returns a new MapExperimentOverridesStore
func NewMapExperimentOverridesStore() *MapExperimentOverridesStore {
	return &MapExperimentOverridesStore{
		OverridesMap: make(map[decision.ExperimentOverrideKey]string),
	}
}

// GetVariation returns the override variation key associated with the given user+experiment key
func (m *MapExperimentOverridesStore) GetVariation(overrideKey decision.ExperimentOverrideKey) (string, bool) {
	m.mutex.RLock()
	variationKey, ok := m.OverridesMap[overrideKey]
	m.mutex.RUnlock()
	return variationKey, ok
}

// SetVariation sets the given variation key as an override for the given user+experiment key
func (m *MapExperimentOverridesStore) SetVariation(overrideKey decision.ExperimentOverrideKey, variationKey string) {
	m.mutex.Lock()
	m.OverridesMap[overrideKey] = variationKey
	m.mutex.Unlock()
}

// RemoveVariation removes the override variation key associated with the argument user+experiment key.
// If there is no override variation key set, this method has no effect.
func (m *MapExperimentOverridesStore) RemoveVariation(overrideKey decision.ExperimentOverrideKey) {
	m.mutex.Lock()
	delete(m.OverridesMap, overrideKey)
	m.mutex.Unlock()
}
