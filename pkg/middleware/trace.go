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

	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func AddTracing(tracerName, spanName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			prop := otel.GetTextMapPropagator()
			propCtx := prop.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			ctx, span := otel.Tracer(tracerName).Start(propCtx, spanName)
			defer span.End()

			span.SetAttributes(
				semconv.HTTPMethodKey.String(r.Method),
				semconv.HTTPRouteKey.String(r.URL.Path),
				semconv.HTTPURLKey.String(r.URL.String()),
				semconv.HTTPHostKey.String(r.Host),
				semconv.HTTPSchemeKey.String(r.URL.Scheme),
				attribute.String(OptlySDKHeader, r.Header.Get(OptlySDKHeader)),
			)

			respWriter := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(respWriter, r.WithContext(ctx))

			span.SetAttributes(
				semconv.HTTPStatusCodeKey.Int(respWriter.Status()),
			)
		}
		return http.HandlerFunc(fn)
	}
}
