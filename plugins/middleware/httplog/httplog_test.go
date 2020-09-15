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

// Package httplog //
package httplog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/agent/plugins/middleware"
)

func TestInit(t *testing.T) {
	name := "httplog"
	if mw, ok := middleware.MiddlewareRegistry[name]; !ok {
		assert.Failf(t, "Plugin not registered", "%s DNE in registry", name)
	} else {
		expected := &httpLog{}
		assert.Equal(t, expected, mw())
	}
}

func TestHandler(t *testing.T) {
	httpLog := &httpLog{}
	handler := httpLog.Handler()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler(http.NotFoundHandler()).ServeHTTP(w, r)
}
