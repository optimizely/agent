/****************************************************************************
 * Copyright 2023, Optimizely, Inc. and contributors                        *
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

// Package odpcache //
package odpcache

import (
	"testing"

	"github.com/optimizely/go-sdk/v2/pkg/cache"
	"github.com/stretchr/testify/assert"
)

type MockODPCache struct {
}

// Lookup is used to retrieve segments
func (i *MockODPCache) Lookup(key string) (segments interface{}) {
	return nil
}

// Save is used to save segments
func (i *MockODPCache) Save(key string, value interface{}) {
}

// Reset is used to reset segments
func (i *MockODPCache) Reset() {
}

func TestAdd(t *testing.T) {
	mockODPCacheCreator := func() cache.Cache {
		return &MockODPCache{}
	}

	Add("mock", mockODPCacheCreator)
	creator := Creators["mock"]()
	if _, ok := creator.(*MockODPCache); !ok {
		assert.Fail(t, "Cannot convert to type InMemoryODPCache")
	}
}

func TestDuplicateKeys(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			assert.Fail(t, "Should have recovered")
		}
	}()

	mockODPCacheCreator := func() cache.Cache {
		return &MockODPCache{}
	}

	Add("mock", mockODPCacheCreator)
	Add("mock", mockODPCacheCreator)
	assert.Fail(t, "Should have panicked")
}

func TestDoesNotExist(t *testing.T) {
	dne := Creators["DNE"]
	assert.Nil(t, dne)
}
