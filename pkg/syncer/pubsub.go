/****************************************************************************
 * Copyright 2023 Optimizely, Inc. and contributors                         *
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

// Package syncer provides synchronization across Agent nodes
package syncer

import (
	"context"
	"errors"
	"time"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/syncer/pubsub"
)

const (
	// PubSubDefaultChan will be used as default pubsub channel name
	PubSubDefaultChan = "optimizely-sync"
	// PubSubRedis is the name of pubsub type of Redis (fire-and-forget)
	PubSubRedis = "redis"
	// PubSubRedisStreams is the name of pubsub type of Redis Streams (persistent)
	PubSubRedisStreams = "redis-streams"
)

type SycnFeatureFlag string

const (
	SyncFeatureFlagNotificaiton SycnFeatureFlag = "sync-feature-flag-notification"
	SycnFeatureFlagDatafile     SycnFeatureFlag = "sync-feature-flag-datafile"
)

type PubSub interface {
	Publish(ctx context.Context, channel string, message interface{}) error
	Subscribe(ctx context.Context, channel string) (chan string, error)
}

func newPubSub(conf config.SyncConfig, featureFlag SycnFeatureFlag) (PubSub, error) {
	if featureFlag == SyncFeatureFlagNotificaiton {
		if conf.Notification.Default == PubSubRedis {
			return getPubSubRedis(conf)
		} else if conf.Notification.Default == PubSubRedisStreams {
			return getPubSubRedisStreams(conf)
		} else {
			return nil, errors.New("pubsub type not supported")
		}
	} else if featureFlag == SycnFeatureFlagDatafile {
		if conf.Datafile.Default == PubSubRedis {
			return getPubSubRedis(conf)
		} else if conf.Datafile.Default == PubSubRedisStreams {
			return getPubSubRedisStreams(conf)
		} else {
			return nil, errors.New("pubsub type not supported")
		}
	}
	return nil, errors.New("provided feature flag not supported")
}

func getPubSubRedis(conf config.SyncConfig) (PubSub, error) {
	pubsubConf, found := conf.Pubsub[PubSubRedis]
	if !found {
		return nil, errors.New("pubsub redis config not found")
	}

	redisConf, ok := pubsubConf.(map[string]interface{})
	if !ok {
		return nil, errors.New("pubsub redis config not valid")
	}

	hostVal, found := redisConf["host"]
	if !found {
		return nil, errors.New("pubsub redis host not found")
	}
	host, ok := hostVal.(string)
	if !ok {
		return nil, errors.New("pubsub redis host not valid, host must be string")
	}

	passwordVal, found := redisConf["password"]
	if !found {
		return nil, errors.New("pubsub redis password not found")
	}
	password, ok := passwordVal.(string)
	if !ok {
		return nil, errors.New("pubsub redis password not valid, password must be string")
	}

	databaseVal, found := redisConf["database"]
	if !found {
		return nil, errors.New("pubsub redis database not found")
	}
	database, ok := databaseVal.(int)
	if !ok {
		return nil, errors.New("pubsub redis database not valid, database must be int")
	}

	// Return original Redis pub/sub implementation (fire-and-forget)
	return &pubsub.Redis{
		Host:     host,
		Password: password,
		Database: database,
	}, nil
}

func getPubSubRedisStreams(conf config.SyncConfig) (PubSub, error) {
	pubsubConf, found := conf.Pubsub[PubSubRedis]
	if !found {
		return nil, errors.New("pubsub redis config not found")
	}

	redisConf, ok := pubsubConf.(map[string]interface{})
	if !ok {
		return nil, errors.New("pubsub redis config not valid")
	}

	hostVal, found := redisConf["host"]
	if !found {
		return nil, errors.New("pubsub redis host not found")
	}
	host, ok := hostVal.(string)
	if !ok {
		return nil, errors.New("pubsub redis host not valid, host must be string")
	}

	passwordVal, found := redisConf["password"]
	if !found {
		return nil, errors.New("pubsub redis password not found")
	}
	password, ok := passwordVal.(string)
	if !ok {
		return nil, errors.New("pubsub redis password not valid, password must be string")
	}

	databaseVal, found := redisConf["database"]
	if !found {
		return nil, errors.New("pubsub redis database not found")
	}
	database, ok := databaseVal.(int)
	if !ok {
		return nil, errors.New("pubsub redis database not valid, database must be int")
	}

	// Parse optional Redis Streams configuration parameters
	batchSize := getIntFromConfig(redisConf, "batch_size", 10)
	flushInterval := getDurationFromConfig(redisConf, "flush_interval", 5*time.Second)
	maxRetries := getIntFromConfig(redisConf, "max_retries", 3)
	retryDelay := getDurationFromConfig(redisConf, "retry_delay", 100*time.Millisecond)
	maxRetryDelay := getDurationFromConfig(redisConf, "max_retry_delay", 5*time.Second)
	connTimeout := getDurationFromConfig(redisConf, "connection_timeout", 10*time.Second)

	// Return Redis Streams implementation with configuration
	return &pubsub.RedisStreams{
		Host:          host,
		Password:      password,
		Database:      database,
		BatchSize:     batchSize,
		FlushInterval: flushInterval,
		MaxRetries:    maxRetries,
		RetryDelay:    retryDelay,
		MaxRetryDelay: maxRetryDelay,
		ConnTimeout:   connTimeout,
	}, nil
}

// getIntFromConfig safely extracts an integer value from config map with default fallback
func getIntFromConfig(config map[string]interface{}, key string, defaultValue int) int {
	if val, found := config[key]; found {
		if intVal, ok := val.(int); ok {
			return intVal
		}
	}
	return defaultValue
}

// getDurationFromConfig safely extracts a duration value from config map with default fallback
func getDurationFromConfig(config map[string]interface{}, key string, defaultValue time.Duration) time.Duration {
	if val, found := config[key]; found {
		if strVal, ok := val.(string); ok {
			if duration, err := time.ParseDuration(strVal); err == nil {
				return duration
			}
		}
	}
	return defaultValue
}
