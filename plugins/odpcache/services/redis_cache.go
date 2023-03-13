/****************************************************************************
 * Copyright 2023, Optimizely, Inc. and contributors                        *
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

// Package services //
package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/optimizely/agent/plugins/odpcache"
	"github.com/optimizely/go-sdk/pkg/odp/cache"
	"github.com/rs/zerolog/log"
)

var ctx = context.Background()

// RedisCache represents the redis implementation of Cache interface
type RedisCache struct {
	Client     *redis.Client
	Expiration time.Duration
	Address    string `json:"host"`
	Password   string `json:"password"`
	Database   int    `json:"database"`
	Timeout    int    `json:"timeout"`
}

// Lookup is used to retrieve segments
func (r *RedisCache) Lookup(key string) (segments interface{}) {
	// This is required in both lookup and save since an old redis instance can also be used
	if r.Client == nil {
		r.initClient()
	}

	if key == "" {
		return
	}

	// Check if segments exist
	result, getError := r.Client.Get(ctx, key).Result()
	if getError != nil {
		log.Error().Msg(getError.Error())
		return
	}

	// Check if result was unmarshalled successfully
	err := json.Unmarshal([]byte(result), &segments)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	return segments
}

// Save is used to save segments
func (r *RedisCache) Save(key string, value interface{}) {

	// This is required in both lookup and save since an old redis instance can also be used
	if r.Client == nil {
		r.initClient()
	}

	if key == "" {
		return
	}

	if finalSegments, err := json.Marshal(value); err == nil {
		// Log error message if something went wrong
		if setError := r.Client.Set(ctx, key, finalSegments, r.Expiration).Err(); setError != nil {
			log.Error().Msg(setError.Error())
		}
	}
}

// Reset is used to reset segments
func (r *RedisCache) Reset() {
	if r.Client != nil {
		r.Client.FlushDB(ctx)
	}
}

func (r *RedisCache) initClient() {
	r.Client = redis.NewClient(&redis.Options{
		Addr:     r.Address,
		Password: r.Password,
		DB:       r.Database,
	})
	r.Expiration = time.Duration(r.Timeout * int(time.Second))
}

func init() {
	redisCacheCreator := func() cache.Cache {
		return &RedisCache{}
	}
	odpcache.Add("redis", redisCacheCreator)
}
