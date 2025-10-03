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

func TestRedisCache_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		json         string
		wantPassword string
		wantErr      bool
	}{
		{
			name:         "auth_token has priority",
			json:         `{"host":"localhost:6379","auth_token":"token123","password":"pass456","database":0}`,
			wantPassword: "token123",
			wantErr:      false,
		},
		{
			name:         "redis_secret when auth_token missing",
			json:         `{"host":"localhost:6379","redis_secret":"secret789","password":"pass456","database":0}`,
			wantPassword: "secret789",
			wantErr:      false,
		},
		{
			name:         "password when others missing",
			json:         `{"host":"localhost:6379","password":"pass456","database":0}`,
			wantPassword: "pass456",
			wantErr:      false,
		},
		{
			name:         "empty when no password fields",
			json:         `{"host":"localhost:6379","database":0}`,
			wantPassword: "",
			wantErr:      false,
		},
		{
			name:         "invalid json",
			json:         `{invalid}`,
			wantPassword: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cache RedisCache
			err := cache.UnmarshalJSON([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && cache.Password != tt.wantPassword {
				t.Errorf("UnmarshalJSON() Password = %v, want %v", cache.Password, tt.wantPassword)
			}
		})
	}
}
