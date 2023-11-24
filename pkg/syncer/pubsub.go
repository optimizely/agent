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

	"github.com/go-redis/redis/v8"
	"github.com/optimizely/agent/config"
)

const (
	// PubSubDefaultChan will be used as default pubsub channel name
	PubSubDefaultChan = "optimizely-sync"
	// PubSubRedis is the name of pubsub type of Redis
	PubSubRedis = "redis"
)

type PubSub interface {
	Publish(ctx context.Context, channel string, message interface{}) error
	Subscribe(ctx context.Context, channel string) (chan string, error)
}

func newPubSub(conf config.SyncConfig) (PubSub, error) {
	if conf.Notification.Default == PubSubRedis {
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

		return &pubsubRedis{
			host:     host,
			password: password,
			database: database,
		}, nil
	}
	return nil, errors.New("pubsub type not supported")
}

type pubsubRedis struct {
	host     string
	password string
	database int
}

func (r *pubsubRedis) Publish(ctx context.Context, channel string, message interface{}) error {
	client := redis.NewClient(&redis.Options{
		Addr:     r.host,
		Password: r.password,
		DB:       r.database,
	})
	defer client.Close()

	return client.Publish(ctx, channel, message).Err()
}

func (r *pubsubRedis) Subscribe(ctx context.Context, channel string) (chan string, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     r.host,
		Password: r.password,
		DB:       r.database,
	})

	// Subscribe to a Redis channel
	pubsub := client.Subscribe(ctx, channel)

	ch := make(chan string)

	go func() {
		for {
			select {
			case <-ctx.Done():
				pubsub.Close()
				client.Close()
				close(ch)
				return
			default:
				msg, err := pubsub.ReceiveMessage(ctx)
				if err != nil {
					continue
				}

				ch <- msg.Payload

			}
		}
	}()
	return ch, nil
}
