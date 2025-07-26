/****************************************************************************
 * Copyright 2021,2023-2024 Optimizely, Inc. and contributors               *
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
	"strings"

	"github.com/go-chi/render"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/go-sdk/v2/pkg/client"
	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/decision"
	"github.com/optimizely/go-sdk/v2/pkg/odp/segment"
)

const DefaultRolloutPrefix = "default-"

// DecideBody defines the request body for decide API
type DecideBody struct {
	UserID               string                            `json:"userId"`
	UserAttributes       map[string]interface{}            `json:"userAttributes"`
	DecideOptions        []string                          `json:"decideOptions"`
	ForcedDecisions      []ForcedDecision                  `json:"forcedDecisions,omitempty"`
	FetchSegments        bool                              `json:"fetchSegments"`
	FetchSegmentsOptions []segment.OptimizelySegmentOption `json:"fetchSegmentsOptions,omitempty"`
}

// ForcedDecision defines Forced Decision
type ForcedDecision struct {
	FlagKey      string `json:"flagKey"`
	RuleKey      string `json:"ruleKey,omitempty"`
	VariationKey string `json:"variationKey"`
}

// DecideOut defines the response
type DecideOut struct {
	client.OptimizelyDecision
	Variables               map[string]interface{} `json:"variables,omitempty"`
	IsEveryoneElseVariation bool                   `json:"isEveryoneElseVariation"`
	CmabUUID                *string                `json:"cmabUUID,omitempty"`
}

// Decide makes feature decisions for the selected query parameters
func Decide(w http.ResponseWriter, r *http.Request) {
	optlyClient, err := middleware.GetOptlyClient(r)
	logger := middleware.GetLogger(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}

	db, err := getUserContextWithOptions(r)
	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	decideOptions, err := decide.TranslateOptions(db.DecideOptions)
	if err != nil {
		RenderError(err, http.StatusBadRequest, w, r)
		return
	}

	optimizelyUserContext := optlyClient.WithTraceContext(r.Context()).CreateUserContext(db.UserID, db.UserAttributes)

	if db.FetchSegments {
		success := optimizelyUserContext.FetchQualifiedSegments(db.FetchSegmentsOptions)
		if !success {
			err := errors.New("failed to fetch qualified segments")
			RenderError(err, http.StatusInternalServerError, w, r)
			return
		}
	}

	// Setting up forced decisions
	for _, fd := range db.ForcedDecisions {
		context := decision.OptimizelyDecisionContext{FlagKey: fd.FlagKey, RuleKey: fd.RuleKey}
		forcedDecision := decision.OptimizelyForcedDecision{VariationKey: fd.VariationKey}
		optimizelyUserContext.SetForcedDecision(context, forcedDecision)
	}

	keys := []string{}
	if err := r.ParseForm(); err == nil {
		keys = r.Form["keys"]
	}

	featureMap := make(map[string]config.OptimizelyFeature)
	cfg := optlyClient.GetOptimizelyConfig()
	if cfg != nil {
		featureMap = cfg.FeaturesMap
	}

	switch len(keys) {
	case 0:
		// Decide All
		decides := optimizelyUserContext.DecideAll(decideOptions)
		decideOuts := []DecideOut{}
		for _, d := range decides {
			var cmabUUID *string
			if d.CmabUUID != nil && *d.CmabUUID != "" {
				cmabUUID = d.CmabUUID
			}
			decideOut := DecideOut{
				OptimizelyDecision:      d,
				Variables:               d.Variables.ToMap(),
				IsEveryoneElseVariation: isEveryoneElseVariation(featureMap[d.FlagKey].DeliveryRules, d.RuleKey),
				CmabUUID:                cmabUUID,
			}
			decideOuts = append(decideOuts, decideOut)
			logger.Debug().Msgf("Feature %q is enabled for user %s? %t", d.FlagKey, d.UserContext.UserID, d.Enabled)
		}
		render.JSON(w, r, decideOuts)
		return
	case 1:
		// Decide single key
		key := keys[0]
		logger.Debug().Str("featureKey", key).Msg("fetching feature decision")
		d := optimizelyUserContext.Decide(key, decideOptions)
		var cmabUUID *string
		if d.CmabUUID != nil && *d.CmabUUID != "" {
			cmabUUID = d.CmabUUID
		}
		decideOut := DecideOut{
			OptimizelyDecision:      d,
			Variables:               d.Variables.ToMap(),
			IsEveryoneElseVariation: isEveryoneElseVariation(featureMap[d.FlagKey].DeliveryRules, d.RuleKey),
			CmabUUID:                cmabUUID,
		}
		render.JSON(w, r, decideOut)
		return
	default:
		// Decide for multiple keys
		decides := optimizelyUserContext.DecideForKeys(keys, decideOptions)
		decideOuts := []DecideOut{}
		for _, d := range decides {
			var cmabUUID *string
			if d.CmabUUID != nil && *d.CmabUUID != "" {
				cmabUUID = d.CmabUUID
			}
			decideOut := DecideOut{
				OptimizelyDecision:      d,
				Variables:               d.Variables.ToMap(),
				IsEveryoneElseVariation: isEveryoneElseVariation(featureMap[d.FlagKey].DeliveryRules, d.RuleKey),
				CmabUUID:                cmabUUID,
			}
			decideOuts = append(decideOuts, decideOut)
			logger.Debug().Msgf("Feature %q is enabled for user %s? %t", d.FlagKey, d.UserContext.UserID, d.Enabled)
		}
		render.JSON(w, r, decideOuts)
		return
	}
}

func getUserContextWithOptions(r *http.Request) (DecideBody, error) {
	var body DecideBody
	err := ParseRequestBody(r, &body)
	if err != nil {
		return DecideBody{}, err
	}

	if body.UserID == "" {
		return DecideBody{}, ErrEmptyUserID
	}

	return body, nil
}

func isEveryoneElseVariation(rules []config.OptimizelyExperiment, ruleKey string) bool {
	for _, r := range rules {
		if r.Key == ruleKey {
			return r.Key == r.ID && strings.HasPrefix(r.Key, DefaultRolloutPrefix)
		}
	}
	return false
}
