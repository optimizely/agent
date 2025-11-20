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

	"github.com/go-redis/redis/v8"
	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/syncer/pubsub"
	"github.com/optimizely/agent/pkg/utils/redisauth"
	"github.com/rs/zerolog/log"
)

const (
	// PubSubDefaultChan will be used as default pubsub channel name
	PubSubDefaultChan = "optimizely-sync"
	// PubSubRedis is the name of pubsub type of Redis (fire-and-forget)
	PubSubRedis = "redis"
	// PubSubRedisStreams is the name of pubsub type of Redis Streams (persistent)
	PubSubRedisStreams = "redis-streams"
)

type SyncFeatureFlag string

const (
	SyncFeatureFlagNotification SyncFeatureFlag = "sync-feature-flag-notification"
	SyncFeatureFlagDatafile     SyncFeatureFlag = "sync-feature-flag-datafile"
)

type PubSub interface {
	Publish(ctx context.Context, channel string, message interface{}) error
	Subscribe(ctx context.Context, channel string) (chan string, error)
}

func newPubSub(conf config.SyncConfig, featureFlag SyncFeatureFlag) (PubSub, error) {
	var defaultPubSub string

	if featureFlag == SyncFeatureFlagNotification {
		defaultPubSub = conf.Notification.Default
	} else if featureFlag == SyncFeatureFlagDatafile {
		defaultPubSub = conf.Datafile.Default
	} else {
		return nil, errors.New("provided feature flag not supported")
	}

	// Only support auto-detection
	if defaultPubSub == PubSubRedis {
		// Use auto-detection (with fallback to Pub/Sub if detection fails)
		return getPubSubWithAutoDetect(conf)
	}

	return nil, errors.New("pubsub type not supported")
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

	// Support multiple auth field names and env var fallback for security scanning compliance
	password := redisauth.GetPassword(redisConf, "REDIS_PASSWORD")

	databaseVal, found := redisConf["database"]
	if !found {
		return nil, errors.New("pubsub redis database not found")
	}
	// YAML/JSON unmarshals numbers as float64, convert to int
	var database int
	switch v := databaseVal.(type) {
	case int:
		database = v
	case float64:
		database = int(v)
	default:
		return nil, errors.New("pubsub redis database not valid, database must be numeric")
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

	// Support multiple auth field names and env var fallback for security scanning compliance
	password := redisauth.GetPassword(redisConf, "REDIS_PASSWORD")

	databaseVal, found := redisConf["database"]
	if !found {
		return nil, errors.New("pubsub redis database not found")
	}
	// YAML/JSON unmarshals numbers as float64, convert to int
	var database int
	switch v := databaseVal.(type) {
	case int:
		database = v
	case float64:
		database = int(v)
	default:
		return nil, errors.New("pubsub redis database not valid, database must be numeric")
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

// getPubSubWithAutoDetect creates a PubSub instance using Redis version auto-detection
// Falls back to Pub/Sub (safe default) if detection fails for any reason
func getPubSubWithAutoDetect(conf config.SyncConfig) (PubSub, error) {
	pubsubConf, found := conf.Pubsub[PubSubRedis]
	if !found {
		return nil, errors.New("pubsub redis config not found")
	}

	redisConf, ok := pubsubConf.(map[string]interface{})
	if !ok {
		return nil, errors.New("pubsub redis config not valid")
	}

	// Get connection details
	hostVal, found := redisConf["host"]
	if !found {
		return nil, errors.New("pubsub redis host not found")
	}
	host, ok := hostVal.(string)
	if !ok {
		return nil, errors.New("pubsub redis host not valid, host must be string")
	}

	password := redisauth.GetPassword(redisConf, "REDIS_PASSWORD")

	databaseVal, found := redisConf["database"]
	if !found {
		return nil, errors.New("pubsub redis database not found")
	}
	var database int
	switch v := databaseVal.(type) {
	case int:
		database = v
	case float64:
		database = int(v)
	default:
		return nil, errors.New("pubsub redis database not valid, database must be numeric")
	}

	// Create temporary Redis client for version detection
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       database,
	})
	defer client.Close()

	// Attempt version detection
	log.Info().Msg("Auto-detecting Redis version to choose best notification implementation...")
	if pubsub.SupportsRedisStreams(client) {
		// Redis >= 5.0 - Use Streams
		return getPubSubRedisStreams(conf)
	}

	// Redis < 5.0 or detection failed - Use Pub/Sub (safe default)
	return getPubSubRedis(conf)
}
