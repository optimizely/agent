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
	"github.com/optimizely/agent/pkg/optimizely"

	"github.com/optimizely/go-sdk/pkg/entities"

	"github.com/go-chi/render"
)

// ActivateBody defines the request body for an activation
type ActivateBody struct {
	UserID         string                 `json:"userId"`
	UserAttributes map[string]interface{} `json:"userAttributes"`
}

// Activate makes feature and experiment decisions for the selected query parameters.
func Activate(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	logger := middleware.GetLogger(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	uc, err := getUserContext(r)
	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	query := r.URL.Query()
	oConf := optlyClient.GetOptimizelyConfig()
	decisions := make([]*optimizely.Decision, 0, len(oConf.ExperimentsMap)+len(oConf.FeaturesMap))
	disableTracking := query.Get("disableTracking") == "true"

	experimentSet := make(map[string]struct{})
	featureSet := make(map[string]struct{})

	// Add filters for type
	for _, filterType := range query["type"] {
		switch filterType {
		case "experiment":
			for key := range oConf.ExperimentsMap {
				experimentSet[key] = struct{}{}
			}
		case "feature":
			for key := range oConf.FeaturesMap {
				featureSet[key] = struct{}{}
			}
		}
	}

	// Add explicit experiments
	for _, key := range query["experimentKey"] {
		experimentSet[key] = struct{}{}
	}

	// Add explicit features
	for _, key := range query["featureKey"] {
		featureSet[key] = struct{}{}
	}

	for key := range experimentSet {
		logger.Debug().Str("experimentKey", key).Msg("fetching experiment decision")
		e, ok := oConf.ExperimentsMap[key]
		if !ok {
			RenderError(fmt.Errorf("experimentKey not-found"), http.StatusNotFound, w, r)
			return
		}

		d, err := optlyClient.ActivateExperiment(&e, uc, disableTracking)
		if err != nil {
			RenderError(err, http.StatusInternalServerError, w, r)
			return
		}

		decisions = append(decisions, d)
	}

	for key := range featureSet {
		logger.Debug().Str("featureKey", key).Msg("fetching feature decision")
		f, ok := oConf.FeaturesMap[key]
		if !ok {
			RenderError(fmt.Errorf("featureKey not-found"), http.StatusNotFound, w, r)
			return
		}

		d, err := optlyClient.ActivateFeature(&f, uc, disableTracking)
		if err != nil {
			RenderError(err, http.StatusInternalServerError, w, r)
			return
		}

		decisions = append(decisions, d)
	}

	decisions = filterDecisions(r, decisions)
	render.JSON(w, r, decisions)
}

func filterDecisions(r *http.Request, decisions []*optimizely.Decision) []*optimizely.Decision {
	enabledFilter := r.URL.Query().Get("enabled")
	if enabledFilter == "" {
		return decisions
	}

	filtered := make([]*optimizely.Decision, 0, len(decisions))

	for _, decision := range decisions {
		if enabledFilter == "true" && !decision.Enabled {
			continue
		}

		if enabledFilter == "false" && decision.Enabled {
			continue
		}

		filtered = append(filtered, decision)
	}

	return filtered
}

func getUserContext(r *http.Request) (entities.UserContext, error) {
	var body ActivateBody
	err := ParseRequestBody(r, &body)
	if err != nil {
		return entities.UserContext{}, err
	}

	return entities.UserContext{ID: body.UserID, Attributes: body.UserAttributes}, nil
}
