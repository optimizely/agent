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

// Package optimizely wraps the Optimizely SDK
package optimizely

import (
	"github.com/optimizely/go-sdk/pkg/entities"
)

// OptlyContext encapsulates the user context.
// TODO Add support for User Profile Service
type OptlyContext struct {
	UserContext *entities.UserContext
}

// NewContext creates the base Context for a user
func NewContext(id string, attributes map[string]interface{}) *OptlyContext {
	userContext := entities.UserContext{
		ID:         id,
		Attributes: attributes,
	}

	ctx := &OptlyContext{
		UserContext: &userContext,
	}

	return ctx
}

// GetUserID returns the user ID from within the UserContext
func (c *OptlyContext) GetUserID() string {
	return c.UserContext.ID
}
