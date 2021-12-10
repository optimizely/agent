/****************************************************************************
 * Copyright 2021, Optimizely, Inc. and contributors                        *
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

// Package userprofileservice //
package userprofileservice

import (
	"sync"

	"github.com/optimizely/go-sdk/pkg/decision"
)

// userProfileServices stores the mapping of UserProfileServices against sdkKey and userProfileServiceName
var userProfileServices = map[string]map[string]Creator{}
var lock sync.RWMutex

// Creator type defines a function for creating an instance of a UserProfileService
type Creator func() decision.UserProfileService

// AddUserProfileService maps userProfileService against sdkKey and userProfileServiceName
// Both sdkKey and userProfileServiceName should be unique. Also, userProfileServiceName should match one of the
// user profile services provided in `client.userProfileServices` in `config.yaml`
func AddUserProfileService(sdkKey, userProfileServiceName string, profileService Creator) {
	lock.Lock()
	defer lock.Unlock()

	if _, ok := userProfileServices[sdkKey]; !ok {
		userProfileServices[sdkKey] = map[string]Creator{}
	}
	userProfileServices[sdkKey][userProfileServiceName] = profileService
}

// GetUserProfileService returns userProfileService mapped against the sdkKey and userProfileServiceName
func GetUserProfileService(sdkKey, userProfileServiceName string) (profileService Creator) {
	lock.RLock()
	defer lock.RUnlock()
	return userProfileServices[sdkKey][userProfileServiceName]
}
