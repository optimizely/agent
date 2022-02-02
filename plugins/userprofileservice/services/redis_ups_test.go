/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                        *
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

	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/stretchr/testify/suite"
)

type RedisUPSTestSuite struct {
	suite.Suite
	ups RedisUserProfileService
}

func (r *RedisUPSTestSuite) SetupTest() {
	// To check if lifo is used by default
	r.ups = RedisUserProfileService{
		Address:  "localhost:6379",
		Password: "", // no password set
		Database: 0,  // use default DB
	}
}

func (r *RedisUPSTestSuite) TestInitClient() {
	r.ups.initClient()
	pong, err := r.ups.Client.Ping(ctx).Result()
	r.Equal("PONG", pong)
	r.Nil(err)
}

func (r *RedisUPSTestSuite) TestFirstSaveOrLookupConfiguresClient() {
	r.Nil(r.ups.Client)

	profile := decision.UserProfile{
		ID: "1",
		ExperimentBucketMap: map[decision.UserDecisionKey]string{
			decision.NewUserDecisionKey("1"): "1",
		},
	}
	// Should initialize redis client on first save call
	r.ups.Save(profile)
	r.NotNil(r.ups.Client)

	r.ups.Client = nil
	// Should initialize redis client on first save call
	r.ups.Lookup("")
	r.NotNil(r.ups.Client)
}

func (r *RedisUPSTestSuite) TestLookupEmptyProfileID() {
	expectedProfile := decision.UserProfile{
		ID:                  "",
		ExperimentBucketMap: map[decision.UserDecisionKey]string{},
	}
	r.Equal(expectedProfile, r.ups.Lookup(""))
}

func (r *RedisUPSTestSuite) TestLookupNotSavedProfileID() {
	expectedProfile := decision.UserProfile{
		ID:                  "",
		ExperimentBucketMap: map[decision.UserDecisionKey]string{},
	}
	r.Equal(expectedProfile, r.ups.Lookup("123"))
}

func (r *RedisUPSTestSuite) TestSaveAndLookup() {
	expectedProfile := decision.UserProfile{
		ID: "1",
		ExperimentBucketMap: map[decision.UserDecisionKey]string{
			decision.NewUserDecisionKey("1"): "1",
		},
	}
	r.ups.Save(expectedProfile)
	r.Equal(expectedProfile, r.ups.Lookup("1"))
}

func TestRedisUPSTestSuite(t *testing.T) {
	suite.Run(t, new(RedisUPSTestSuite))
}
