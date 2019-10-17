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

// Package handlers //
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	// "testing/iotest"

	"github.com/stretchr/testify/assert"
)

type ErrorReader struct{}

func (r *ErrorReader) Read(p []byte) (n int, err error) {
	err = fmt.Errorf("error")
	return
}

type TestEntity struct {
	Name  string `json:"name"`
	Value int16  `json:"value"`
}

func TestRenderError(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	RenderError(fmt.Errorf("new error"), http.StatusBadRequest, rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "{\"error\":\"new error\"}\n", rec.Body.String())
}

func TestParseRequestBody(t *testing.T) {
	expected := TestEntity{
		Name:  "test name",
		Value: 1,
	}
	jsonEntity, err := json.Marshal(expected)
	assert.NoError(t, err)
	req := httptest.NewRequest("GET", "/", bytes.NewBuffer(jsonEntity))

	var actual TestEntity
	err = ParseRequestBody(req, &actual)
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func TestParseRequestBodyError(t *testing.T) {
	req := httptest.NewRequest("GET", "/", bytes.NewBufferString("not json"))

	var actual TestEntity
	err := ParseRequestBody(req, &actual)
	assert.Error(t, err)
}

func TestParseRequestBodyNil(t *testing.T) {
	req := httptest.NewRequest("GET", "/", new(ErrorReader))

	var actual TestEntity
	err := ParseRequestBody(req, &actual)
	assert.Error(t, err)
}
