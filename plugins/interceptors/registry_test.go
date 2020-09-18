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

// Package middleware //
package interceptors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testMiddleware struct {
	called bool
}

func (m *testMiddleware) Handler() func(http.Handler) http.Handler {
	m.called = true
	return nil
}

func TestAdd(t *testing.T) {
	Add("test", func() Interceptor { return &testMiddleware{} })
	mw := Interceptors["test"]()
	mw.Handler()
	if tmw, ok := mw.(*testMiddleware); ok {
		assert.True(t, tmw.called)
	} else {
		assert.Fail(t, "Cannot convert to type testMiddleware")
	}
}

func TestDoesNotExist(t *testing.T) {
	dne := Interceptors["DNE"]
	assert.Nil(t, dne)
}
