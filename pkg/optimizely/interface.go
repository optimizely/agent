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
	optimizelyconfig "github.com/optimizely/go-sdk/pkg/config"
)

// Cache defines a basic interface for retrieving an instance of the OptlyClient keyed off of the SDK Key
type Cache interface {
	GetClient(sdkKey string) (*OptlyClient, error)
	UpdateConfigs(sdkKey string)
}

// SyncedConfigManager has the basic ConfigManager methods plus the SyncConfig method to trigger immediate updates
type SyncedConfigManager interface {
	optimizelyconfig.ProjectConfigManager
	SyncConfig()
}
