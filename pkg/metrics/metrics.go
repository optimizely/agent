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

// Package metrics //
package metrics

import (
	"sync"

	go_kit_metrics "github.com/go-kit/kit/metrics"

	go_kit_expvar "github.com/go-kit/kit/metrics/expvar"
	"github.com/rs/zerolog/log"
)

// CounterPrefix stores the prefix for Counter
const (
	CounterPrefix = "counter"
	GaugePrefix   = "gauge"
	TimerPrefix   = "timer"
)

// Registry initializes expvar metrics registry
type Registry struct {
	metricsCounterVars   map[string]go_kit_metrics.Counter
	metricsGaugeVars     map[string]go_kit_metrics.Gauge
	metricsHistogramVars map[string]go_kit_metrics.Histogram

	gaugeLock     sync.RWMutex
	counterLock   sync.RWMutex
	histogramLock sync.RWMutex
}

// NewRegistry initializes metrics registry
func NewRegistry() *Registry {

	return &Registry{
		metricsCounterVars:   map[string]go_kit_metrics.Counter{},
		metricsGaugeVars:     map[string]go_kit_metrics.Gauge{},
		metricsHistogramVars: map[string]go_kit_metrics.Histogram{},
	}
}

// GetCounter gets go-kit expvar Counter
func (m *Registry) GetCounter(key string) go_kit_metrics.Counter {

	if key == "" {
		log.Warn().Msg("metrics counter key is empty")
		return nil
	}

	combinedKey := CounterPrefix + "." + key

	m.counterLock.Lock()
	defer m.counterLock.Unlock()
	if val, ok := m.metricsCounterVars[combinedKey]; ok {
		return val
	}

	return m.createCounter(combinedKey)
}

// GetGauge gets go-kit expvar Gauge
func (m *Registry) GetGauge(key string) go_kit_metrics.Gauge {

	if key == "" {
		log.Warn().Msg("metrics gauge key is empty")
		return nil
	}

	combinedKey := GaugePrefix + "." + key

	m.gaugeLock.Lock()
	defer m.gaugeLock.Unlock()
	if val, ok := m.metricsGaugeVars[combinedKey]; ok {
		return val
	}
	return m.createGauge(combinedKey)
}

// GetTimer gets go-kit expvar Counter
func (m *Registry) GetTimer(key string) go_kit_metrics.Counter {
	if key == "" {
		log.Warn().Msg("metrics timer key is empty")
		return nil
	}

	combinedKey := TimerPrefix + "." + key

	m.counterLock.Lock()
	defer m.counterLock.Unlock()
	if val, ok := m.metricsCounterVars[combinedKey]; ok {
		return val
	}

	return m.createCounter(combinedKey)
}

// GetHistogram gets go-kit Histogram
func (m *Registry) GetHistogram(prefix, key string) go_kit_metrics.Histogram {

	if key == "" {
		log.Warn().Msg("metrics gauge key is empty")
		return nil
	}

	combinedKey := prefix + "." + key

	m.histogramLock.Lock()
	defer m.histogramLock.Unlock()
	if val, ok := m.metricsHistogramVars[combinedKey]; ok {
		return val
	}
	return m.createHistogram(combinedKey)
}

func (m *Registry) createGauge(key string) *go_kit_expvar.Gauge {
	gaugeVar := go_kit_expvar.NewGauge(key)
	m.metricsGaugeVars[key] = gaugeVar
	return gaugeVar

}

func (m *Registry) createCounter(key string) *go_kit_expvar.Counter {
	counterVar := go_kit_expvar.NewCounter(key)
	m.metricsCounterVars[key] = counterVar
	return counterVar

}

func (m *Registry) createHistogram(key string) *go_kit_expvar.Histogram {
	histogramVar := go_kit_expvar.NewHistogram(key, 50)
	m.metricsHistogramVars[key] = histogramVar
	return histogramVar

}
