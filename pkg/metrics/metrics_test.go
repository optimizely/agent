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

func TestGetHandler(t *testing.T) {
	assert.NotNil(t, GetHandler(""))
	assert.NotNil(t, GetHandler("expvar"))
	assert.NotNil(t, GetHandler("123131231"))
}

func TestCounterValid(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry()
	counter := metricsRegistry.GetCounter("metrics")
	counter.Add(12)
	counter.Add(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	assert.Equal(t, 35.0, expVarMap["counter.metrics"])

}

func TestCounterMultipleRetrievals(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry()
	counterKey := "next_counter_metrics"
	counter := metricsRegistry.GetCounter(counterKey)
	counter.Add(12)

	nextCounter := metricsRegistry.GetCounter(counterKey)
	nextCounter.Add(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	assert.Equal(t, 35.0, expVarMap["counter."+counterKey])
}

func TestCounterEmptyKey(t *testing.T) {

	metricsRegistry := NewRegistry()
	counter := metricsRegistry.GetCounter("")

	assert.Nil(t, counter)

}

func TestGaugeValid(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry()
	gauge := metricsRegistry.GetGauge("metrics")
	gauge.Set(12)
	gauge.Set(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	assert.Equal(t, 23.0, expVarMap["gauge.metrics"])

}

func TestGaugeMultipleRetrievals(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry()
	guageKey := "next_gauge_metrics"
	gauge := metricsRegistry.GetGauge(guageKey)
	gauge.Set(12)
	nextGauge := metricsRegistry.GetGauge(guageKey)
	nextGauge.Set(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	assert.Equal(t, 23.0, expVarMap["gauge."+guageKey])

}

func TestGaugeEmptyKey(t *testing.T) {

	metricsRegistry := NewRegistry()
	gauge := metricsRegistry.GetGauge("")

	assert.Nil(t, gauge)

}

func TestHistorgramValid(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry()
	histogram := metricsRegistry.GetHistogram("metrics")
	histogram.Observe(12)
	histogram.Observe(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	assert.Equal(t, 12.0, expVarMap["metrics.p50"])
	assert.Equal(t, 23.0, expVarMap["metrics.p99"])

}

func TestHistogramMultipleRetrievals(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry()
	histogramKey := "next_histogram_metrics"
	histogram := metricsRegistry.GetHistogram(histogramKey)
	histogram.Observe(12)
	nextGauge := metricsRegistry.GetHistogram(histogramKey)
	nextGauge.Observe(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	assert.Equal(t, 12.0, expVarMap["next_histogram_metrics.p50"])
	assert.Equal(t, 23.0, expVarMap["next_histogram_metrics.p99"])

}

func TestHistogramEmptyKey(t *testing.T) {

	metricsRegistry := NewRegistry()
	histogram := metricsRegistry.GetHistogram("")

	assert.Nil(t, histogram)

}
func TestTimerValid(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry()
	timer := metricsRegistry.NewTimer("metrics")
	timer.Update(12)
	timer.Update(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	assert.Equal(t, 2.0, expVarMap["timer.metrics.hits"])
	assert.Equal(t, 35.0, expVarMap["timer.metrics.responseTime"])
	assert.Equal(t, 12.0, expVarMap["timer.metrics.responseTimeHist.p50"])
	assert.Equal(t, 23.0, expVarMap["timer.metrics.responseTimeHist.p99"])
}

func TestTimerMultipleRetrievals(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry()
	timerKey := "next_timer_metrics"
	timer := metricsRegistry.NewTimer(timerKey)
	timer.Update(12)
	nextTimer := metricsRegistry.NewTimer(timerKey)
	nextTimer.Update(23)

	expvar.Handler().ServeHTTP(rec, req)

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)
	assert.Equal(t, 2.0, expVarMap["timer.next_timer_metrics.hits"])
	assert.Equal(t, 12.0, expVarMap["timer.next_timer_metrics.responseTimeHist.p50"])
	assert.Equal(t, 23.0, expVarMap["timer.next_timer_metrics.responseTimeHist.p99"])

}

func TestToSnakeCase(t *testing.T) {
	assert.Equal(t, "", toSnakeCase(""))
	assert.Equal(t, "abc", toSnakeCase("abc"))
	assert.Equal(t, "abc_123", toSnakeCase("abc_123"))
	assert.Equal(t, "abc_efg", toSnakeCase("abcEfg"))
	assert.Equal(t, "timer_activate_response_time", toSnakeCase("timer.activate.responseTime"))
	assert.Equal(t, "timer_activate_response_time_hist_p95", toSnakeCase("timer.activate.responseTimeHist.p95"))
	assert.Equal(t, "timer_create_api_access_token_response_time_hist_p50", toSnakeCase("timer.create-api-access-token.responseTimeHist.p50"))
	assert.Equal(t, "timer_get_config_response_time_hist_p50", toSnakeCase("timer.get-config.responseTimeHist.p50"))
	assert.Equal(t, "timer_track_event_response_time_hist_p50", toSnakeCase("timer.track-event.responseTimeHist.p50"))
	assert.Equal(t, "counter_dispatcher_success_flush", toSnakeCase("counter.dispatcher.successFlush"))
	assert.Equal(t, "timer_get_config_response_time", toSnakeCase("timer.get-config.responseTime"))
	assert.Equal(t, "timer_get_config_hits", toSnakeCase("timer.get-config.hits"))
}
