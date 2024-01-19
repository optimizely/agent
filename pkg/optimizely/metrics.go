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
	"github.com/optimizely/agent/pkg/metrics"
	go_sdk_metrics "github.com/optimizely/go-sdk/v2/pkg/metrics"
)

// MetricsRegistry initializes metrics registry
type MetricsRegistry struct {
	registry *metrics.Registry
}

// NewRegistry initializes metrics registry
func NewRegistry(registry *metrics.Registry) *MetricsRegistry {
	return &MetricsRegistry{registry: registry}
}

// GetCounter gets sdk Counter
func (m *MetricsRegistry) GetCounter(key string) go_sdk_metrics.Counter {
	return m.registry.GetCounter(key)
}

// GetGauge gets sdk Gauge
func (m *MetricsRegistry) GetGauge(key string) go_sdk_metrics.Gauge {
	return m.registry.GetGauge(key)
}
