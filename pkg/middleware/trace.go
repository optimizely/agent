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
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"net/http"
	"sync"

	"github.com/optimizely/agent/config"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type traceIDGenerator struct {
	sync.Mutex
	randSource       *rand.Rand
	traceIDHeaderKey string
}

func NewTraceIDGenerator(traceIDHeaderKey string) *traceIDGenerator {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	return &traceIDGenerator{
		randSource:       rand.New(rand.NewSource(rngSeed)),
		traceIDHeaderKey: traceIDHeaderKey,
	}
}

func (gen *traceIDGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	gen.Lock()
	defer gen.Unlock()
	sid := trace.SpanID{}
	_, _ = gen.randSource.Read(sid[:])
	return sid
}

func (gen *traceIDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	gen.Lock()
	defer gen.Unlock()
	tid := trace.TraceID{}
	_, _ = gen.randSource.Read(tid[:])
	sid := trace.SpanID{}
	_, _ = gen.randSource.Read(sid[:])

	// read trace id from header if provided
	traceIDHeader := ctx.Value(gen.traceIDHeaderKey)
	if val, ok := traceIDHeader.(string); ok {
		if val != "" {
			headerTraceId, err := trace.TraceIDFromHex(val)
			if err == nil {
				tid = headerTraceId
			} else {
				log.Error().Err(err).Msg("failed to parse trace id from header, invalid trace id")
			}
		}
	}

	return tid, sid
}

func AddTracing(conf config.TracingConfig, tracerName, spanName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			pctx := context.WithValue(r.Context(), conf.OpenTelemetry.TraceIDHeaderKey, r.Header.Get(conf.OpenTelemetry.TraceIDHeaderKey))

			ctx, span := otel.Tracer(tracerName).Start(pctx, spanName)
			defer span.End()

			span.SetAttributes(
				semconv.HTTPMethodKey.String(r.Method),
				semconv.HTTPRouteKey.String(r.URL.Path),
				semconv.HTTPURLKey.String(r.URL.String()),
				semconv.HTTPHostKey.String(r.Host),
				semconv.HTTPSchemeKey.String(r.URL.Scheme),
				attribute.String(OptlySDKHeader, r.Header.Get(OptlySDKHeader)),
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
