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

// Package services //
package services

import (
	"sync"

	"github.com/optimizely/agent/plugins/userprofileservice"
	"github.com/optimizely/go-sdk/pkg/decision"
)

// InMemoryUserProfileService represents the in-memory implementation of UserProfileService interface
type InMemoryUserProfileService struct {
	Capacity        int `json:"capacity"`
	orderedProfiles chan decision.UserProfile
	isReady         bool
	ProfilesMap     map[string]decision.UserProfile
	lock            sync.RWMutex
}

// Lookup is used to retrieve past bucketing decisions for users
func (u *InMemoryUserProfileService) Lookup(userID string) decision.UserProfile {
	var profile decision.UserProfile
	// Check if UPS is ready
	if !u.isReady {
		return profile
	}
	u.lock.RLock()
	defer u.lock.RUnlock()
	if userProfile, ok := u.ProfilesMap[userID]; ok {
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

	// initialize properties with provided capacity
	if !u.isReady {
		// Initialize with capacity only if required
		if u.Capacity > 0 {
			u.orderedProfiles = make(chan decision.UserProfile, u.Capacity)
			u.ProfilesMap = make(map[string]decision.UserProfile, u.Capacity)
		} else {
			u.ProfilesMap = map[string]decision.UserProfile{}
		}
		u.isReady = true
	}

	// check if profile does not exist already
	if _, ok := u.ProfilesMap[profile.ID]; !ok {
		// Check if capacity has reached, if so, pop the oldest entry
		if u.Capacity > 0 && len(u.ProfilesMap) == u.Capacity {
			select {
			// pop entry from ordered list
			case p := <-u.orderedProfiles:
				// remove entry from map aswell
				delete(u.ProfilesMap, p.ID)
			default:
			}
		}
		// Only push to channel if needed
		if u.Capacity > 0 {
			// push new entry to ordered list
			u.orderedProfiles <- profile
		}
	}
	// Save new profile to map
	u.ProfilesMap[profile.ID] = profile
}

func init() {
	inMemoryUPSCreator := func() decision.UserProfileService {
		return &InMemoryUserProfileService{
			ProfilesMap:     make(map[string]decision.UserProfile),
			orderedProfiles: make(chan decision.UserProfile),
		}
	}
	userprofileservice.Add("in-memory", inMemoryUPSCreator)
}
