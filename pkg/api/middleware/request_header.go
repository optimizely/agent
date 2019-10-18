/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
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

// Package middleware //
package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

// SetRequestID sets request ID obtained from the request header itself or from newly generated id
func SetRequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get(OptlyRequestHeader)
		if header == "" {
			header = uuid.New().String()
			r.Header.Add(OptlyRequestHeader, header)
		}
		w.Header().Set(OptlyRequestHeader, header)

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
