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
package optlyd

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/entities"
)

var once sync.Once
var optlyClient *client.OptimizelyClient

func getOptimizely() *client.OptimizelyClient {

	// TODO handle failure to prevent deadlocks.
	once.Do(func() { // <-- atomic, does not allow repeating
		optimizelyFactory := &client.OptimizelyFactory{
			SDKKey: "UEeRJy1PfrfrEM6pGCGdz6",
		}

		var err error
		optlyClient, err = optimizelyFactory.StaticClient()

		if err != nil {
			fmt.Printf("Error instantiating client: %s", err)
			return
		}
	})

	return optlyClient
}

//--
// Error response payloads & renderers
//--

// ErrResponse renderer type for handling all sorts of errors.
//
// In the best case scenario, the excellent github.com/pkg/errors package
// helps reveal information on the error, setting it on Err, and in the Render()
// method, using it to set the application-specific error code in AppCode.
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (f *Feature) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

// ActivateExperiment - Return variation for an experiment and record impression
func ActivateExperiment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// ActivateFeature - Return the feature and record impression
func ActivateFeature(w http.ResponseWriter, r *http.Request) {
	featureKey := chi.URLParam(r, "featureKey")
	userID := r.URL.Query().Get("userId")

	if userID == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("missing userID")))
		return
	}

	app := getOptimizely()

	if app == nil {
		render.Render(w, r, ErrRender(errors.New("invalid optimizely instance")))
		return
	}

	user := entities.UserContext{
		ID:         userID,
		Attributes: map[string]interface{}{},
	}

	enabled, _ := app.IsFeatureEnabled(featureKey, user)

	var featureVariation FeatureVariation
	featureVariation.Key = featureKey

	feature := &Feature{
		Enabled: enabled,
		Key:     featureKey,
	}

	w.WriteHeader(http.StatusOK)
	render.Render(w, r, feature)
}

// GetExperiment - Return variation for a given user and experiment
func GetExperiment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// GetExperimentDecision - Return variation decision for an experiment
func GetExperimentDecision(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// GetExperimentVariation - Return variation for a given user and experiment
func GetExperimentVariation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// GetFeature - Return the feature
func GetFeature(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// GetFeatureDecision - Return the decided feature variant
func GetFeatureDecision(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// GetFeatureVariation - Return the feature variant
func GetFeatureVariation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// GetStats - Server metrics
func GetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// ListExperimentVariations - Return variation for a given user and experiment
func ListExperimentVariations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// ListExperiments - Return all experiments
func ListExperiments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// ListFeatureVariations - Return feature variations
func ListFeatureVariations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// ListFeatures - Return all features
func ListFeatures(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// TrackEvent - Track event endpoint
func TrackEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}
