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
	"expvar"
	"sync"

	"github.com/optimizely/go-sdk/pkg/metrics"

	go_kit_expvar "github.com/go-kit/kit/metrics/expvar"
	"github.com/rs/zerolog/log"
)

// Registry initializes expvar metrics registry
type Registry struct {
	prefix             string
	metricsCounterVars map[string]*go_kit_expvar.Counter
	metricsGaugeVars   map[string]*go_kit_expvar.Gauge

	gaugeLock   sync.RWMutex
	counterLock sync.RWMutex
}

// NewRegistry initializes metrics registry
func NewRegistry(prefix string) *Registry {

	return &Registry{
		prefix:             prefix,
		metricsCounterVars: map[string]*go_kit_expvar.Counter{},
		metricsGaugeVars:   map[string]*go_kit_expvar.Gauge{},
	}
}

// GetCounter gets go-kit expvar Counter
func (m *Registry) GetCounter(key string) metrics.Counter {

	if key == "" {
		log.Warn().Msg("metrics counter key is empty")
		return &metrics.BasicCounter{}
	}
	combinedKey := m.prefix + "." + key
	if expvar.Get(combinedKey) == nil {
		return m.createCounter(combinedKey)
	}

	m.counterLock.RLock()
	defer m.counterLock.RUnlock()
	if val, ok := m.metricsCounterVars[combinedKey]; ok {
		return val
	}
	log.Warn().Msg("unable to get counter metrics for key " + combinedKey)
	return &metrics.BasicCounter{}
}

// GetGauge gets go-kit expvar Gauge
func (m *Registry) GetGauge(key string) metrics.Gauge {
	if key == "" {
		log.Info().Msg("metrics gauge key is empty")
		return &metrics.BasicGauge{}
	}

	combinedKey := m.prefix + "." + key
	if expvar.Get(combinedKey) == nil {
		return m.createGauge(combinedKey)
	}
	m.gaugeLock.RLock()
	defer m.gaugeLock.RUnlock()
	if val, ok := m.metricsGaugeVars[combinedKey]; ok {
		return val
	}
	log.Warn().Msg("unable to get gauge metrics for key " + combinedKey)
	return &metrics.BasicGauge{}
}

func (m *Registry) createGauge(key string) *go_kit_expvar.Gauge {
	m.gaugeLock.Lock()
	defer m.gaugeLock.Unlock()
	gaugeVar := go_kit_expvar.NewGauge(key)
	m.metricsGaugeVars[key] = gaugeVar
	return gaugeVar

}

func (m *Registry) createCounter(key string) *go_kit_expvar.Counter {
	m.counterLock.Lock()
	defer m.counterLock.Unlock()
	counterVar := go_kit_expvar.NewCounter(key)
	m.metricsCounterVars[key] = counterVar
	return counterVar

}
