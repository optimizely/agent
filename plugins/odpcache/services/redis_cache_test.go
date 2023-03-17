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

type RedisCacheTestSuite struct {
	suite.Suite
	cache RedisCache
}

func (r *RedisCacheTestSuite) SetupTest() {
	r.cache = RedisCache{
		Address:  "100",
		Password: "10",
		Database: 1,
		Timeout:  utils.Duration{Duration: 100 * time.Second},
	}
}

func (r *RedisCacheTestSuite) TestFirstSaveOrLookupConfiguresClient() {
	r.Nil(r.cache.Client)

	// Should initialize redis client on first save call
	r.cache.Save("1", []string{"1"})
	r.NotNil(r.cache.Client)

	r.cache.Client = nil
	// Should initialize redis client on first save call
	r.cache.Lookup("")
	r.NotNil(r.cache.Client)
}

func (r *RedisCacheTestSuite) TestLookupEmptyKey() {
	r.Nil(r.cache.Lookup(""))
}

func (r *RedisCacheTestSuite) TestLookupNotSavedKey() {
	r.Nil(r.cache.Lookup("123"))
}

func TestRedisCacheTestSuite(t *testing.T) {
	suite.Run(t, new(RedisCacheTestSuite))
}
