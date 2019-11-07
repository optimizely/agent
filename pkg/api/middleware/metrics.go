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

// Package middleware //
package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/expvar"
)

const metricPrefix = "timers."

type contextString string

const responseTime = contextString("responseTime")

// Metrics struct contains url hit counts, response time and its histogram
type Metrics struct {
	HitCounts             metrics.Counter
	ResponseTime          metrics.Counter
	ResponseTimeHistogram metrics.Histogram
}

// NewMetrics initialized metrics
func NewMetrics(key string) *Metrics {

	uniqueName := metricPrefix + key

	return &Metrics{
		HitCounts:             expvar.NewCounter(uniqueName + ".counts"),
		ResponseTime:          expvar.NewCounter(uniqueName + ".responseTime"),
		ResponseTimeHistogram: expvar.NewHistogram(uniqueName+".responseTimeHist", 50),
	}
}

// Metricize updates counts, total response time, and response time histogram
// for each URL hit, key being a combination of a method and route pattern
func Metricize(key string) func(http.Handler) http.Handler {
	singleMetric := NewMetrics(key)

	f := func(h http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {

			singleMetric.HitCounts.Add(1)
			ctx := r.Context()
			startTime, ok := ctx.Value(responseTime).(time.Time)
			if ok {
				defer func() {
					endTime := time.Now()
					timeDiff := endTime.Sub(startTime).Seconds()
					singleMetric.ResponseTime.Add(timeDiff)
					singleMetric.ResponseTimeHistogram.Observe(timeDiff)
				}()
			}

			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
	return f
}

// SetTime middleware sets the start time in request context
func SetTime(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {

		ctx := context.WithValue(r.Context(), responseTime, time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
