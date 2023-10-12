/****************************************************************************
 * Copyright 2023 Optimizely, Inc. and contributors                        *
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
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func AddTracing(tracerName, spanName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx, span := otel.Tracer(tracerName).Start(r.Context(), spanName)
			defer span.End()

			span.SetAttributes(
				semconv.HTTPMethodKey.String(r.Method),
				semconv.HTTPRouteKey.String(r.URL.Path),
				semconv.HTTPURLKey.String(r.URL.String()),
				semconv.HTTPHostKey.String(r.Host),
				semconv.HTTPSchemeKey.String(r.URL.Scheme),
				attribute.String(OptlySDKHeader, r.Header.Get(OptlySDKHeader)),
			)

			rec := &statusRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rec, r.WithContext(ctx))

			span.SetAttributes(
				semconv.HTTPStatusCodeKey.Int(rec.statusCode),
			)
		}
		return http.HandlerFunc(fn)
	}
}
