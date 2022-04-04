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

// Package datafilecacheservice //
package datafilecacheservice

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RedisCacheServiceTestSuite struct {
	suite.Suite
	service RedisCacheService
	ctx     context.Context
	cancel  func()
}

func (r *RedisCacheServiceTestSuite) SetupTest() {
	// To check if lifo is used by default
	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.service = RedisCacheService{
		Address:  "100",
		Password: "10",
		Database: 1,
	}
}

func (r *RedisCacheServiceTestSuite) TearDownTest() {
	r.cancel()
}

func (r *RedisCacheServiceTestSuite) TestFirstSaveOrLookupConfiguresClient() {
	r.Nil(r.service.Client)

	// Should initialize redis client on first SetDatafileInCacheService call
	r.service.SetDatafileInCacheService(r.ctx, "123", `{"abs":123,}`)
	r.NotNil(r.service.Client)

	r.service.Client = nil
	// Should initialize redis client on first GetDatafileFromCacheService call
	r.service.GetDatafileFromCacheService(r.ctx, "123")
	r.NotNil(r.service.Client)
}

func TestRedisUPSTestSuite(t *testing.T) {
	suite.Run(t, new(RedisCacheServiceTestSuite))
}
