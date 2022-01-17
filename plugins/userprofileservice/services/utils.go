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
	"encoding/json"
	"fmt"

	"github.com/optimizely/go-sdk/pkg/decision"
)

// convertUserProfileToMap converts User Profile object to map
func convertUserProfileToMap(userProfile decision.UserProfile) map[string]interface{} {
	experimentBucketMap := map[string]interface{}{}
	for k, v := range userProfile.ExperimentBucketMap {
		experimentBucketMap[k.ExperimentID] = map[string]interface{}{k.Field: v}
	}
	return map[string]interface{}{
		"user_id":               userProfile.ID,
		"experiment_bucket_map": experimentBucketMap,
	}
}

// convertToUserProfile converts map to User Profile object
func convertToUserProfile(profileDict map[string]interface{}) decision.UserProfile {
	userProfile := decision.UserProfile{}
	userID, ok := profileDict["user_id"].(string)
	j, _ := json.Marshal(profileDict)
	fmt.Printf("%v", string(j))
	if !ok {
		return userProfile
	}
	if experimentBucketMap, ok := profileDict["experiment_bucket_map"].(map[string]interface{}); ok {
		userProfile.ID = userID
		userProfile.ExperimentBucketMap = make(map[decision.UserDecisionKey]string)
		for k, v := range experimentBucketMap {
			decisionKey := decision.NewUserDecisionKey(k)
			if bucketMap, ok := v.(map[string]interface{}); ok {
				userProfile.ExperimentBucketMap[decisionKey] = bucketMap[decisionKey.Field].(string)
			}
		}
	}
	return userProfile
}
