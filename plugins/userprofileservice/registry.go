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
	"fmt"

	"github.com/optimizely/go-sdk/v2/pkg/decision"
)

// Creator type defines a function for creating an instance of a UserProfileService
type Creator func() decision.UserProfileService

// Creators stores the mapping of Creator against userProfileServiceName
var Creators = map[string]Creator{}

// Add registers a creator against userProfileServiceName
func Add(userProfileServiceName string, creator Creator) {
	if _, ok := Creators[userProfileServiceName]; ok {
		panic(fmt.Sprintf("UserProfileService with name %q already exists", userProfileServiceName))
	}
	Creators[userProfileServiceName] = creator
}
