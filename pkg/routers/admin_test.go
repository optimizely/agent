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

// Package routers //
package routers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/optimizely/agent/config"
	"github.com/stretchr/testify/assert"
)

func TestAdminAllowedContentTypeMiddleware(t *testing.T) {

	conf := config.NewDefaultConfig()
	router := NewAdminRouter(*conf)

	// Testing unsupported content type
	body := "<request> <parameters> <email>test@123.com</email> </parameters> </request>"
	req := httptest.NewRequest("POST", "/oauth/token", bytes.NewBuffer([]byte(body)))
	req.Header.Add(contentTypeHeaderKey, "application/xml")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)

	// Testing supported content type
	body = `{"email":"test@123.com"}`
	req = httptest.NewRequest("POST", "/oauth/token", bytes.NewBuffer([]byte(body)))
	req.Header.Add(contentTypeHeaderKey, "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
