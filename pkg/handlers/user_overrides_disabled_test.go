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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/suite"
)

type UserOverridesDisabledTestSuite struct {
	suite.Suite
	mux *chi.Mux
}

// Setup Mux
func (suite *UserOverridesDisabledTestSuite) SetupTest() {

	mux := chi.NewMux()
	userAPI := new(DisabledUserOverrideHandler)

	mux.Put("/experiments/{experimentKey}", userAPI.SetForcedVariation)
	mux.Delete("/experiments/{experimentKey}", userAPI.RemoveForcedVariation)

	suite.mux = mux
}

func (suite *UserOverridesDisabledTestSuite) TestSetForcedVariation() {

	override := &UserOverrideBody{VariationKey: "variation_enabled"}
	body, err := json.Marshal(override)
	suite.NoError(err)

	req := httptest.NewRequest("PUT", "/experiments/feature_key", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusForbidden, rec.Code)

	response := string(rec.Body.Bytes())
	suite.Equal("Overrides not enabled\n", response)

}
func (suite *UserOverridesDisabledTestSuite) TestRemoveForcedVariation() {
	req := httptest.NewRequest("DELETE", "/experiments/feature_key", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusForbidden, rec.Code)

	response := string(rec.Body.Bytes())
	suite.Equal("Overrides not enabled\n", response)
}

func TestUserOverrideDisabledTestSuite(t *testing.T) {
	suite.Run(t, new(UserOverridesDisabledTestSuite))
}
