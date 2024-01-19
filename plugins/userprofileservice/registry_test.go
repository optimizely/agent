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

// Package userprofileservice //
package userprofileservice

import (
	"testing"

	"github.com/optimizely/go-sdk/v2/pkg/decision"
	"github.com/stretchr/testify/assert"
)

type mockUserProfileService struct {
}

// Lookup is used to retrieve past bucketing decisions for users
func (u *mockUserProfileService) Lookup(userID string) decision.UserProfile {
	return decision.UserProfile{}
}

// Save is used to save bucketing decisions for users
func (u *mockUserProfileService) Save(profile decision.UserProfile) {
}

func TestAdd(t *testing.T) {
	mockUPSCreator := func() decision.UserProfileService {
		return &mockUserProfileService{}
	}

	Add("mock", mockUPSCreator)
	creator := Creators["mock"]()
	if _, ok := creator.(*mockUserProfileService); !ok {
		assert.Fail(t, "Cannot convert to type InMemoryUserProfileService")
	}
}

func TestDuplicateKeys(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			assert.Fail(t, "Should have recovered")
		}
	}()

	mockUPSCreator := func() decision.UserProfileService {
		return &mockUserProfileService{}
	}

	Add("mock", mockUPSCreator)
	Add("mock", mockUPSCreator)
	assert.Fail(t, "Should have panicked")
}

func TestDoesNotExist(t *testing.T) {
	dne := Creators["DNE"]
	assert.Nil(t, dne)
}
