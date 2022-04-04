/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                        *
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

// Package datafilecacheservice //
package datafilecacheservice

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

// DatafileCacheService represents interface for datafileCacheService
type DatafileCacheService interface {
	GetDatafileFromCacheService(ctx context.Context, sdkKey string) string
	SetDatafileInCacheService(ctx context.Context, sdkKey, datafile string)
}

// RedisCacheService represents the redis implementation of DatafileCacheService interface
type RedisCacheService struct {
	Client   *redis.Client
	Address  string `json:"host"`
	Password string `json:"password"`
	Database int    `json:"database"`
}

// GetDatafileFromCacheService returns the saved datafile from the cache service
func (r *RedisCacheService) GetDatafileFromCacheService(ctx context.Context, sdkKey string) string {
	if r.Client == nil {
		r.initClient()
	}
	datafile, err := r.Client.Get(ctx, sdkKey).Result()
	if err != nil {
		log.Error().Msg(err.Error())
		return ""
	}
	return datafile
}

// SetDatafileInCacheService saves the datafile in the cache service
func (r *RedisCacheService) SetDatafileInCacheService(ctx context.Context, sdkKey, datafile string) {
	if r.Client == nil {
		r.initClient()
	}
	if setError := r.Client.Set(ctx, sdkKey, datafile, 0).Err(); setError != nil {
		log.Error().Msg(setError.Error())
	}
}

func (r *RedisCacheService) initClient() {
	r.Client = redis.NewClient(&redis.Options{
		Addr:     r.Address,
		Password: r.Password,
		DB:       r.Database,
	})
}
