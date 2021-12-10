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

// Package memory //
package memory

import (
	"sync"

	"github.com/optimizely/go-sdk/pkg/decision"
)

// InMemoryUserProfileService represents the in-memory implementation of UserProfileService interface
type InMemoryUserProfileService struct {
	profiles map[string]decision.UserProfile
	lock     sync.RWMutex
}

// NewInMemoryUserProfileService returns new instance of InMemoryUserProfileService
func NewInMemoryUserProfileService() *InMemoryUserProfileService {
	return &InMemoryUserProfileService{
		profiles: make(map[string]decision.UserProfile),
	}
}

// Lookup is used to retrieve past bucketing decisions for users
func (u *InMemoryUserProfileService) Lookup(userID string) decision.UserProfile {
	var profile decision.UserProfile
	u.lock.RLock()
	defer u.lock.RUnlock()
	if userProfile, ok := u.profiles[userID]; ok {
		profile = userProfile
	}
	return profile
}

// Save is used to save bucketing decisions for users
func (u *InMemoryUserProfileService) Save(profile decision.UserProfile) {
	if profile.ID == "" {
		return
	}

	u.lock.Lock()
	defer u.lock.Unlock()
	u.profiles[profile.ID] = profile
}
