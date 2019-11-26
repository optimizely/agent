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
	"github.com/optimizely/sidedoor/pkg/event"
	"testing"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	events "github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/sidedoor/pkg/optimizelytest"
)

var counter int

type CacheTestSuite struct {
	suite.Suite
	cache *OptlyCache
}

func (suite *CacheTestSuite) SetupTest() {
	suite.cache = &OptlyCache{
		loader:   mockLoader,
		optlyMap: cmap.New(),
	}
}

func (suite *CacheTestSuite) TestGetCacheHit() {
	optltyClient1, err1 := suite.cache.GetClient("one")
	optltyClient2, err2 := suite.cache.GetClient("one")

	suite.NoError(err1)
	suite.NoError(err2)
	suite.Equal(optltyClient1, optltyClient2)
}

func (suite *CacheTestSuite) TestGetCacheMiss() {
	optltyClient1, err1 := suite.cache.GetClient("one")
	optltyClient2, err2 := suite.cache.GetClient("two")

	suite.NoError(err1)
	suite.NoError(err2)
	suite.NotEqual(optltyClient1, optltyClient2)
}

func (suite *CacheTestSuite) TestGetError() {
	_, err1 := suite.cache.GetClient("ERROR")
	suite.Error(err1)
}

func (suite *CacheTestSuite) TestInit() {
	viper.SetDefault("optimizely.sdkKeys", "one")
	suite.cache.init()
	suite.True(suite.cache.optlyMap.Has("one"))
	suite.False(suite.cache.optlyMap.Has("two"))
}

func (suite *CacheTestSuite) TestGetEventProcessorWithQueueSize() {
	viper.SetDefault(EPQSize, 1000)
	ep := GetOptlyEventProcessor()
	if bep, ok := ep.(*events.BatchEventProcessor); ok {
		suite.True(bep.MaxQueueSize == 1000)
	}
}

func (suite *CacheTestSuite) TestGetEventProcessorWithBatchSize() {
	viper.SetDefault(EPBSize, 30)
	ep := GetOptlyEventProcessor()
	if bep, ok := ep.(*events.BatchEventProcessor); ok {
		suite.True(bep.BatchSize == 30)
	}
}

func (suite *CacheTestSuite) TestGetEventProcessorWithNSQ() {
	viper.SetDefault(EPBSize, 30)
	viper.SetDefault(NSQEnabled, true)
	viper.SetDefault(NSQConsumer, true)
	viper.SetDefault(NSQProducer, true)
	viper.SetDefault(NSQStartEmbedded, false)

	ep := GetOptlyEventProcessor()
	if bep, ok := ep.(*events.BatchEventProcessor); ok {
		suite.True(bep.BatchSize == 30)
		if nsq, ok := bep.Q.(*event.NSQQueue); ok {
			suite.NotNil(nsq.Consumer)
			suite.NotNil(nsq.Producer)
		} else {
			suite.True(false)
		}
	}
}

func (suite *CacheTestSuite) TestGetEventProcessorWithoutNSQ() {
	viper.SetDefault(EPBSize, 30)

	ep := GetOptlyEventProcessor()
	if bep, ok := ep.(*events.BatchEventProcessor); ok {
		suite.True(bep.BatchSize == 30)
		if _, ok := bep.Q.(*event.NSQQueue); ok {
			suite.True(false)
		} else {
			suite.True(true)
		}
	}
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

	return &OptlyClient{tc.OptimizelyClient, nil, tc.ForcedVariations}, nil
}
