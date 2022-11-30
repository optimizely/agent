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
	"regexp"
	"strings"
	"sync"

	go_kit_metrics "github.com/go-kit/kit/metrics"
	go_kit_expvar "github.com/go-kit/kit/metrics/expvar"
	go_kit_prometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

// CounterPrefix stores the prefix for Counter
const (
	CounterPrefix = "counter"
	GaugePrefix   = "gauge"
	TimerPrefix   = "timer"
)

const (
	expVarPackage     = "expvar"
	prometheusPackage = "prometheus"
)

// Timer is the collection of some timers
type Timer struct {
	hits      go_kit_metrics.Counter
	totalTime go_kit_metrics.Counter
	histogram go_kit_metrics.Histogram
}

// NewTimer constructs Timer
func (m *Registry) NewTimer(key string) *Timer {
	if key == "" {
		log.Warn().Msg("metrics timer key is empty")
		return nil
	}
	combinedKey := TimerPrefix + "." + key

	m.timerLock.Lock()
	defer m.timerLock.Unlock()
	if val, ok := m.metricsTimerVars[combinedKey]; ok {
		return val
	}

	return m.createTimer(combinedKey)
}

// Update timer components
func (t *Timer) Update(delta float64) {
	t.hits.Add(1)
	t.totalTime.Add(delta)
	t.histogram.Observe(delta)
}

// Registry initializes expvar metrics registry
type Registry struct {
	metricsCounterVars   map[string]go_kit_metrics.Counter
	metricsGaugeVars     map[string]go_kit_metrics.Gauge
	metricsHistogramVars map[string]go_kit_metrics.Histogram
	metricsTimerVars     map[string]*Timer
	MetricsType          string

	gaugeLock     sync.RWMutex
	counterLock   sync.RWMutex
	histogramLock sync.RWMutex
	timerLock     sync.RWMutex
}

// NewRegistry initializes metrics registry
func NewRegistry() *Registry {

	return &Registry{
		metricsCounterVars:   map[string]go_kit_metrics.Counter{},
		metricsGaugeVars:     map[string]go_kit_metrics.Gauge{},
		metricsHistogramVars: map[string]go_kit_metrics.Histogram{},
		metricsTimerVars:     map[string]*Timer{},
	}
}

// GetCounter gets go-kit Counter
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

// GetGauge gets go-kit Gauge
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

// GetHistogram gets go-kit expvar Histogram
func (m *Registry) GetHistogram(key string) go_kit_metrics.Histogram {
	if key == "" {
		log.Warn().Msg("metrics histogram key is empty")
		return nil
	}

	m.histogramLock.Lock()
	defer m.histogramLock.Unlock()
	if val, ok := m.metricsHistogramVars[key]; ok {
		return val
	}
	return m.createHistogram(key)
}

func (m *Registry) createGauge(key string) (gaugeVar go_kit_metrics.Gauge) {
	// This is required since naming convention for every package differs
	name := m.getPackageSupportedName(key)
	switch m.MetricsType {
	case prometheusPackage:
		gaugeVar = go_kit_prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: name,
		}, []string{})
	default:
		gaugeVar = go_kit_expvar.NewGauge(name)
	}
	m.metricsGaugeVars[key] = gaugeVar
	return
}

func (m *Registry) createCounter(key string) (counterVar go_kit_metrics.Counter) {
	// This is required since naming convention for every package differs
	name := m.getPackageSupportedName(key)
	switch m.MetricsType {
	case prometheusPackage:
		counterVar = go_kit_prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name: name,
		}, []string{})
	default:
		counterVar = go_kit_expvar.NewCounter(name)
	}
	m.metricsCounterVars[key] = counterVar
	return
}

func (m *Registry) createHistogram(key string) (histogramVar go_kit_metrics.Histogram) {
	// This is required since naming convention for every package differs
	name := m.getPackageSupportedName(key)
	switch m.MetricsType {
	case prometheusPackage:
		histogramVar = go_kit_prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Name:    name,
			Buckets: []float64{50},
		}, []string{})
	default:
		histogramVar = go_kit_expvar.NewHistogram(name, 50)
	}
	m.metricsHistogramVars[key] = histogramVar
	return
}

func (m *Registry) createTimer(key string) *Timer {
	timerVar := &Timer{
		hits:      m.createCounter(key + ".hits"),
		totalTime: m.createCounter(key + ".responseTime"),
		histogram: m.createHistogram(key + ".responseTimeHist"),
	}
	m.metricsTimerVars[key] = timerVar
	return timerVar
}

// getPackageSupportedName converts name to package supported type
func (m *Registry) getPackageSupportedName(name string) string {
	switch m.MetricsType {
	case prometheusPackage:
		// https://prometheus.io/docs/practices/naming/
		v := strings.Replace(name, "-", "_", -1)
		strArray := strings.Split(v, ".")
		convertedArray := []string{}
		for _, v := range strArray {
			convertedArray = append(convertedArray, toSnakeCase(v))
		}
		return strings.Join(convertedArray, "_")
	default:
		return name
	}
}

func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
