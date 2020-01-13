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

	"github.com/optimizely/agent/pkg/metrics"

	"github.com/stretchr/testify/assert"
)

type JSON map[string]interface{}

func TestCounterValid(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := metrics.NewRegistry()
	metricsSDKRegistry := NewRegistry(metricsRegistry)

	counter := metricsSDKRegistry.GetCounter("metrics")
	counter.Add(12)
	counter.Add(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	assert.Equal(t, 35.0, expVarMap["counter.metrics"])

}

func TestGaugeValid(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := metrics.NewRegistry()
	metricsSDKRegistry := NewRegistry(metricsRegistry)

	gauge := metricsSDKRegistry.GetGauge("metrics")
	gauge.Set(12)
	gauge.Set(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	assert.Equal(t, 23.0, expVarMap["gauge.metrics"])

}
