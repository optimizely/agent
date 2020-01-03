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
	"strconv"
)

// Metrics initializes expvar metrics
type Metrics struct {
	counts *expvar.Map

	prefix string
}

// NewMetrics initializes metrics
func NewMetrics(prefix, collectionName string) *Metrics {

	return &Metrics{
		counts: expvar.NewMap(collectionName),
		prefix: prefix,
	}
}

// Inc increments value for given key by one
func (m *Metrics) Inc(key string) {
	mergedKey := m.prefix + "." + key
	m.counts.Add(mergedKey, 1)
}

// Set value for given key
func (m *Metrics) Set(key string, val int64) {
	mergedKey := m.prefix + "." + key
	v := expvar.Int{}
	v.Add(val)
	m.counts.Set(mergedKey, &v)
}

// Get returns value for given key
func (m *Metrics) Get(key string) int64 {
	mergedKey := m.prefix + "." + key
	v := m.counts.Get(mergedKey)
	vStr := v.String()
	i, err := strconv.ParseInt(vStr, 10, 64)
	if err != nil {
		panic(err)
	}
	return i

}
