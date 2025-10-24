/****************************************************************************
 * Copyright 2025, Optimizely, Inc. and contributors                        *
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

	"github.com/go-redis/redis/v8"
	"github.com/optimizely/agent/plugins/cmabcache"
	"github.com/optimizely/agent/plugins/utils"
	"github.com/optimizely/go-sdk/v2/pkg/cache"
	"github.com/rs/zerolog/log"
)

var ctx = context.Background()

// RedisCache represents the redis implementation of Cache interface for CMAB
type RedisCache struct {
	Client   *redis.Client
	Address  string         `json:"host"`
	Password string         `json:"password"`
	Database int            `json:"database"`
	Timeout  utils.Duration `json:"timeout"`
}

// Lookup is used to retrieve cached CMAB decisions
func (r *RedisCache) Lookup(key string) interface{} {
	// This is required in both lookup and save since an old redis instance can also be used
	if r.Client == nil {
		r.initClient()
	}

	if key == "" {
		return nil
	}

	// Check if decision exists
	result, getError := r.Client.Get(ctx, key).Result()
	if getError != nil {
		if getError != redis.Nil {
			log.Error().Err(getError).Msg("Failed to get CMAB decision from Redis")
		}
		return nil
	}

	// Unmarshal the cached decision
	// The CMAB cache stores the decision object directly
	var cachedDecision interface{}
	err := json.Unmarshal([]byte(result), &cachedDecision)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal CMAB decision from Redis")
		return nil
	}
	return cachedDecision
}

// Save is used to save CMAB decisions
func (r *RedisCache) Save(key string, value interface{}) {
	// This is required in both lookup and save since an old redis instance can also be used
	if r.Client == nil {
		r.initClient()
	}

	if key == "" {
		return
	}

	// Marshal the decision value
	finalDecision, err := json.Marshal(value)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal CMAB decision for Redis")
		return
	}

	// Save to Redis with TTL
	if setError := r.Client.Set(ctx, key, finalDecision, r.Timeout.Duration).Err(); setError != nil {
		log.Error().Err(setError).Msg("Failed to save CMAB decision to Redis")
	}
}

// Remove is used to remove a specific CMAB decision from cache
func (r *RedisCache) Remove(key string) {
	// This is required since remove can be called before lookup and save
	if r.Client == nil {
		r.initClient()
	}

	if key == "" {
		return
	}

	if delError := r.Client.Del(ctx, key).Err(); delError != nil {
		log.Error().Err(delError).Msg("Failed to remove CMAB decision from Redis")
	}
}

// Reset is used to reset all CMAB decisions
func (r *RedisCache) Reset() {
	// This is required since reset can be called before lookup and save
	if r.Client == nil {
		r.initClient()
	}

	if r.Client != nil {
		if flushError := r.Client.FlushDB(ctx).Err(); flushError != nil {
			log.Error().Err(flushError).Msg("Failed to flush CMAB cache in Redis")
		}
	}
}

func (r *RedisCache) initClient() {
	r.Client = redis.NewClient(&redis.Options{
		Addr:     r.Address,
		Password: r.Password,
		DB:       r.Database,
	})
}

func init() {
	redisCacheCreator := func() cache.Cache {
		return &RedisCache{}
	}
	cmabcache.Add("redis", redisCacheCreator)
}
