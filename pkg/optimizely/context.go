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
	"errors"
	"os"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/entities"
)

var once sync.Once
var optlyClient *client.OptimizelyClient

// Context encapsulates the user and optimizely sdk
type Context struct {
	userContext *entities.UserContext
	optlyClient *client.OptimizelyClient
}

// NewContext creates a new UserContext and shared OptimizelyClient
func NewContext(id string, attributes map[string]interface{}) *Context {
	userContext := entities.UserContext{
		ID:         id,
		Attributes: attributes,
	}
	context := &Context{
		userContext: &userContext,
		optlyClient: getOptimizely(),
	}

	return context
}

// IsFeatureEnabled calls the OptimizelyClient with the current UserContext
func (context *Context) IsFeatureEnabled(featureKey string) (bool, error) {
	app := context.optlyClient

	if app == nil {
		return false, errors.New("invalid optimizely instance")
	}

	return app.IsFeatureEnabled(featureKey, *context.userContext)
}

// TODO Support multiple SDK keys
func getOptimizely() *client.OptimizelyClient {

	// TODO handle failure to prevent deadlocks.
	once.Do(func() { // <-- atomic, does not allow repeating
		sdkKey := os.Getenv("SDK_KEY")
		sublogger := log.With().Str("sdkKey", sdkKey).Logger()
		sublogger.Info().Msg("Fetching new OptimizelyClient")

		optimizelyFactory := &client.OptimizelyFactory{
			// TODO parameterize
			SDKKey: sdkKey,
		}

		var err error
		optlyClient, err = optimizelyFactory.StaticClient()

		if err != nil {
			sublogger.Error().Err(err).Msg("Initializing OptimizelyClient")
			return
		}
	})

	return optlyClient
}
