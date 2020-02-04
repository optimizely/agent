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

// Package handlers //
package handlers

import (
	"fmt"
	"net/http"

	"github.com/optimizely/agent/pkg/middleware"
)

// OverrideBody defines the request body for an override
type OverrideBody struct {
	UserID        string `json:"userId"`
	ExperimentKey string `json:"experimentKey"`
	VariationKey  string `json:"variationKey"`
}

// Override is used to set forced variations for a given experiment or feature test
func Override(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	logger := middleware.GetLogger(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	var body OverrideBody
	if ParseRequestBody(r, &body) != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	var experimentKey string
	// If the override key belonged to a feature attempt to map to an experiment
	if feature, ctxErr := middleware.GetFeature(r); ctxErr == nil {
		logger.Debug().Str("featureKey", feature.Key).Msg("deciding on feature")

		// determine the experiment key given the feature and variation key
		// this is likely wrong
		for ek, e := range feature.ExperimentsMap {
			if _, ok := e.VariationsMap[body.VariationKey]; ok {
				experimentKey = ek
				logger.Debug().Str("featureKey", feature.Key).Str("experimentKey", experimentKey).Msg("found override experiment key for feature")
				break
			}
		}

		RenderError(fmt.Errorf("cannot map feature key: %s to an experiment", feature.Key), http.StatusBadRequest, w, r)
		return

	}

	// If experiment key was used then
	if experiment, ctxErr := middleware.GetExperiment(r); ctxErr == nil {
		experimentKey = experiment.Key
	}

	logger.Debug().Str("experimentKey", experimentKey).Str("variationKey", body.VariationKey).Msg("setting override")
	wasSet, err := optlyClient.SetForcedVariation(experimentKey, body.UserID, body.VariationKey)
	switch {
	case err != nil:
		RenderError(err, http.StatusInternalServerError, w, r)
	case wasSet:
		w.WriteHeader(http.StatusCreated)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
