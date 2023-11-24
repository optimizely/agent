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

// Package pubsub provides pubsub functionality for the agent syncer
package pubsub

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	Host     string
	Password string
	Database int
}

func (r *Redis) Publish(ctx context.Context, channel string, message interface{}) error {
	client := redis.NewClient(&redis.Options{
		Addr:     r.Host,
		Password: r.Password,
		DB:       r.Database,
	})
	defer client.Close()

	return client.Publish(ctx, channel, message).Err()
}

func (r *Redis) Subscribe(ctx context.Context, channel string) (chan string, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     r.Host,
		Password: r.Password,
		DB:       r.Database,
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
