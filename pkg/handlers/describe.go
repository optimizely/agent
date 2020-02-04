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

// Package handlers //
package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/optimizely/agent/pkg/middleware"
)

// DescribeAll returns the entire OptimizelyConfig object
func DescribeAll(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	conf := optlyClient.GetOptimizelyConfig()
	render.JSON(w, r, conf)
}

// Describe returns the associated Experiment or Feature definition.
func Describe(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r)

	if feature, err := middleware.GetFeature(r); err == nil {
		logger.Debug().Str("featureKey", feature.Key).Msg("deciding on feature")
		render.JSON(w, r, feature)
		return
	}

	if experiment, err := middleware.GetExperiment(r); err == nil {
		logger.Debug().Str("experimentKey", experiment.Key).Msg("deciding on experiment")
		render.JSON(w, r, experiment)
		return
	}

	RenderError(errors.New("entity does not exist"), http.StatusInternalServerError, w, r)
}
