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

var metricPrefix = "prefix"

func TestCounterValid(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry(metricPrefix)
	counter := metricsRegistry.GetCounter("counter_metrics")
	counter.Add(12)
	counter.Add(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	assert.Equal(t, 35.0, expVarMap["prefix.counter_metrics"])

}

func TestCounterEmptyPrefix(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry("")
	counter := metricsRegistry.GetCounter("empty_counter_metrics")
	counter.Add(12)
	counter.Add(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	_, ok := expVarMap["prefix.empty_counter_metrics"]
	assert.False(t, ok)
}

func TestCounterMultipleRetrievals(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry(metricPrefix)
	counterKey := "next_counter_metrics"
	counter := metricsRegistry.GetCounter(counterKey)
	counter.Add(12)

	nextCounter := metricsRegistry.GetCounter(counterKey)
	nextCounter.Add(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	assert.Equal(t, 35.0, expVarMap["prefix."+counterKey])
}

func TestGaugeValid(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry(metricPrefix)
	gauge := metricsRegistry.GetGauge("gauge_metrics")
	gauge.Set(12)
	gauge.Set(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	assert.Equal(t, 23.0, expVarMap["prefix.gauge_metrics"])

}

func TestGaugeEmptyPrefix(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry("")
	gauge := metricsRegistry.GetGauge("empty_gauge_metrics")
	gauge.Set(12)
	gauge.Set(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	_, ok := expVarMap["prefix.empty_gauge_metrics"]
	assert.False(t, ok)
}

func TestGaugeMultipleRetrievals(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry(metricPrefix)
	guageKey := "next_gauge_metrics"
	gauge := metricsRegistry.GetGauge(guageKey)
	gauge.Set(12)
	nextGauge := metricsRegistry.GetGauge(guageKey)
	nextGauge.Set(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	assert.Equal(t, 23.0, expVarMap["prefix."+guageKey])

}