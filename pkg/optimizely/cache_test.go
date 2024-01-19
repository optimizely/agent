/****************************************************************************
 * Copyright 2019,2021,2023-2024 Optimizely, Inc. and contributors          *
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

	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/metrics"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
	"github.com/optimizely/agent/plugins/odpcache"
	odpCacheServices "github.com/optimizely/agent/plugins/odpcache/services"
	"github.com/optimizely/agent/plugins/userprofileservice"
	"github.com/optimizely/agent/plugins/userprofileservice/services"
	sdkconfig "github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/odp/cache"
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
		loader:                mockLoader,
		optlyMap:              cmap.New(),
		userProfileServiceMap: cmap.New(),
		odpCacheMap:           cmap.New(),
		ctx:                   ctx,
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

func (suite *CacheTestSuite) TestNewCache() {
	agentMetricsRegistry := metrics.NewRegistry("")
	sdkMetricsRegistry := NewRegistry(agentMetricsRegistry)

	// To improve coverage
	optlyCache := NewCache(context.Background(), config.AgentConfig{}, sdkMetricsRegistry, nil)
	suite.NotNil(optlyCache)
}

func (suite *CacheTestSuite) TestSetUserProfileService() {
	suite.cache.SetUserProfileService("one", "a")

	actual, ok := suite.cache.userProfileServiceMap.Get("one")
	suite.True(ok)
	suite.Equal("a", actual)

	suite.cache.SetUserProfileService("one", "b")
	actual, ok = suite.cache.userProfileServiceMap.Get("one")
	suite.True(ok)
	suite.Equal("a", actual)
}

func (suite *CacheTestSuite) TestSetODPCache() {
	suite.cache.SetODPCache("one", "a")

	actual, ok := suite.cache.odpCacheMap.Get("one")
	suite.True(ok)
	suite.Equal("a", actual)

	suite.cache.SetODPCache("one", "b")
	actual, ok = suite.cache.odpCacheMap.Get("one")
	suite.True(ok)
	suite.Equal("a", actual)
}

func (suite *CacheTestSuite) TestGetUserProfileServiceJSONErrorCases() {
	conf := config.ClientConfig{
		UserProfileService: map[string]interface{}{"services": map[string]interface{}{
			"in-memory": map[string]interface{}{
				"capacity": []string{"dummy"},
			}},
		},
	}

	// json unmarshal error case
	suite.cache.SetUserProfileService("one", "in-memory")
	userProfileService := getServiceWithType(userProfileServicePlugin, "one", suite.cache.userProfileServiceMap, conf.UserProfileService)
	suite.Nil(userProfileService)

	// json marshal error case
	conf.UserProfileService = map[string]interface{}{"services": map[string]interface{}{
		"in-memory": map[string]interface{}{
			"capacity": make(chan int),
		}},
	}
	userProfileService = getServiceWithType(userProfileServicePlugin, "one", suite.cache.userProfileServiceMap, conf.UserProfileService)
	suite.Nil(userProfileService)
}

func (suite *CacheTestSuite) TestGetODPCacheJSONErrorCases() {
	conf := config.ClientConfig{
		ODP: config.OdpConfig{
			SegmentsCache: map[string]interface{}{"services": map[string]interface{}{
				"in-memory": map[string]interface{}{
					"size": []string{"dummy"},
				}},
			},
		},
	}

	// json unmarshal error case
	suite.cache.SetODPCache("one", "in-memory")
	odpcache := getServiceWithType(odpCachePlugin, "one", suite.cache.odpCacheMap, conf.ODP.SegmentsCache)
	suite.Nil(odpcache)

	// json marshal error case
	conf.ODP.SegmentsCache = map[string]interface{}{"services": map[string]interface{}{
		"in-memory": map[string]interface{}{
			"size": make(chan int),
		}},
	}
	odpcache = getServiceWithType(odpCachePlugin, "one", suite.cache.odpCacheMap, conf.ODP.SegmentsCache)
	suite.Nil(odpcache)
}

func (suite *CacheTestSuite) TestNoUserProfileServicesProvidedInConfig() {
	conf := config.ClientConfig{
		UserProfileService: map[string]interface{}{},
	}
	suite.cache.SetUserProfileService("one", "in-memory")
	userProfileService := getServiceWithType(userProfileServicePlugin, "one", suite.cache.userProfileServiceMap, conf.UserProfileService)
	suite.Nil(userProfileService)
}

func (suite *CacheTestSuite) TestNoODPCacheProvidedInConfig() {
	conf := config.ClientConfig{
		ODP: config.OdpConfig{
			SegmentsCache: map[string]interface{}{},
		},
	}
	suite.cache.SetODPCache("one", "in-memory")
	odpCache := getServiceWithType(odpCachePlugin, "one", suite.cache.odpCacheMap, conf.ODP.SegmentsCache)
	suite.Nil(odpCache)
}

func (suite *CacheTestSuite) TestUPSForSDKKeyNotProvidedInConfig() {
	conf := config.ClientConfig{
		UserProfileService: map[string]interface{}{"default": "in-memory", "services": map[string]interface{}{
			"in-memory": map[string]interface{}{
				"capacity":        0,
				"storageStrategy": "fifo",
			}},
		},
	}
	suite.cache.SetUserProfileService("one", "dummy")
	userProfileService := getServiceWithType(userProfileServicePlugin, "one", suite.cache.userProfileServiceMap, conf.UserProfileService)
	suite.Nil(userProfileService)
}

func (suite *CacheTestSuite) TestODPCacheForSDKKeyNotProvidedInConfig() {
	conf := config.ClientConfig{
		ODP: config.OdpConfig{
			SegmentsCache: map[string]interface{}{"default": "in-memory", "services": map[string]interface{}{
				"in-memory": map[string]interface{}{
					"size":    0,
					"timeout": "0s",
				}},
			},
		},
	}
	suite.cache.SetODPCache("one", "dummy")
	odpCache := getServiceWithType(odpCachePlugin, "one", suite.cache.odpCacheMap, conf.ODP.SegmentsCache)
	suite.Nil(odpCache)
}

func (suite *CacheTestSuite) TestNoCreatorAddedforUPS() {
	conf := config.ClientConfig{
		UserProfileService: map[string]interface{}{"default": "dummy", "services": map[string]interface{}{
			"dummy": map[string]interface{}{
				"capacity":        0,
				"storageStrategy": "fifo",
			}},
		},
	}
	suite.cache.SetUserProfileService("one", "dummy")
	userProfileService := getServiceWithType(userProfileServicePlugin, "one", suite.cache.userProfileServiceMap, conf.UserProfileService)
	suite.Nil(userProfileService)
}

func (suite *CacheTestSuite) TestNoCreatorAddedforODPCache() {
	conf := config.ClientConfig{
		ODP: config.OdpConfig{
			SegmentsCache: map[string]interface{}{"default": "dummy", "services": map[string]interface{}{
				"dummy": map[string]interface{}{
					"size":    0,
					"timeout": "0s",
				}},
			},
		},
	}
	suite.cache.SetODPCache("one", "dummy")
	odpCache := getServiceWithType(odpCachePlugin, "one", suite.cache.odpCacheMap, conf.ODP.SegmentsCache)
	suite.Nil(odpCache)
}

func (suite *CacheTestSuite) TestNilCreatorAddedforUPS() {
	upCreator := func() decision.UserProfileService {
		return nil
	}
	userprofileservice.Add("dummy", upCreator)

	conf := config.ClientConfig{
		UserProfileService: map[string]interface{}{"default": "dummy", "services": map[string]interface{}{
			"dummy": map[string]interface{}{
				"capacity":        0,
				"storageStrategy": "fifo",
			}},
		},
	}
	suite.cache.SetUserProfileService("one", "dummy")
	userProfileService := getServiceWithType(userProfileServicePlugin, "one", suite.cache.userProfileServiceMap, conf.UserProfileService)
	suite.Nil(userProfileService)
}

func (suite *CacheTestSuite) TestNilCreatorAddedforODPCache() {
	cacheCreator := func() cache.Cache {
		return nil
	}
	odpcache.Add("dummy", cacheCreator)

	conf := config.ClientConfig{
		ODP: config.OdpConfig{
			SegmentsCache: map[string]interface{}{"default": "dummy", "services": map[string]interface{}{
				"dummy": map[string]interface{}{
					"size":    0,
					"timeout": "0s",
				}},
			},
		},
	}
	suite.cache.SetODPCache("one", "dummy")
	odpCache := getServiceWithType(odpCachePlugin, "one", suite.cache.odpCacheMap, conf.ODP.SegmentsCache)
	suite.Nil(odpCache)
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

	return &OptlyClient{tc.OptimizelyClient, nil, tc.ForcedVariations, nil, nil}, nil
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

type MockODPCache struct {
	Path string `json:"path"`
	Addr string `json:"addr"`
	Port int    `json:"port"`
}

// Lookup is used to retrieve segments
func (i *MockODPCache) Lookup(key string) (segments interface{}) {
	return nil
}

// Save is used to save segments
func (i *MockODPCache) Save(key string, value interface{}) {
}

// Reset is used to reset segments
func (i *MockODPCache) Reset() {
}

var doOnce sync.Once // required since we only need to read datafile once

type DefaultLoaderTestSuite struct {
	suite.Suite
	registry            *MetricsRegistry
	bp                  *event.BatchEventProcessor
	upsMap, odpCacheMap cmap.ConcurrentMap
	bpFactory           func(options ...event.BPOptionConfig) *event.BatchEventProcessor
	pcFactory           func(sdkKey string, options ...sdkconfig.OptionFunc) SyncedConfigManager
}

func (s *DefaultLoaderTestSuite) SetupTest() {
	// Need the registry to be created only once since it panics if we create gauges with the same name again and again
	doOnce.Do(func() {
		s.registry = &MetricsRegistry{metrics.NewRegistry("")}
	})
	s.upsMap = cmap.New()
	s.odpCacheMap = cmap.New()
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
		UserProfileService: map[string]interface{}{"default": "in-memory", "services": map[string]interface{}{
			"in-memory": map[string]interface{}{
				"capacity":        0,
				"storageStrategy": "fifo",
			}},
		},
		ODP: config.OdpConfig{
			EventsRequestTimeout: 10 * time.Second,
			EventsFlushInterval:  1 * time.Second,
			SegmentsCache: map[string]interface{}{"default": "in-memory", "services": map[string]interface{}{
				"in-memory": map[string]interface{}{
					"size":    100,
					"timeout": "5s",
				}},
			},
			Disable:                true,
			SegmentsRequestTimeout: 10 * time.Second,
		},
	}

	loader := defaultLoader(config.AgentConfig{Client: conf}, s.registry, nil, s.upsMap, s.odpCacheMap, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)

	s.Equal(conf.FlushInterval, s.bp.FlushInterval)
	s.Equal(conf.BatchSize, s.bp.BatchSize)
	s.Equal(conf.QueueSize, s.bp.MaxQueueSize)
	s.Equal(conf.EventURL, s.bp.EventEndPoint)
	s.NotNil(client.UserProfileService)
	s.NotNil(client.odpCache)

	inMemoryUps, ok := client.UserProfileService.(*services.InMemoryUserProfileService)
	s.True(ok)
	s.Equal(0, inMemoryUps.Capacity)
	s.Equal("fifo", inMemoryUps.StorageStrategy)

	inMemoryODPCache, ok := client.odpCache.(*odpCacheServices.InMemoryCache)
	s.True(ok)
	s.Equal(100, inMemoryODPCache.Size)
	s.Equal(5*time.Second, inMemoryODPCache.Timeout.Duration)

	_, err = loader("invalid!")
	s.Error(err)
}

func (s *DefaultLoaderTestSuite) TestUPSAndODPCacheHeaderOverridesDefaultKey() {
	conf := config.ClientConfig{
		FlushInterval: 321 * time.Second,
		BatchSize:     1234,
		QueueSize:     5678,
		EventURL:      "https://localhost/events",
		SdkKeyRegex:   "sdkkey",
		UserProfileService: map[string]interface{}{"default": "", "services": map[string]interface{}{
			"in-memory": map[string]interface{}{
				"capacity":        100,
				"storageStrategy": "fifo",
			}},
		},
		ODP: config.OdpConfig{
			SegmentsCache: map[string]interface{}{"default": "", "services": map[string]interface{}{
				"in-memory": map[string]interface{}{
					"size":    100,
					"timeout": "5s",
				}},
			},
		},
	}

	tmpUPSMap := cmap.New()
	tmpUPSMap.Set("sdkkey", "in-memory")

	tmpOdpCacheMap := cmap.New()
	tmpOdpCacheMap.Set("sdkkey", "in-memory")

	loader := defaultLoader(config.AgentConfig{Client: conf}, s.registry, nil, tmpUPSMap, tmpOdpCacheMap, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)

	s.Equal(conf.FlushInterval, s.bp.FlushInterval)
	s.Equal(conf.BatchSize, s.bp.BatchSize)
	s.Equal(conf.QueueSize, s.bp.MaxQueueSize)
	s.Equal(conf.EventURL, s.bp.EventEndPoint)
	s.NotNil(client.UserProfileService)
	s.NotNil(client.odpCache)

	inMemoryUps, ok := client.UserProfileService.(*services.InMemoryUserProfileService)
	s.True(ok)
	s.Equal(100, inMemoryUps.Capacity)
	s.Equal("fifo", inMemoryUps.StorageStrategy)

	inMemoryODPCache, ok := client.odpCache.(*odpCacheServices.InMemoryCache)
	s.True(ok)
	s.Equal(100, inMemoryODPCache.Size)
	s.Equal(5*time.Second, inMemoryODPCache.Timeout.Duration)
}

func (s *DefaultLoaderTestSuite) TestAddedByDefaultProfileServicesAndODPCache() {
	s.NotNil(userprofileservice.Creators["in-memory"])
	_, ok := (userprofileservice.Creators["in-memory"]()).(*services.InMemoryUserProfileService)
	s.True(ok)

	s.NotNil(userprofileservice.Creators["redis"])
	_, ok = (userprofileservice.Creators["redis"]()).(*services.RedisUserProfileService)
	s.True(ok)

	s.NotNil(userprofileservice.Creators["rest"])
	_, ok = (userprofileservice.Creators["rest"]()).(*services.RestUserProfileService)
	s.True(ok)

	s.NotNil(odpcache.Creators["redis"])
	_, ok = (odpcache.Creators["redis"]()).(*odpCacheServices.RedisCache)
	s.True(ok)

	s.NotNil(odpcache.Creators["in-memory"])
	_, ok = (odpcache.Creators["in-memory"]()).(*odpCacheServices.InMemoryCache)
	s.True(ok)
}

func (s *DefaultLoaderTestSuite) TestFirstSaveConfiguresClientForRedisUPSAndODPCache() {
	conf := config.ClientConfig{
		UserProfileService: map[string]interface{}{"default": "redis", "services": map[string]interface{}{
			"redis": map[string]interface{}{
				"host":     "100",
				"password": "10",
				"database": 1,
			},
		}},
		ODP: config.OdpConfig{
			SegmentsCache: map[string]interface{}{"default": "redis", "services": map[string]interface{}{
				"redis": map[string]interface{}{
					"host":     "100",
					"password": "10",
					"database": 1,
					"timeout":  "1s",
				},
			}},
		},
	}
	loader := defaultLoader(config.AgentConfig{Client: conf}, s.registry, nil, s.upsMap, s.odpCacheMap, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)
	s.NotNil(client.UserProfileService)
	s.NotNil(client.odpCache)
	profile := decision.UserProfile{
		ID: "1",
		ExperimentBucketMap: map[decision.UserDecisionKey]string{
			decision.NewUserDecisionKey("1"): "1",
		},
	}
	// Should initialize redis client on first save call
	client.UserProfileService.Save(profile)
	client.odpCache.Save("1", profile)

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
	} else {
		s.Failf("UserProfileService not registered", "%s DNE in registry", "redis")
	}

	if testRedisODPCache, ok := client.odpCache.(*odpCacheServices.RedisCache); ok {
		s.Equal("100", testRedisODPCache.Address)
		s.Equal("10", testRedisODPCache.Password)
		s.Equal(1, testRedisODPCache.Database)
		s.Equal(1*time.Second, testRedisODPCache.Timeout.Duration)

		// Check if redis client was instantiated with updated config
		s.NotNil(testRedisODPCache.Client)
		s.Equal(testRedisODPCache.Address, testRedisODPCache.Client.Options().Addr)
		s.Equal(testRedisODPCache.Password, testRedisODPCache.Client.Options().Password)
		s.Equal(testRedisODPCache.Database, testRedisODPCache.Client.Options().DB)
		return
	} else {
		s.Failf("ODPCache not registered", "%s DNE in registry", "redis")
	}
}

func (s *DefaultLoaderTestSuite) TestFirstSaveConfiguresLRUCacheForInMemoryCache() {
	conf := config.ClientConfig{
		ODP: config.OdpConfig{
			SegmentsCache: map[string]interface{}{"default": "in-memory", "services": map[string]interface{}{
				"in-memory": map[string]interface{}{
					"size":    100,
					"timeout": "10s",
				},
			}},
		},
	}
	loader := defaultLoader(config.AgentConfig{Client: conf}, s.registry, nil, s.upsMap, s.odpCacheMap, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)
	s.NotNil(client.odpCache)

	profile := decision.UserProfile{
		ID: "1",
		ExperimentBucketMap: map[decision.UserDecisionKey]string{
			decision.NewUserDecisionKey("1"): "1",
		},
	}
	client.odpCache.Save("1", profile)

	if testInMemoryODPCache, ok := client.odpCache.(*odpCacheServices.InMemoryCache); ok {
		s.Equal(100, testInMemoryODPCache.Size)
		s.Equal(10*time.Second, testInMemoryODPCache.Timeout.Duration)

		// Check if lru cache was instantiated
		s.NotNil(testInMemoryODPCache.LRUCache)
		return
	} else {
		s.Failf("ODPCache not registered", "%s DNE in registry", "redis")
	}
}

func (s *DefaultLoaderTestSuite) TestHttpClientInitializesByDefaultRestUPS() {
	conf := config.ClientConfig{
		UserProfileService: map[string]interface{}{"default": "rest", "services": map[string]interface{}{
			"rest": map[string]interface{}{},
		}},
	}
	loader := defaultLoader(config.AgentConfig{Client: conf}, s.registry, nil, s.upsMap, s.odpCacheMap, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)
	s.NotNil(client.UserProfileService)

	if testRestUPS, ok := client.UserProfileService.(*services.RestUserProfileService); ok {
		// Check if rest client was instantiated with updated config
		s.NotNil(testRestUPS.Requester)
		return
	}
	s.Failf("UserProfileService not registered", "%s DNE in registry", "rest")
}

func (s *DefaultLoaderTestSuite) TestLoaderWithValidUserProfileServices() {
	upCreator := func() decision.UserProfileService {
		return &MockUserProfileService{}
	}
	userprofileservice.Add("mock2", upCreator)

	conf := config.ClientConfig{
		UserProfileService: map[string]interface{}{"default": "mock2", "services": map[string]interface{}{
			"mock2": map[string]interface{}{
				"path": "http://test.com",
				"addr": "1.2.1.2-abc",
				"port": 8080,
			},
		}},
	}
	loader := defaultLoader(config.AgentConfig{Client: conf}, s.registry, nil, s.upsMap, s.odpCacheMap, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)

	s.NotNil(client.UserProfileService)
	if mockUPS, ok := client.UserProfileService.(*MockUserProfileService); ok {
		s.Equal("http://test.com", mockUPS.Path)
		s.Equal("1.2.1.2-abc", mockUPS.Addr)
		s.Equal(8080, mockUPS.Port)
		return
	}
	s.Failf("UserProfileService not registered", "%s DNE in registry", "mock2")
}

func (s *DefaultLoaderTestSuite) TestLoaderWithValidODPCache() {
	odpCacheCreator := func() cache.Cache {
		return &MockODPCache{}
	}
	odpcache.Add("mock2", odpCacheCreator)

	conf := config.ClientConfig{
		ODP: config.OdpConfig{
			SegmentsCache: map[string]interface{}{"default": "mock2", "services": map[string]interface{}{
				"mock2": map[string]interface{}{
					"path": "http://test.com",
					"addr": "1.2.1.2-abc",
					"port": 8080,
				},
			}},
		},
	}
	loader := defaultLoader(config.AgentConfig{Client: conf}, s.registry, nil, s.upsMap, s.odpCacheMap, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)

	s.NotNil(client.odpCache)
	if mockODPCache, ok := client.odpCache.(*MockODPCache); ok {
		s.Equal("http://test.com", mockODPCache.Path)
		s.Equal("1.2.1.2-abc", mockODPCache.Addr)
		s.Equal(8080, mockODPCache.Port)
		return
	}
	s.Failf("ODPCache not registered", "%s DNE in registry", "mock2")
}

func (s *DefaultLoaderTestSuite) TestLoaderWithEmptyUserProfileServices() {
	upCreator := func() decision.UserProfileService {
		return &MockUserProfileService{}
	}
	userprofileservice.Add("mock", upCreator)

	conf := config.ClientConfig{
		UserProfileService: map[string]interface{}{},
	}
	loader := defaultLoader(config.AgentConfig{Client: conf}, s.registry, nil, s.upsMap, s.odpCacheMap, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)
	s.Nil(client.UserProfileService)
}

func (s *DefaultLoaderTestSuite) TestLoaderWithEmptyODPCache() {
	odpCacheCreator := func() cache.Cache {
		return &MockODPCache{}
	}
	odpcache.Add("mock", odpCacheCreator)

	conf := config.ClientConfig{
		ODP: config.OdpConfig{
			SegmentsCache: map[string]interface{}{},
		},
	}
	loader := defaultLoader(config.AgentConfig{Client: conf}, s.registry, nil, s.upsMap, s.odpCacheMap, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)
	s.Nil(client.odpCache)
}

func (s *DefaultLoaderTestSuite) TestLoaderWithNoDefaultUserProfileServices() {
	upCreator := func() decision.UserProfileService {
		return &MockUserProfileService{}
	}
	userprofileservice.Add("mock3", upCreator)

	conf := config.ClientConfig{
		UserProfileService: map[string]interface{}{"default": "", "services": map[string]interface{}{
			"mock3": map[string]interface{}{},
		}},
	}
	loader := defaultLoader(config.AgentConfig{Client: conf}, s.registry, nil, s.upsMap, s.odpCacheMap, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)
	s.Nil(client.UserProfileService)
}

func (s *DefaultLoaderTestSuite) TestLoaderWithNoDefaultODPCache() {
	odpCacheCreator := func() cache.Cache {
		return &MockODPCache{}
	}
	odpcache.Add("mock3", odpCacheCreator)

	conf := config.ClientConfig{
		ODP: config.OdpConfig{
			SegmentsCache: map[string]interface{}{"default": "", "services": map[string]interface{}{
				"mock3": map[string]interface{}{},
			}},
		},
	}
	loader := defaultLoader(config.AgentConfig{Client: conf}, s.registry, nil, s.upsMap, s.odpCacheMap, s.pcFactory, s.bpFactory)
	client, err := loader("sdkkey")
	s.NoError(err)
	s.Nil(client.odpCache)
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
