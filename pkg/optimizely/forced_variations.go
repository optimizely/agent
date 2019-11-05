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

// Package optimizely wraps the Optimizely SDK
package optimizely

import (
	"fmt"

	"github.com/optimizely/go-sdk/pkg/decision"
	cmap "github.com/orcaman/concurrent-map"
)

// CMapExpOverridesStore is an implementation of the ExperimentOverrideStore interface from the Go SDK, based on a ConcurrentMap
// The GetVariation, SetVariation, and RemoveVariation methods are safe to call concurrently from multiple goroutines.
type CMapExpOverridesStore struct {
	overrides cmap.ConcurrentMap
}

// NewCMapExpOverridesStore creates a new CMapExpOverridesStore
func NewCMapExpOverridesStore() *CMapExpOverridesStore {
	return &CMapExpOverridesStore{
		overrides: cmap.New(),
	}
}

func cMapKeyOfExpOverrideKey(overrideKey decision.ExperimentOverrideKey) string {
	// Note: We are assuming that it's acceptable to use a space as the delimiter
	// This can break at any time if the app starts allowing spaces in experiment keys
	return fmt.Sprintf("%v %v", overrideKey.ExperimentKey, overrideKey.UserID)
}

// GetVariation returns the override for the given experiment and user from the overrides ConcurrentMap.
func (c *CMapExpOverridesStore) GetVariation(overrideKey decision.ExperimentOverrideKey) (string, bool) {
	cmapVal, ok := c.overrides.Get(cMapKeyOfExpOverrideKey(overrideKey))
	if !ok {
		return "", ok
	}
	variationKey, typeAssertionOk := cmapVal.(string)
	if !typeAssertionOk {
		// TODO: Invariant violated. This means a non-string value got into the overrides map. How to log/warn? Show a link to file an issue?
		return "", typeAssertionOk
	}
	return variationKey, typeAssertionOk
}

// SetVariation sets the argument forced variation for the argument experiment and user
func (c *CMapExpOverridesStore) SetVariation(overrideKey decision.ExperimentOverrideKey, variationKey string) {
	c.overrides.Set(cMapKeyOfExpOverrideKey(overrideKey), variationKey)
}

// RemoveVariation removes any forced variation for the argument experiment and user
func (c *CMapExpOverridesStore) RemoveVariation(overrideKey decision.ExperimentOverrideKey) {
	c.overrides.Remove(cMapKeyOfExpOverrideKey(overrideKey))
}
