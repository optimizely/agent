/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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

// Package interceptors //
package interceptors

import (
	"net/http"
)

// Interceptor interface for defining a middleware-style plugin
type Interceptor interface {
	Handler() func(http.Handler) http.Handler
}

// Creator type defines a function for creating an instance of a Interceptor
type Creator func() Interceptor

// Interceptors stores the mapping of  Creators
var Interceptors = map[string]Creator{}

// Add function registers a Middleware Creator
func Add(name string, creator Creator) {
	Interceptors[name] = creator
}
