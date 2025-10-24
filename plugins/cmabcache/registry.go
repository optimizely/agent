/****************************************************************************
 * Copyright 2025, Optimizely, Inc. and contributors                        *
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

// Package cmabcache //
package cmabcache

import (
	"fmt"

	"github.com/optimizely/go-sdk/v2/pkg/cache"
)

// Creator type defines a function for creating an instance of a Cache
type Creator func() cache.Cache

// Creators stores the mapping of Creator against cmabCacheName
var Creators = map[string]Creator{}

// Add registers a creator against cmabCacheName
func Add(cmabCacheName string, creator Creator) {
	if _, ok := Creators[cmabCacheName]; ok {
		panic(fmt.Sprintf("CMAB Cache with name %q already exists", cmabCacheName))
	}
	Creators[cmabCacheName] = creator
}
