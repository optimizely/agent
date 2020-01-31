/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                        *
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
	"net/http"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"

	"github.com/optimizely/go-sdk/pkg/entities"

	"github.com/go-chi/render"
	"github.com/pkg/errors"
)

// ActivateBody
type ActivateBody struct {
	UserID         string                 `json:"userId"`
	UserAttributes map[string]interface{} `json:"userAttributes"`
}

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

	if feature, err := middleware.GetFeature(r); err == nil {
		logger.Debug().Str("featureKey", feature.Key).Msg("deciding on feature")
		decision, err := optlyClient.ActivateFeature(feature, uc)
		if err != nil {
			RenderError(err, http.StatusInternalServerError, w, r)
			return
		}

		render.JSON(w, r, decision)
		return
	}

	if experiment, err := middleware.GetExperiment(r); err == nil {
		logger.Debug().Str("experimentKey", experiment.Key).Msg("deciding on experiment")
		decision, err := optlyClient.ActivateExperiment(experiment, uc)
		if err != nil {
			RenderError(err, http.StatusInternalServerError, w, r)
			return
		}

		render.JSON(w, r, decision)
		return
	}

	RenderError(errors.New("entity does not exist"), http.StatusInternalServerError, w, r)
}

// ActivateAll currently only support listing all feature decisions.
func ActivateAll(w http.ResponseWriter, r *http.Request) {
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

	var decisions []*optimizely.Decision
	oConf := optlyClient.GetOptimizelyConfig()

	logger.Debug().Msg("iterate over all experiment decisions")
	for _, e := range oConf.ExperimentsMap {
		d, err := optlyClient.ActivateExperiment(&e, uc)
		if err != nil {
			RenderError(err, http.StatusInternalServerError, w, r)
			return
		}

		if !d.Enabled {
			continue
		}

		decisions = append(decisions, d)
	}

	logger.Debug().Msg("iterate over all feature decisions")
	for _, f := range oConf.FeaturesMap {
		d, err := optlyClient.ActivateFeature(&f, uc)
		if err != nil {
			RenderError(err, http.StatusInternalServerError, w, r)
			return
		}

		if !d.Enabled {
			continue
		}

		decisions = append(decisions, d)
	}

	render.JSON(w, r, decisions)
}

func getUserContext(r *http.Request) (entities.UserContext, error) {
	var body ActivateBody
	err := ParseRequestBody(r, &body)
	if err != nil {
		return entities.UserContext{}, err
	}

	return entities.UserContext{ID: body.UserID, Attributes: body.UserAttributes}, nil
}
