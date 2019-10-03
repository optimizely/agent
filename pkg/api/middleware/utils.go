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
	"fmt"
	"net/http"

	"github.com/optimizely/sidedoor/pkg/optimizely"
)

// GetOptlyClient is a utility to extract the OptlyClient from the http request context.
func GetOptlyClient(r *http.Request) (*optimizely.OptlyClient, error) {
	optlyClient, ok := r.Context().Value(OptlyClientKey).(*optimizely.OptlyClient)
	if !ok {
		return nil, fmt.Errorf("optlyClient not available")
	}

	return optlyClient, nil
}

// GetOptlyContext is a utility to extract the OptlyContext from the http request context.
func GetOptlyContext(r *http.Request) (*optimizely.OptlyContext, error) {
	optlyContext, ok := r.Context().Value(UserContextKey).(*optimizely.OptlyContext)
	if !ok {
		return nil, fmt.Errorf("optlyContext not available")
	}

	return optlyContext, nil
}
