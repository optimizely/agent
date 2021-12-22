/****************************************************************************
 * Copyright 2019,2021 Optimizely, Inc. and contributors                    *
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
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	sdkconfig "github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/event"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/metrics"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
	"github.com/optimizely/agent/plugins/userprofileservice"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/agent/plugins/userprofileservice/services"
)

var counter int

type CacheTestSuite struct {
	suite.Suite
	cache  *OptlyCache
	cancel func()
}

func (suite *CacheTestSuite) SetupTest() {
	ctx, cancel := context.WithCancel(context.Background())

	suite.cache = &OptlyCache{
		loader:   mockLoader,
		optlyMap: cmap.New(),
		ctx:      ctx,
	}

	suite.cancel = cancel
}

func (suite *CacheTestSuite) TearDownTest() {
	suite.cancel()
	suite.cache.Wait()
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
	suite.cache.Init([]string{"one", "three:four"})
	suite.True(suite.cache.optlyMap.Has("one"))
	suite.True(suite.cache.optlyMap.Has("three:four"))
	suite.False(suite.cache.optlyMap.Has("two"))
}

func (suite *CacheTestSuite) TestUpdateConfigs() {
	_, _ = suite.cache.GetClient("one")
	_, _ = suite.cache.GetClient("one:two")
	_, _ = suite.cache.GetClient("one:three")

	suite.cache.UpdateConfigs("one")
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

	return &OptlyClient{tc.OptimizelyClient, nil, tc.ForcedVariations, nil}, nil
}

type MockUserProfileService struct {
	Path string `json:"path"`
	Addr string `json:"addr"`
	Port int    `json:"port"`
}

// Lookup is used to retrieve past bucketing decisions for users
func (u *MockUserProfileService) Lookup(userID string) decision.UserProfile {
	return decision.UserProfile{}
}

// Save is used to save bucketing decisions for users
func (u *MockUserProfileService) Save(profile decision.UserProfile) {
}

var doOnce sync.Once // required since we only need to read datafile once

type DefaultLoaderTestSuite struct {
	suite.Suite
	registry  *MetricsRegistry
	bp        *event.BatchEventProcessor
	bpFactory func(options ...event.BPOptionConfig) *event.BatchEventProcessor
	pcFactory func(sdkKey string, options ...sdkconfig.OptionFunc) SyncedConfigManager
}

func (s *DefaultLoaderTestSuite) SetupTest() {
	// Need the registry to be created only once since it panics if we create gauges with the same name again and again
	doOnce.Do(func() {
		s.registry = &MetricsRegistry{metrics.NewRegistry()}
	})
	s.bpFactory = func(options ...event.BPOptionConfig) *event.BatchEventProcessor {
		s.bp = event.NewBatchEventProcessor(options...)
		return s.bp
	}
	// Note we're NOT testing that the ConfigManager was configured properly
	// This would require a bit larger refactor since the optimizelyFactory.Client takes a few liberties
	s.pcFactory = func(sdkKey string, options ...sdkconfig.OptionFunc) SyncedConfigManager {
		return MockConfigManager{}
	}
}

func (s *DefaultLoaderTestSuite) TestDefaultLoader() {
	conf := config.ClientConfig{
		FlushInterval: 321 * time.Second,
		BatchSize:     1234,
		QueueSize:     5678,
		EventURL:      "https://localhost/events",
		SdkKeyRegex:   "sdkkey",
		UserProfileServices: map[string]interface{}{"default": "in-memory", "services": map[string]interface{}{
			"in-memory": map[string]interface{}{
				"capacity": 0,
				"order":    "fifo",
			}},
		},
	}

	loader := defaultLoader(conf, s.registry, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)

	s.Equal(conf.FlushInterval, s.bp.FlushInterval)
	s.Equal(conf.BatchSize, s.bp.BatchSize)
	s.Equal(conf.QueueSize, s.bp.MaxQueueSize)
	s.Equal(conf.EventURL, s.bp.EventEndPoint)
	s.NotNil(client.UserProfileService)

	_, err = loader("invalid!")
	s.Error(err)
}

func (s *DefaultLoaderTestSuite) TestInMemoryAndRedisUpsAddedByDefault() {
	s.NotNil(userprofileservice.Creators["in-memory"])
	_, ok := (userprofileservice.Creators["in-memory"]()).(*services.InMemoryUserProfileService)
	s.True(ok)

	s.NotNil(userprofileservice.Creators["redis"])
	_, ok = (userprofileservice.Creators["redis"]()).(*services.RedisUserProfileService)
	s.True(ok)
}

func (s *DefaultLoaderTestSuite) TestConfigureClientForRedisUPS() {
	conf := config.ClientConfig{
		UserProfileServices: map[string]interface{}{"default": "redis", "services": map[string]interface{}{
			"redis": map[string]interface{}{
				"host":     "100",
				"password": "10",
				"database": 1,
			},
		}},
	}
	loader := defaultLoader(conf, s.registry, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)

	s.NotNil(client.UserProfileService)
	if testRedisUPS, ok := client.UserProfileService.(*services.RedisUserProfileService); ok {
		s.Equal("100", testRedisUPS.Address)
		s.Equal("10", testRedisUPS.Password)
		s.Equal(1, testRedisUPS.Database)

		// Check if redis client was instantiated with updated config
		s.NotNil(testRedisUPS.Client)
		s.Equal(testRedisUPS.Address, testRedisUPS.Client.Options().Addr)
		s.Equal(testRedisUPS.Password, testRedisUPS.Client.Options().Password)
		s.Equal(testRedisUPS.Database, testRedisUPS.Client.Options().DB)

		return
	}
	s.Failf("UserProfileService not registered", "%s DNE in registry", "redis")
}

func (s *DefaultLoaderTestSuite) TestLoaderWithValidUserProfileServices() {
	upCreator := func() decision.UserProfileService {
		return &MockUserProfileService{}
	}
	userprofileservice.Add("mock2", upCreator)

	conf := config.ClientConfig{
		UserProfileServices: map[string]interface{}{"default": "mock2", "services": map[string]interface{}{
			"mock2": map[string]interface{}{
				"path": "http://test.com",
				"addr": "1.2.1.2-abc",
				"port": 8080,
			},
		}},
	}
	loader := defaultLoader(conf, s.registry, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)

	s.NotNil(client.UserProfileService)
	if testRedisUPS, ok := client.UserProfileService.(*MockUserProfileService); ok {
		s.Equal("http://test.com", testRedisUPS.Path)
		s.Equal("1.2.1.2-abc", testRedisUPS.Addr)
		s.Equal(8080, testRedisUPS.Port)
		return
	}
	s.Failf("UserProfileService not registered", "%s DNE in registry", "redis")
}

func (s *DefaultLoaderTestSuite) TestLoaderWithEmptyUserProfileServices() {
	upCreator := func() decision.UserProfileService {
		return &MockUserProfileService{}
	}
	userprofileservice.Add("mock", upCreator)

	conf := config.ClientConfig{
		UserProfileServices: map[string]interface{}{},
	}
	loader := defaultLoader(conf, s.registry, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)

	s.Nil(client.UserProfileService)
}

func (s *DefaultLoaderTestSuite) TestLoaderWithNoDefaultUserProfileServices() {
	upCreator := func() decision.UserProfileService {
		return &MockUserProfileService{}
	}
	userprofileservice.Add("mock3", upCreator)

	conf := config.ClientConfig{
		UserProfileServices: map[string]interface{}{"default": "", "services": map[string]interface{}{
			"mock3": map[string]interface{}{},
		}},
	}
	loader := defaultLoader(conf, s.registry, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)
	s.Nil(client.UserProfileService)
}

func (s *DefaultLoaderTestSuite) TestDefaultRegexValidator() {

	scenarios := []struct {
		input    string
		expected bool
	}{
		{"1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_", true},
		{"12sdkKey:datafileAccessToken89", true},
		{"!@#$%^&*()", false},
		{"abc123!", false},
		{"", false},
		{":", false},
		{"abc:def:hij", false},
		{"abc:", false},
	}

	conf := config.NewDefaultConfig()
	validator := regexValidator(conf.Client.SdkKeyRegex)
	for _, scenario := range scenarios {
		s.Equal(scenario.expected, validator(scenario.input))
	}
}

func TestDefaultLoaderTestSuite(t *testing.T) {
	suite.Run(t, new(DefaultLoaderTestSuite))
}
