/****************************************************************************
 * Copyright 2019,2023 Optimizely, Inc. and contributors                    *
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
	"expvar"
	"net/http"
	"regexp"
	"strings"
	"sync"

	go_kit_metrics "github.com/go-kit/kit/metrics"
	go_kit_expvar "github.com/go-kit/kit/metrics/expvar"
	go_kit_prometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

// CounterPrefix stores the prefix for Counter
const (
	CounterPrefix   = "counter"
	GaugePrefix     = "gauge"
	TimerPrefix     = "timer"
	ConstantsPrefix = "constants"
)

const (
	prometheusPackage = "prometheus"
)

// GetHandler returns request handler for provided metrics package type
func GetHandler(packageType string) http.Handler {
	switch packageType {
	case prometheusPackage:
		return promhttp.Handler()
	default:
		// expvar
		return expvar.Handler()
	}
}

// Timer is the collection of some timers
type Timer struct {
	hits      go_kit_metrics.Counter
	totalTime go_kit_metrics.Counter
	histogram go_kit_metrics.Histogram
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
	metricsType          string

	gaugeLock     sync.RWMutex
	counterLock   sync.RWMutex
	histogramLock sync.RWMutex
	timerLock     sync.RWMutex
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

// NewRegistry initializes metrics registry
func NewRegistry(metricsType string) *Registry {

	registry := &Registry{
		metricsCounterVars:   map[string]go_kit_metrics.Counter{},
		metricsGaugeVars:     map[string]go_kit_metrics.Gauge{},
		metricsHistogramVars: map[string]go_kit_metrics.Histogram{},
		metricsTimerVars:     map[string]*Timer{},
		metricsType:          metricsType,
	}
	registry.addDefaultConstantMetrics()
	return registry
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

// GetHistogram gets go-kit Histogram
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
	switch m.metricsType {
	case prometheusPackage:
		gaugeVar = go_kit_prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: name,
		}, []string{})
	default:
		// Default expvar
		gaugeVar = go_kit_expvar.NewGauge(name)
	}
	m.metricsGaugeVars[key] = gaugeVar
	return
}

func (m *Registry) createCounter(key string) (counterVar go_kit_metrics.Counter) {
	// This is required since naming convention for every package differs
	name := m.getPackageSupportedName(key)
	switch m.metricsType {
	case prometheusPackage:
		counterVar = go_kit_prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name: name,
		}, []string{})
	default:
		// Default expvar
		counterVar = go_kit_expvar.NewCounter(name)
	}
	m.metricsCounterVars[key] = counterVar
	return
}

func (m *Registry) createHistogram(key string) (histogramVar go_kit_metrics.Histogram) {
	// This is required since naming convention for every package differs
	name := m.getPackageSupportedName(key)
	switch m.metricsType {
	case prometheusPackage:
		histogramVar = go_kit_prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Name: name,
		}, []string{})
	default:
		// Default expvar
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

func (m *Registry) addDefaultConstantMetrics() {
	createCounterWithLabels := func(kv map[string]string) {
		switch m.metricsType {
		case prometheusPackage:
			labelNames := []string{}
			for k := range kv {
				labelNames = append(labelNames, k)
			}
			promauto.NewCounterVec(stdprometheus.CounterOpts{
				Name: ConstantsPrefix,
				Help: "Agent constants.",
			}, labelNames).With(kv)
		default:
			// Default expvar
			for k, v := range kv {
				expvar.NewString(k).Set(v)
			}
		}
	}
	// These can be retrieved from yaml configuration aswell
	// Needs to be discussed before writing tests
	sampleLabels := map[string]string{
		"host": "www.apple.com", "host1": "www.apple.com",
	}
	createCounterWithLabels(sampleLabels)
}

// getPackageSupportedName converts name to package supported type
func (m *Registry) getPackageSupportedName(name string) string {
	switch m.metricsType {
	case prometheusPackage:
		// https://prometheus.io/docs/practices/naming/
		return toSnakeCase(name)
	default:
		// Default expvar
		return name
	}
}

func toSnakeCase(name string) string {
	v := strings.Replace(name, "-", "_", -1)
	strArray := strings.Split(v, ".")
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	convertedArray := []string{}

	for _, v := range strArray {
		snake := matchFirstCap.ReplaceAllString(v, "${1}_${2}")
		snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
		convertedArray = append(convertedArray, strings.ToLower(snake))
	}
	return strings.Join(convertedArray, "_")
}
