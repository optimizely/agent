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

// Package middleware //
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/agent/pkg/optimizely"

	"github.com/optimizely/go-sdk/pkg/config"
)

// GetOptlyClient is a utility to extract the OptlyClient from the http request context.
func TestGetOptlyClient(t *testing.T) {
	expected := new(optimizely.OptlyClient)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), OptlyClientKey, expected)

	actual, err := GetOptlyClient(req.WithContext(ctx))
	assert.NoError(t, err)
	assert.Same(t, expected, actual)
}

func TestGetOptlyClientWithError(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	_, err := GetOptlyClient(req)
	assert.Error(t, err)
	assert.Equal(t, "optlyClient not available", err.Error())
}

func TestGetLogger(t *testing.T) {
	out := &bytes.Buffer{}
	req := httptest.NewRequest("GET", "/", nil)

	req.Header.Set(OptlyRequestHeader, "12345")
	logger := GetLogger(req)
	newLogger := logger.Output(out)
	newLogger.Info().Msg("some_message")

	assert.Contains(t, out.String(), `"requestId":"12345"`)
}

func TestGetFeature(t *testing.T) {
	expected := &config.OptimizelyFeature{Key: "one"}

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), OptlyFeatureKey, expected)

	actual, err := GetFeature(req.WithContext(ctx))
	assert.Equal(t, expected, actual)
	assert.NoError(t, err)
}

func TestGetFeatureNotSet(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	actual, err := GetFeature(req)
	assert.Nil(t, actual)
	assert.Error(t, err)
}

func TestGetExperiment(t *testing.T) {
	expected := &config.OptimizelyExperiment{Key: "one"}

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), OptlyExperimentKey, expected)

	actual, err := GetExperiment(req.WithContext(ctx))
	assert.Equal(t, expected, actual)
	assert.NoError(t, err)
}

func TestGetExperimentNotSet(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	actual, err := GetExperiment(req)
	assert.Nil(t, actual)
	assert.Error(t, err)
}

func TestCoerceType(t *testing.T) {
	assert.Equal(t, int64(1), CoerceType("1"))
	assert.Equal(t, int64(10), CoerceType("10"))

	assert.Equal(t, true, CoerceType("true"))
	assert.Equal(t, false, CoerceType("false"))

	assert.Equal(t, 1.00, CoerceType("1.0"))
	assert.Equal(t, 1.01, CoerceType("1.01"))

	assert.Equal(t, "1.0a", CoerceType("1.0a"))
	assert.Equal(t, "True", CoerceType("True"))
	assert.Equal(t, "False", CoerceType("False"))
}

func TestCoerceTypeQuoted(t *testing.T) {
	assert.Equal(t, "1", CoerceType(`"1"`))
	assert.Equal(t, "10", CoerceType(`"10"`))

	assert.Equal(t, "true", CoerceType(`"true"`))
	assert.Equal(t, "false", CoerceType(`"false"`))

	assert.Equal(t, "1.0", CoerceType(`"1.0"`))
	assert.Equal(t, "1.01", CoerceType(`"1.01"`))

	assert.Equal(t, "1.0a", CoerceType(`"1.0a"`))
	assert.Equal(t, "True", CoerceType(`"True"`))
	assert.Equal(t, "False", CoerceType(`"False"`))
}

func assertError(t *testing.T, rec *httptest.ResponseRecorder, msg string, code int) {
	assert.Equal(t, code, rec.Code)

	var actual ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	assert.NoError(t, err)

	assert.Equal(t, ErrorResponse{Error: msg}, actual)
}
