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

func (im *InMemoryCacheTestSuite) TestLookInitializesCache() {
	im.Nil(im.cache.LRUCache)
	im.cache.Lookup("abc")
	im.NotNil(im.cache.LRUCache)
}

func (im *InMemoryCacheTestSuite) TestSaveInitializesCacheAndSaves() {
	im.Nil(im.cache.LRUCache)
	im.cache.Save("1", "100")
	im.NotNil(im.cache.LRUCache)
	im.Equal("100", im.cache.Lookup("1"))
}

func (im *InMemoryCacheTestSuite) TestResetWithoutInitialization() {
	im.Nil(im.cache.LRUCache)
	im.cache.Reset()
	im.Nil(im.cache.LRUCache)
}

func (im *InMemoryCacheTestSuite) TestResetAfterSave() {
	im.Nil(im.cache.LRUCache)
	im.cache.Save("1", "100")
	im.NotNil(im.cache.LRUCache)
	im.Equal("100", im.cache.Lookup("1"))
	im.cache.Reset()
	im.Nil(im.cache.Lookup("1"))
}

func TestInMemoryCacheTestSuite(t *testing.T) {
	suite.Run(t, new(InMemoryCacheTestSuite))
}
