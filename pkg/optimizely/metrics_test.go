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

// Package optimizely //
package optimizely

import (
	"encoding/json"
	"expvar"
	"net/http/httptest"
	"testing"

	"github.com/optimizely/go-sdk/pkg/event"

	"github.com/stretchr/testify/assert"
)

type JSON map[string]interface{}

var sufixList = []string{".queueSize", ".successFlush", ".failFlush", ".retryFlush"}
var metricPrefix = "dispatcher"

func TestMetrics(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	EPMetrics := &event.DefaultMetrics{QueueSize: 20, FailFlushCount: 1, SuccessFlushCount: 3, RetryFlushCount: 5}
	metrics := NewMetrics(metricPrefix)
	metrics.SetMetrics(EPMetrics)
	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)

	for _, item := range sufixList {
		expectedKey := metricPrefix + item
		value, ok := expVarMap[expectedKey]
		assert.True(t, ok)
		switch item {
		case ".queueSize":
			assert.Equal(t, 20.0, value)
		case ".successFlush":
			assert.Equal(t, 3.0, value)
		case ".failFlush":
			assert.Equal(t, 1.0, value)
		case ".retryFlush":
			assert.Equal(t, 5.0, value)

		}
	}
}
