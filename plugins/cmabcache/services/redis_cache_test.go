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

type RedisCacheTestSuite struct {
	suite.Suite
	cache RedisCache
}

func (r *RedisCacheTestSuite) SetupTest() {
	r.cache = RedisCache{
		Address:  "invalid-redis-host:6379",
		Password: "test-password",
		Database: 1,
		Timeout:  utils.Duration{Duration: 100 * time.Second},
	}
}

func (r *RedisCacheTestSuite) TestFirstSaveOrLookupConfiguresClient() {
	r.Nil(r.cache.Client)

	// Should initialize redis client on first save call
	decision := map[string]interface{}{
		"variationID": "var1",
	}
	r.cache.Save("key1", decision)
	r.NotNil(r.cache.Client)

	r.cache.Client = nil
	// Should initialize redis client on first lookup call
	r.cache.Lookup("key1")
	r.NotNil(r.cache.Client)
}

func (r *RedisCacheTestSuite) TestLookupEmptyKey() {
	r.Nil(r.cache.Lookup(""))
}

func (r *RedisCacheTestSuite) TestSaveEmptyKey() {
	r.cache.Save("", "value")
	// Should not panic, client should be initialized
	r.NotNil(r.cache.Client)
}

func (r *RedisCacheTestSuite) TestLookupNotSavedKey() {
	// This will fail to connect to Redis but shouldn't panic
	r.Nil(r.cache.Lookup("nonexistent-key"))
}

func (r *RedisCacheTestSuite) TestRemoveEmptyKey() {
	r.cache.Remove("")
	// Should not panic
	r.NotNil(r.cache.Client)
}

func (r *RedisCacheTestSuite) TestRemoveInitializesClient() {
	r.Nil(r.cache.Client)
	r.cache.Remove("key1")
	r.NotNil(r.cache.Client)
}

func (r *RedisCacheTestSuite) TestResetInitializesClient() {
	r.Nil(r.cache.Client)
	r.cache.Reset()
	r.NotNil(r.cache.Client)
}

func (r *RedisCacheTestSuite) TestClientConfiguration() {
	r.cache.initClient()

	r.NotNil(r.cache.Client)
	r.Equal("invalid-redis-host:6379", r.cache.Client.Options().Addr)
	r.Equal("test-password", r.cache.Client.Options().Password)
	r.Equal(1, r.cache.Client.Options().DB)
}

func TestRedisCacheTestSuite(t *testing.T) {
	suite.Run(t, new(RedisCacheTestSuite))
}
