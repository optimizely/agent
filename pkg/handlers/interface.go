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
)

// FeatureAPI defines the supported feature apis.
type FeatureAPI interface {
	GetFeature(w http.ResponseWriter, r *http.Request)
	ListFeatures(w http.ResponseWriter, r *http.Request)
}

// ExperimentAPI defines the supported experiment apis.
type ExperimentAPI interface {
	GetExperiment(w http.ResponseWriter, r *http.Request)
	ListExperiments(w http.ResponseWriter, r *http.Request)
}

// UserAPI defines the supported user scoped APIs.
type UserAPI interface {
	ListFeatures(w http.ResponseWriter, r *http.Request)
	GetFeature(w http.ResponseWriter, r *http.Request)
	TrackFeatures(w http.ResponseWriter, r *http.Request)
	TrackFeature(w http.ResponseWriter, r *http.Request)

	TrackEvent(w http.ResponseWriter, r *http.Request)

	ActivateExperiment(w http.ResponseWriter, r *http.Request)
	GetVariation(w http.ResponseWriter, r *http.Request)
}

// UserOverrideAPI defines supported override functionality
type UserOverrideAPI interface {
	SetForcedVariation(w http.ResponseWriter, r *http.Request)
	RemoveForcedVariation(w http.ResponseWriter, r *http.Request)
}
