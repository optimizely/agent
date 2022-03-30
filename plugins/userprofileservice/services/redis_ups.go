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

// Package services //
package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	"github.com/optimizely/agent/plugins/userprofileservice"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/rs/zerolog/log"
)

// RedisUserProfileService represents the redis implementation of UserProfileService interface
type RedisUserProfileService struct {
	Ctx        context.Context
	Client     *redis.Client
	Locker     *redislock.Client
	Expiration time.Duration
	Address    string `json:"host"`
	Password   string `json:"password"`
	Database   int    `json:"database"`
}

// Lookup is used to retrieve past bucketing decisions for users
func (u *RedisUserProfileService) Lookup(userID string) (profile decision.UserProfile) {
	profile = decision.UserProfile{
		ID:                  "",
		ExperimentBucketMap: make(map[decision.UserDecisionKey]string),
	}

	// This is required in both lookup and save since an old redis instance can also be used
	if u.Client == nil {
		u.initClient()
	}

	if userID == "" {
		return profile
	}

	// Check if profile exists
	result, getError := u.Client.Get(u.Ctx, userID).Result()
	if getError != nil {
		log.Error().Msg(getError.Error())
		return profile
	}

	// Check if result was unmarshalled successfully
	experimentBucketMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(result), &experimentBucketMap)
	if err != nil {
		log.Error().Msg(err.Error())
		return profile
	}

	// Converting result to profile
	return convertToUserProfile(map[string]interface{}{userIDKey: userID, experimentBucketMapKey: experimentBucketMap}, userIDKey)
}

// Save is used to save bucketing decisions for users
func (u *RedisUserProfileService) Save(profile decision.UserProfile) {

	// This is required in both lookup and save since an old redis instance can also be used
	if u.Client == nil {
		u.initClient()
	}

	if profile.ID == "" {
		return
	}

	experimentBucketMap := map[string]interface{}{}
	for k, v := range profile.ExperimentBucketMap {
		experimentBucketMap[k.ExperimentID] = map[string]string{k.Field: v}
	}

	if finalProfile, err := json.Marshal(experimentBucketMap); err == nil {
		// Log error message if something went wrong
		// Try to obtain lock.

		// TODO: need to discuss the time duration for the lock
		lock, err := u.Locker.Obtain(u.Ctx, profile.ID, 100*time.Millisecond, nil)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}
		// Release lock after setting value
		defer lock.Release(u.Ctx)
		// Safely set the new value
		if setError := u.Client.Set(u.Ctx, profile.ID, finalProfile, u.Expiration).Err(); setError != nil {
			log.Error().Msg(setError.Error())
		}
	}
}

// AddContext is used to set context in RedisUserProfileService
func (u *RedisUserProfileService) AddContext(ctx context.Context) {
	u.Ctx = ctx
}

func (u *RedisUserProfileService) initClient() {
	u.Client = redis.NewClient(&redis.Options{
		Addr:     u.Address,
		Password: u.Password,
		DB:       u.Database,
	})
	// Create a new lock client.
	u.Locker = redislock.New(u.Client)
}

func init() {
	redisUPSCreator := func() decision.UserProfileService {
		return &RedisUserProfileService{
			Expiration: 0 * time.Second,
		}
	}
	userprofileservice.Add("redis", redisUPSCreator)
}
