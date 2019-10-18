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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/sidedoor/pkg/optimizely"
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

func TestGetOptlyContext(t *testing.T) {
	expected := new(optimizely.OptlyContext)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), OptlyContextKey, expected)

	actual, err := GetOptlyContext(req.WithContext(ctx))
	assert.NoError(t, err)
	assert.Same(t, expected, actual)
}

func TestGetOptlyContextWithError(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	_, err := GetOptlyContext(req)
	assert.Error(t, err)
	assert.Equal(t, "optlyContext not available", err.Error())
}

func TestGetLogger(t *testing.T) {
	out := &bytes.Buffer{}
	req := httptest.NewRequest("GET", "/", nil)

	req.Header.Set(OptlyRequestHeader, "12345")
	req.Header.Set(OptlySDKHeader, "some_key")
	logger := GetLogger(req)
	newLogger := logger.Output(out)
	newLogger.Info().Msg("some_message")

	assert.Contains(t, out.String(), `"requestId":"12345"`)
	assert.Contains(t, out.String(), `"sdkKey":"some_key"`)

}
