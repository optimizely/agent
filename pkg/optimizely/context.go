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

	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// Context encapsulates the user and optimizely sdk
type Context struct {
	userContext *entities.UserContext
	optlyClient *client.OptimizelyClient
}

// NewContext creates a new UserContext and shared OptimizelyClient
func NewContext(id string, attributes map[string]interface{}) *Context {
	return NewContextWithOptimizely(id, attributes, GetOptimizely())
}

// NewContextWithOptimizely creates a new UserContext and a given OptimizelyClient
func NewContextWithOptimizely(id string, attributes map[string]interface{}, optlyClient *client.OptimizelyClient) *Context {
	userContext := entities.UserContext{
		ID:         id,
		Attributes: attributes,
	}
	context := &Context{
		userContext: &userContext,
		optlyClient: optlyClient,
	}

	return context
}

// GetAndTrackFeature calls the OptimizelyClient with the current UserContext and tracks an impression
func (context *Context) GetAndTrackFeature(featureKey string) (enabled bool, variableMap map[string]string, err error) {
	// TODO implement impression tracking. Not sure if this is sdk or sidedoor responsibility
	return context.GetFeature(featureKey)
}

<<<<<<< HEAD
// GetFeature calls the OptimizelyClient with the current UserContext this does NOT track experiment conversions
func (context *Context) GetFeature(featureKey string) (enabled bool, variableMap map[string]string, err error) {
	oc := context.optlyClient
	if oc == nil {
		return enabled, variableMap, errors.New("invalid optimizely instance")
	}
=======

// SetOptimizely sets the Optimizely client
func SetOptimizely() {
	sdkKey := os.Getenv("SDK_KEY")
	sublogger := log.With().Str("sdkKey", sdkKey).Logger()
	sublogger.Info().Msg("Fetching new OptimizelyClient")

	optimizelyFactory := &client.OptimizelyFactory{
		// TODO parameterize
		SDKKey: sdkKey,
	}

	var err error
	// TODO: Currently we are re-setting the entire client. Need to change this to only set config.
	optlyClient, err = optimizelyFactory.StaticClient()

	if err != nil {
		sublogger.Error().Err(err).Msg("Initializing OptimizelyClient")
		return
	}
}

// TODO Support multiple SDK keys
func getOptimizely() *client.OptimizelyClient {

	// TODO handle failure to prevent deadlocks.
	once.Do(func() { // <-- atomic, does not allow repeating
		SetOptimizely()
	})
>>>>>>> More changes

	return oc.GetAllFeatureVariables(featureKey, *context.userContext)
}
