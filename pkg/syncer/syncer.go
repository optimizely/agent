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
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/optimizely/agent/config"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/rs/zerolog"
)

const (
	// PubSubDefaultChan will be used as default pubsub channel name
	PubSubDefaultChan = "optimizely-sync"
	// PubSubRedis is the name of pubsub type of Redis
	PubSubRedis = "redis"
)

var (
	ncCache   = make(map[string]*RedisSyncer)
	mutexLock = &sync.Mutex{}
)

// Event holds the notification event with it's type
type Event struct {
	Type    notification.Type `json:"type"`
	Message interface{}       `json:"message"`
}

// RedisSyncer defines Redis pubsub configuration
type RedisSyncer struct {
	ctx      context.Context
	Host     string
	Password string
	Database int
	Channel  string
	logger   *zerolog.Logger
	sdkKey   string
}

// NewRedisSyncer returns an instance of RedisNotificationSyncer
func NewRedisSyncer(logger *zerolog.Logger, conf config.SyncConfig, sdkKey string) (*RedisSyncer, error) {
	mutexLock.Lock()
	defer mutexLock.Unlock()

	if nc, found := ncCache[sdkKey]; found {
		return nc, nil
	}

	if !conf.Notification.Enable {
		return nil, errors.New("notification syncer is not enabled")
	}
	if conf.Notification.Default != PubSubRedis {
		return nil, errors.New("redis syncer is not set as default")
	}
	if conf.Pubsub == nil {
		return nil, errors.New("redis config is not given")
	}

	redisConfig, found := conf.Pubsub[PubSubRedis].(map[string]interface{})
	if !found {
		return nil, errors.New("redis pubsub config not found")
	}

	host, ok := redisConfig["host"].(string)
	if !ok {
		return nil, errors.New("redis host not provided in correct format")
	}
	password, ok := redisConfig["password"].(string)
	if !ok {
		return nil, errors.New("redis password not provider in correct format")
	}
	database, ok := redisConfig["database"].(int)
	if !ok {
		return nil, errors.New("redis database not provided in correct format")
	}
	channel, ok := redisConfig["channel"].(string)
	if !ok {
		channel = PubSubDefaultChan
	}

	if logger == nil {
		logger = &zerolog.Logger{}
	}

	nc := &RedisSyncer{
		ctx:      context.Background(),
		Host:     host,
		Password: password,
		Database: database,
		Channel:  channel,
		logger:   logger,
		sdkKey:   sdkKey,
	}
	ncCache[sdkKey] = nc
	return nc, nil
}

func (r *RedisSyncer) WithContext(ctx context.Context) *RedisSyncer {
	r.ctx = ctx
	return r
}

// AddHandler is empty but needed to implement notification.Center interface
func (r *RedisSyncer) AddHandler(_ notification.Type, _ func(interface{})) (int, error) {
	return 0, nil
}

// RemoveHandler is empty but needed to implement notification.Center interface
func (r *RedisSyncer) RemoveHandler(_ int, t notification.Type) error {
	return nil
}

// Send will send the notification to the specified channel in the Redis pubsub
func (r *RedisSyncer) Send(t notification.Type, n interface{}) error {
	event := Event{
		Type:    t,
		Message: n,
	}

	jsonEvent, err := json.Marshal(event)
	if err != nil {
		return err
	}

	client := redis.NewClient(&redis.Options{
		Addr:     r.Host,
		Password: r.Password,
		DB:       r.Database,
	})
	defer client.Close()
	channel := GetChannelForSDKKey(r.Channel, r.sdkKey)

	if err := client.Publish(r.ctx, channel, jsonEvent).Err(); err != nil {
		r.logger.Err(err).Msg("failed to publish json event to pub/sub")
		return err
	}
	return nil
}

func GetChannelForSDKKey(channel, key string) string {
	return fmt.Sprintf("%s-%s", channel, key)
}
