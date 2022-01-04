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
		Address:  "100",
		Password: "10",
		Database: 1,
	}
}

func (r *RedisUPSTestSuite) TestFirstSaveConfiguresClient() {
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
}

func (r *RedisUPSTestSuite) TestLookupNilClient() {
	r.Nil(r.ups.Client)

	expected := decision.UserProfile{
		ID:                  "",
		ExperimentBucketMap: map[decision.UserDecisionKey]string{},
	}
	actual := r.ups.Lookup("1")
	r.Equal(expected, actual)
}

func TestRedisUPSTestSuite(t *testing.T) {
	suite.Run(t, new(RedisUPSTestSuite))
}
