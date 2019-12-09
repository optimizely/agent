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
	"expvar"

	"github.com/optimizely/go-sdk/pkg/event"
)

type Metrics struct {
	QueueSize    *expvar.Int
	SuccessFlush *expvar.Int
	FailFlush    *expvar.Int
	RetryFlush   *expvar.Int
}

// NewMetrics initializes metrics
func NewMetrics(prefixKey string) *Metrics {

	return &Metrics{
		QueueSize:    expvar.NewInt(prefixKey + ".QueueSize"),
		SuccessFlush: expvar.NewInt(prefixKey + ".SuccessFlush"),
		FailFlush:    expvar.NewInt(prefixKey + ".FailFlush"),
		RetryFlush:   expvar.NewInt(prefixKey + ".RetryFlush"),
	}
}

func (m *Metrics) SetMetrics(defaultMetrics *event.DefaultMetrics) {
	m.QueueSize.Set(int64(defaultMetrics.QueueSize))
	m.FailFlush.Set(defaultMetrics.FailFlushCount)
	m.RetryFlush.Set(defaultMetrics.RetryFlushCount)
	m.SuccessFlush.Set(defaultMetrics.SuccessFlushCount)
}
