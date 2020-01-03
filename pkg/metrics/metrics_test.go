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
package metrics

import (
	"encoding/json"
	"expvar"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type JSON map[string]interface{}

var metricPrefix = "dispatcher"
var collectionName = "counter"

func TestMetrics(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metrics := NewMetrics(metricPrefix, collectionName)
	metrics.Set("metric1", 20)
	metrics.Set("metric2", 123)
	metrics.Inc("metric3")

	for i := 0; i < 3; i++ {
		metrics.Inc("metric4")
	}
	for i := 0; i < 5; i++ {
		metrics.Inc("metric5")
	}
	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	counterExpVarMap := expVarMap[collectionName].(map[string]interface{})

	assert.Len(t, counterExpVarMap, 5)
	assert.Equal(t, 20.0, counterExpVarMap["dispatcher.metric1"])
	assert.Equal(t, 123.0, counterExpVarMap["dispatcher.metric2"])
	assert.Equal(t, 3.0, counterExpVarMap["dispatcher.metric4"])
	assert.Equal(t, 1.0, counterExpVarMap["dispatcher.metric3"])
	assert.Equal(t, 5.0, counterExpVarMap["dispatcher.metric5"])

}
