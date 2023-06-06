/****************************************************************************
 * Copyright 2023, Optimizely, Inc. and contributors                        *
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
	"expvar"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusCounterValid(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry(prometheusPackage)
	counter := metricsRegistry.GetCounter("metrics")
	counter.Add(12)
	counter.Add(23)

	expvar.Handler().ServeHTTP(rec, req)

	promhttp.Handler().ServeHTTP(rec, req)
	resp, err := ioutil.ReadAll(rec.Body)
	assert.Nil(t, err)
	strResponse := string(resp)
	strings.Contains(strResponse, "counter_metrics 35")
}

func TestPrometheusCounterMultipleRetrievals(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry(prometheusPackage)
	counterKey := "next_counter_metrics"
	counter := metricsRegistry.GetCounter(counterKey)
	counter.Add(12)

	nextCounter := metricsRegistry.GetCounter(counterKey)
	nextCounter.Add(23)

	promhttp.Handler().ServeHTTP(rec, req)
	resp, err := ioutil.ReadAll(rec.Body)
	assert.Nil(t, err)
	strResponse := string(resp)
	strings.Contains(strResponse, "counter_"+counterKey+"35")
}

func TestPrometheusCounterEmptyKey(t *testing.T) {

	metricsRegistry := NewRegistry(prometheusPackage)
	counter := metricsRegistry.GetCounter("")
	assert.Nil(t, counter)
}

func TestPrometheusGaugeValid(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry(prometheusPackage)
	gauge := metricsRegistry.GetGauge("metrics")
	gauge.Set(12)
	gauge.Set(23)

	promhttp.Handler().ServeHTTP(rec, req)
	resp, err := ioutil.ReadAll(rec.Body)
	assert.Nil(t, err)
	strResponse := string(resp)
	strings.Contains(strResponse, "gauge_metrics")
}

func TestPrometheusGaugeMultipleRetrievals(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry(prometheusPackage)
	guageKey := "next_gauge_metrics"
	gauge := metricsRegistry.GetGauge(guageKey)
	gauge.Set(12)
	nextGauge := metricsRegistry.GetGauge(guageKey)
	nextGauge.Set(23)

	promhttp.Handler().ServeHTTP(rec, req)
	resp, err := ioutil.ReadAll(rec.Body)
	assert.Nil(t, err)
	strResponse := string(resp)
	strings.Contains(strResponse, "gauge_"+guageKey+" 23")
}

func TestPrometheusGaugeEmptyKey(t *testing.T) {

	metricsRegistry := NewRegistry(prometheusPackage)
	gauge := metricsRegistry.GetGauge("")
	assert.Nil(t, gauge)
}

func TestPrometheusHistogramValid(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry(prometheusPackage)
	histogram := metricsRegistry.GetHistogram("metrics")
	histogram.Observe(12)
	histogram.Observe(23)

	promhttp.Handler().ServeHTTP(rec, req)
	resp, err := ioutil.ReadAll(rec.Body)
	assert.Nil(t, err)
	strResponse := string(resp)
	strings.Contains(strResponse, "timer_metrics_response_time_hist_sum 35")
	strings.Contains(strResponse, "timer_metrics_response_time_hist_count 2")
}

func TestPrometheusHistogramMultipleRetrievals(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry(prometheusPackage)
	histogramKey := "next_histogram_metrics"
	histogram := metricsRegistry.GetHistogram(histogramKey)
	histogram.Observe(12)
	nextGauge := metricsRegistry.GetHistogram(histogramKey)
	nextGauge.Observe(23)

	promhttp.Handler().ServeHTTP(rec, req)
	resp, err := ioutil.ReadAll(rec.Body)
	assert.Nil(t, err)
	strResponse := string(resp)
	strings.Contains(strResponse, "timer_metrics_response_time_hist_sum 35")
	strings.Contains(strResponse, "timer_metrics_response_time_hist_count 2")
}

func TestPrometheusHistogramEmptyKey(t *testing.T) {

	metricsRegistry := NewRegistry(prometheusPackage)
	histogram := metricsRegistry.GetHistogram("")

	assert.Nil(t, histogram)
}

func TestPrometheusTimerValid(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry(prometheusPackage)
	timer := metricsRegistry.NewTimer("metrics")
	timer.Update(12)
	timer.Update(23)

	promhttp.Handler().ServeHTTP(rec, req)
	resp, err := ioutil.ReadAll(rec.Body)
	assert.Nil(t, err)
	strResponse := string(resp)
	strings.Contains(strResponse, "timer_metrics_response_time_hist_sum 35")
	strings.Contains(strResponse, "timer_metrics_response_time_hist_count 2")
	strings.Contains(strResponse, `timer_metrics_response_time_hist_bucket{le="+Inf"} 2`)
}

func TestPrometheusTimerMultipleRetrievals(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	metricsRegistry := NewRegistry(prometheusPackage)
	timerKey := "next_timer_metrics"
	timer := metricsRegistry.NewTimer(timerKey)
	timer.Update(12)
	nextTimer := metricsRegistry.NewTimer(timerKey)
	nextTimer.Update(23)

	promhttp.Handler().ServeHTTP(rec, req)
	resp, err := ioutil.ReadAll(rec.Body)
	assert.Nil(t, err)
	strResponse := string(resp)
	strings.Contains(strResponse, "timer_metrics_response_time_hist_sum 35")
	strings.Contains(strResponse, "timer_metrics_response_time_hist_count 2")
	strings.Contains(strResponse, `timer_metrics_response_time_hist_bucket{le="+Inf"} 2`)
}
