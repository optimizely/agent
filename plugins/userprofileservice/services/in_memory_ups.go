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
	Capacity int `json:"capacity"`
	// Order defines the priority order. Supported values include fifo and lifo.
	Order               string `json:"order"`
	ProfilesMap         map[string]decision.UserProfile
	fifoOrderedProfiles chan string
	lifoOrderedProfiles []string
	lock                sync.RWMutex
	isReady             bool
}

// Lookup is used to retrieve past bucketing decisions for users
func (u *InMemoryUserProfileService) Lookup(userID string) (profile decision.UserProfile) {
	u.lock.RLock()
	defer u.lock.RUnlock()

	// Check if UPS is ready
	if !u.isReady {
		return profile
	}

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
			switch u.Order {
			case "lifo":
				u.lifoOrderedProfiles = []string{}
			default:
				// fifo by default
				u.fifoOrderedProfiles = make(chan string, u.Capacity)
			}
			u.ProfilesMap = make(map[string]decision.UserProfile, u.Capacity)
		} else {
			u.ProfilesMap = map[string]decision.UserProfile{}
		}
		u.isReady = true
	}

	// check if profile does not exist already
	if _, ok := u.ProfilesMap[profile.ID]; !ok {
		if u.Capacity > 0 {
			// Check if capacity has reached, if so, pop the entry from ordered list and map
			if len(u.ProfilesMap) == u.Capacity {
				var oldProfile string
				// pop entry from ordered list
				switch u.Order {
				case "lifo":
					n := len(u.lifoOrderedProfiles) - 1
					oldProfile = u.lifoOrderedProfiles[n]
					u.lifoOrderedProfiles[n] = "" // Erase element (write zero value)
					u.lifoOrderedProfiles = u.lifoOrderedProfiles[:n]
				default:
					// fifo by default
					oldProfile = <-u.fifoOrderedProfiles
				}
				// remove entry from map
				delete(u.ProfilesMap, oldProfile)
			}

			// Push new entry to ordered list
			switch u.Order {
			case "lifo":
				u.lifoOrderedProfiles = append(u.lifoOrderedProfiles, profile.ID)
			default:
				// fifo by default
				u.fifoOrderedProfiles <- profile.ID
			}
		}
	}
	// Save new profile to map
	u.ProfilesMap[profile.ID] = profile
}

func init() {
	inMemoryUPSCreator := func() decision.UserProfileService {
		return &InMemoryUserProfileService{}
	}
	userprofileservice.Add("in-memory", inMemoryUPSCreator)
}
