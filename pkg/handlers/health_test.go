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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandlerBothServicesStarted(t *testing.T) {
	mux := chi.NewMux()
	mux.Get("/health", Health)

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, "Status code differs")
	expected := string(`{"status":"ok"}`)
	assert.JSONEq(t, expected, rec.Body.String(), "Response body differs")
}