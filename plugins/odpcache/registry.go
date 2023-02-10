/****************************************************************************
 * Copyright 2023, Optimizely, Inc. and contributors                        *
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

// Package odpcache //
package odpcache

import (
	"fmt"

	"github.com/optimizely/go-sdk/pkg/odp/cache"
)

// Creator type defines a function for creating an instance of a Cache
type Creator func() cache.Cache

// Creators stores the mapping of Creator against odpCacheName
var Creators = map[string]Creator{}

// Add registers a creator against odpCacheName
func Add(odpCacheName string, creator Creator) {
	if _, ok := Creators[odpCacheName]; ok {
		panic(fmt.Sprintf("ODP Cache with name %q already exists", odpCacheName))
	}
	Creators[odpCacheName] = creator
}
