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

	"github.com/optimizely/go-sdk/pkg/config"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"

	"github.com/optimizely/go-sdk/pkg/entities"

	"github.com/go-chi/render"
)

type keyMap map[string]string

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

	km := make(keyMap)
	err = parseTypeParameter(query["type"], oConf, km)
	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	err = parseExperimentKeys(query["experimentKey"], oConf, km)
	if err != nil {
		RenderError(err, http.StatusNotFound, w, r)
		return
	}

	err = parseFeatureKeys(query["featureKey"], oConf, km)
	if err != nil {
		RenderError(err, http.StatusNotFound, w, r)
		return
	}

	for key, value := range km {
		var d *optimizely.Decision

		switch value {
		case "experiment":
			logger.Debug().Str("experimentKey", key).Msg("fetching experiment decision")
			d, err = optlyClient.ActivateExperiment(key, uc, disableTracking)
		case "feature":
			logger.Debug().Str("featureKey", key).Msg("fetching feature decision")
			d, err = optlyClient.ActivateFeature(key, uc, disableTracking)
		default:
			err = fmt.Errorf(`type "%s" not supported`, value)
		}

		if err != nil {
			RenderError(err, http.StatusBadRequest, w, r)
			return
		}

		decisions = append(decisions, d)
	}

	decisions = filterDecisions(r, decisions)
	render.JSON(w, r, decisions)
}

func parseExperimentKeys(keys []string, oConf *config.OptimizelyConfig, km keyMap) error {
	for _, key := range keys {
		_, ok := oConf.ExperimentsMap[key]
		if !ok {
			return fmt.Errorf("experimentKey not-found")
		}

		km[key] = "experiment"
	}

	return nil
}

func parseFeatureKeys(keys []string, oConf *config.OptimizelyConfig, km keyMap) error {
	for _, key := range keys {
		_, ok := oConf.FeaturesMap[key]
		if !ok {
			return fmt.Errorf("featureKey not-found")
		}

		km[key] = "feature"
	}

	return nil
}

func parseTypeParameter(types []string, oConf *config.OptimizelyConfig, km keyMap) error {
	for _, filterType := range types {
		switch filterType {
		case "experiment":
			for key := range oConf.ExperimentsMap {
				km[key] = "experiment"
			}
		case "feature":
			for key := range oConf.FeaturesMap {
				km[key] = "feature"
			}
		default:
			return fmt.Errorf(`type "%s" not supported`, filterType)
		}
	}

	return nil
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
