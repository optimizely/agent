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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func TestAddTracing(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	middleware := http.Handler(AddTracing("test-tracer", "test-span")(handler))

	// Serve the request through the middleware
	middleware.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	if body := rr.Body.String(); body != "OK" {
		t.Errorf("Expected response body %v, but got %v", "OK", body)
	}

	if typeHeader := rr.Header().Get("Content-Type"); typeHeader != "application/text" {
		t.Errorf("Expected Content-Type header %v, but got %v", "application/text", typeHeader)
	}
}

func TestNewIDs(t *testing.T) {
	gen := NewTraceIDGenerator()
	n := 1000

	for i := 0; i < n; i++ {
		traceID, spanID := gen.NewIDs(context.Background())
		assert.Truef(t, traceID.IsValid(), "trace id: %s", traceID.String())
		assert.Truef(t, spanID.IsValid(), "span id: %s", spanID.String())
	}
}

func TestNewSpanID(t *testing.T) {
	gen := NewTraceIDGenerator()
	testTraceID := [16]byte{123, 123}
	n := 1000

	for i := 0; i < n; i++ {
		spanID := gen.NewSpanID(context.Background(), testTraceID)
		assert.Truef(t, spanID.IsValid(), "span id: %s", spanID.String())
	}
}

func TestNewSpanIDWithInvalidTraceID(t *testing.T) {
	gen := NewTraceIDGenerator()
	spanID := gen.NewSpanID(context.Background(), trace.TraceID{})
	assert.Truef(t, spanID.IsValid(), "span id: %s", spanID.String())
}

func TestTraceIDWithGivenHeaderValue(t *testing.T) {
	gen := NewTraceIDGenerator()

	traceID := "9b8eac67e332c6f8baf1e013de6891bb"

	ctx := context.WithValue(context.Background(), OptlyTraceIDHeader, traceID)
	genTraceID, _ := gen.NewIDs(ctx)
	assert.Truef(t, genTraceID.IsValid(), "trace id: %s", genTraceID.String())
	assert.Equal(t, traceID, genTraceID.String())
}
