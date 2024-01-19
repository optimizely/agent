/****************************************************************************
 * Copyright 2020,2023-2024 Optimizely, Inc. and contributors               *
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
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/entities"
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
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	logger := middleware.GetLogger(r)

	uc, err := getUserContext(r)
	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	query := r.URL.Query()
	oConf := optlyClient.WithTraceContext(r.Context()).GetOptimizelyConfig()
	decisions := make([]*optimizely.Decision, 0, len(oConf.ExperimentsMap)+len(oConf.FeaturesMap))
	disableTracking := query.Get("disableTracking") == "true"

	kmap := make(keyMap)
	err = parseTypeParameter(query["type"], oConf, kmap)
	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	parseExperimentKeys(query["experimentKey"], oConf, kmap)

	parseFeatureKeys(query["featureKey"], oConf, kmap)

	for key, value := range kmap {
		var d *optimizely.Decision

		switch value {
		case "experiment":
			logger.Debug().Str("experimentKey", key).Msg("fetching experiment decision")
			d, err = optlyClient.ActivateExperiment(r.Context(), key, uc, disableTracking)
		case "feature":
			logger.Debug().Str("featureKey", key).Msg("fetching feature decision")
			d, err = optlyClient.ActivateFeature(r.Context(), key, uc, disableTracking)
		case "experimentKey-not-found":
			logger.Debug().Str("experimentKey", key).Msg("experimentKey not found")
			d = &optimizely.Decision{
				UserID:        uc.ID,
				ExperimentKey: key,
				Error:         "experimentKey not found",
			}
			err = nil
		case "featureKey-not-found":
			logger.Debug().Str("featureKey", key).Msg("featureKey not found")
			d = &optimizely.Decision{
				UserID:     uc.ID,
				FeatureKey: key,
				Error:      "featureKey not found",
			}
			err = nil
		default:
			err = fmt.Errorf(`type %q not supported`, value)
		}

		if err != nil {
			RenderError(err, http.StatusBadRequest, w, r)
			return
		}
		decisions = append(decisions, d)
	}

	decisions = filterDecisions(r, decisions)
	logger.Info().Msgf("Made activate decisions for user %s", uc.ID)
	render.JSON(w, r, decisions)
}

func parseExperimentKeys(keys []string, oConf *config.OptimizelyConfig, kmap keyMap) {
	for _, key := range keys {
		_, ok := oConf.ExperimentsMap[key]
		if !ok {
			kmap[key] = "experimentKey-not-found"
		} else {
			kmap[key] = "experiment"
		}
	}
}

func parseFeatureKeys(keys []string, oConf *config.OptimizelyConfig, kmap keyMap) {
	for _, key := range keys {
		_, ok := oConf.FeaturesMap[key]
		if !ok {
			kmap[key] = "featureKey-not-found"
		} else {
			kmap[key] = "feature"
		}
	}
}

func parseTypeParameter(types []string, oConf *config.OptimizelyConfig, kmap keyMap) error {
	for _, filterType := range types {
		switch filterType {
		case "experiment":
			for key := range oConf.ExperimentsMap {
				kmap[key] = "experiment"
			}
		case "feature":
			for key := range oConf.FeaturesMap {
				kmap[key] = "feature"
			}
		default:
			return fmt.Errorf(`type %q not supported`, filterType)
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

// ErrEmptyUserID is a constant error if userId is omitted from the request
var ErrEmptyUserID = errors.New(`missing "userId" in request payload`)

func getUserContext(r *http.Request) (entities.UserContext, error) {
	var body ActivateBody
	err := ParseRequestBody(r, &body)
	if err != nil {
		return entities.UserContext{}, err
	}

	if body.UserID == "" {
		return entities.UserContext{}, ErrEmptyUserID
	}

	return entities.UserContext{ID: body.UserID, Attributes: body.UserAttributes}, nil
}
