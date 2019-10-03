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
	"testing"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/sidedoor/pkg/optimizelytest"
)

var counter int

type CacheTestSuite struct {
	suite.Suite
	cache *OptlyCache
}

func (suite *CacheTestSuite) SetupTest() {
	suite.cache = &OptlyCache{
		defaultKey: "default",
		loader:     mockLoader,
		optlyMap:   cmap.New(),
	}
}

func (suite *CacheTestSuite) TestGetDefault() {
	optltyClient1, err1 := suite.cache.GetDefault()
	optltyClient2, err2 := suite.cache.GetDefault()

	suite.NoError(err1)
	suite.NoError(err2)
	suite.Equal(optltyClient1, optltyClient2)
}

func (suite *CacheTestSuite) TestGetCacheHit() {
	optltyClient1, err1 := suite.cache.Get("one")
	optltyClient2, err2 := suite.cache.Get("one")

	suite.NoError(err1)
	suite.NoError(err2)
	suite.Equal(optltyClient1, optltyClient2)
}

func (suite *CacheTestSuite) TestGetCacheMiss() {
	optltyClient1, err1 := suite.cache.Get("one")
	optltyClient2, err2 := suite.cache.Get("two")

	suite.NoError(err1)
	suite.NoError(err2)
	suite.NotEqual(optltyClient1, optltyClient2)
}

func (suite *CacheTestSuite) TestGetError() {
	_, err1 := suite.cache.Get("ERROR")
	suite.Error(err1)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}

func mockLoader(sdkKey string) (*OptlyClient, error) {
	if sdkKey == "ERROR" {
		return &OptlyClient{}, fmt.Errorf("Error")
	}

	counter++
	tc := optimizelytest.NewClient()
	tc.ProjectConfig.ProjectID = sdkKey

	return &OptlyClient{tc.OptimizelyClient, nil}, nil
}
