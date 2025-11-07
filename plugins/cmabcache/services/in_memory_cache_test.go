/****************************************************************************
 * Copyright 2025, Optimizely, Inc. and contributors                        *
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

// Package services //
package services

import (
	"testing"
	"time"

	"github.com/optimizely/agent/plugins/utils"
	"github.com/stretchr/testify/suite"
)

type InMemoryCacheTestSuite struct {
	suite.Suite
	cache InMemoryCache
}

func (im *InMemoryCacheTestSuite) SetupTest() {
	im.cache = InMemoryCache{
		Size:    10000,
		Timeout: utils.Duration{Duration: 600 * time.Second},
	}
}

func (im *InMemoryCacheTestSuite) TestLookupInitializesCache() {
	im.Nil(im.cache.LRUCache)
	im.cache.Lookup("abc")
	im.NotNil(im.cache.LRUCache)
}

func (im *InMemoryCacheTestSuite) TestSaveInitializesCacheAndSaves() {
	im.Nil(im.cache.LRUCache)

	// Test with CMAB decision object
	decision := map[string]interface{}{
		"variationID":    "variation_1",
		"attributesHash": "hash123",
		"cmabUUID":       "uuid-456",
	}

	im.cache.Save("user123:exp456", decision)
	im.NotNil(im.cache.LRUCache)

	result := im.cache.Lookup("user123:exp456")
	im.NotNil(result)
	im.Equal(decision, result)
}

func (im *InMemoryCacheTestSuite) TestLookupReturnsNilForMissingKey() {
	im.cache.Save("key1", "value1")
	im.Nil(im.cache.Lookup("nonexistent"))
}

func (im *InMemoryCacheTestSuite) TestRemoveWithoutInitialization() {
	im.Nil(im.cache.LRUCache)
	im.cache.Remove("key1")
	im.Nil(im.cache.LRUCache)
}

func (im *InMemoryCacheTestSuite) TestRemoveDeletesKey() {
	decision := map[string]interface{}{
		"variationID": "variation_1",
	}

	im.cache.Save("user123:exp456", decision)
	im.NotNil(im.cache.Lookup("user123:exp456"))

	im.cache.Remove("user123:exp456")
	im.Nil(im.cache.Lookup("user123:exp456"))
}

func (im *InMemoryCacheTestSuite) TestResetWithoutInitialization() {
	im.Nil(im.cache.LRUCache)
	im.cache.Reset()
	im.Nil(im.cache.LRUCache)
}

func (im *InMemoryCacheTestSuite) TestResetAfterSave() {
	im.Nil(im.cache.LRUCache)

	im.cache.Save("key1", "value1")
	im.cache.Save("key2", "value2")
	im.NotNil(im.cache.LRUCache)

	im.Equal("value1", im.cache.Lookup("key1"))
	im.Equal("value2", im.cache.Lookup("key2"))

	im.cache.Reset()

	im.Nil(im.cache.Lookup("key1"))
	im.Nil(im.cache.Lookup("key2"))
}

func (im *InMemoryCacheTestSuite) TestMultipleDecisions() {
	decision1 := map[string]interface{}{
		"variationID":    "var1",
		"attributesHash": "hash1",
	}
	decision2 := map[string]interface{}{
		"variationID":    "var2",
		"attributesHash": "hash2",
	}

	im.cache.Save("user1:exp1", decision1)
	im.cache.Save("user2:exp1", decision2)

	im.Equal(decision1, im.cache.Lookup("user1:exp1"))
	im.Equal(decision2, im.cache.Lookup("user2:exp1"))
}

func TestInMemoryCacheTestSuite(t *testing.T) {
	suite.Run(t, new(InMemoryCacheTestSuite))
}
