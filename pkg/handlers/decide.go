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

	"github.com/go-chi/render"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/pkg/errors"

	"github.com/optimizely/agent/pkg/middleware"
	//"github.com/optimizely/agent/pkg/optimizely"
)

// DecisionBody
type DecisionContext struct {
	userId         string
	userAttributes map[string]interface{}
}

func Decide(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	logger := middleware.GetLogger(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	var body DecisionContext
	err = ParseRequestBody(r, &body)
	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	uc := entities.UserContext{
		ID:         body.userId,
		Attributes: body.userAttributes,
	}

	logger.Info().Msg("attempting to activate")

	if feature, err := middleware.GetFeature(r); err == nil {
		decision, err := optlyClient.GetFeatureDecision(feature, uc)
		if err != nil {
			RenderError(err, http.StatusInternalServerError, w, r)
			return
		}

		render.JSON(w, r, decision)
		return
	}

	if experiment, err := middleware.GetExperiment(r); err == nil {
		decision, err := optlyClient.GetExperimentDecision(experiment, uc)
		if err != nil {
			RenderError(err, http.StatusInternalServerError, w, r)
			return
		}

		render.JSON(w, r, decision)
		return
	}

	RenderError(errors.New("entity does not exist"), http.StatusInternalServerError, w, r)
}

func DecideAll(w http.ResponseWriter, r *http.Request) {
	//optlyClient, err := middleware.GetOptlyClient(r)
	//logger := middleware.GetLogger(r)
	//if err != nil {
	//	RenderError(err, http.StatusInternalServerError, w, r)
	//	return
	//}
}
