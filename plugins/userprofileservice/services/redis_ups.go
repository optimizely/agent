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

	"github.com/go-redis/redis/v8"
	"github.com/optimizely/agent/plugins/userprofileservice"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/rs/zerolog/log"
)

var ctx = context.Background()

// RedisUserProfileService represents the redis implementation of UserProfileService interface
type RedisUserProfileService struct {
	Client     *redis.Client
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

	if u.Client == nil || userID == "" {
		return profile
	}

	// Check if profile exists
	result, getError := u.Client.Get(ctx, userID).Result()
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
	profile.ID = userID
	for experimentID, bucketMap := range experimentBucketMap {
		decisionKey := decision.UserDecisionKey{
			ExperimentID: experimentID,
		}
		if finalBucketMap, ok := bucketMap.(map[string]interface{}); ok {
			for field, variationKey := range finalBucketMap {
				if strVariationKey, ok := variationKey.(string); ok {
					decisionKey.Field = field
					profile.ExperimentBucketMap[decisionKey] = strVariationKey
				}
			}
		}
	}
	return profile
}

// Save is used to save bucketing decisions for users
func (u *RedisUserProfileService) Save(profile decision.UserProfile) {

	if u.Client == nil {
		u.Client = redis.NewClient(&redis.Options{
			Addr:     u.Address,
			Password: u.Password,
			DB:       u.Database,
		})
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
		if setError := u.Client.Set(ctx, profile.ID, finalProfile, u.Expiration).Err(); setError != nil {
			log.Error().Msg(setError.Error())
		}
	}
}

func init() {
	redisUPSCreator := func() decision.UserProfileService {
		return &RedisUserProfileService{
			Expiration: 0 * time.Second,
		}
	}
	userprofileservice.Add("redis", redisUPSCreator)
}
