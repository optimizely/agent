/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/optimizely/agent/config"

	"github.com/go-chi/render"
	"golang.org/x/sync/errgroup"
)

// BatchResponse has the structure for the final response
type BatchResponse struct {
	StartedAt     time.Time           `json:"startedAt"`
	EndedAt       time.Time           `json:"endedAt"`
	ErrorCount    int                 `json:"errorCount"`
	ResponseItems []ResponseCollector `json:"response"`

	lock sync.Mutex
}

// NewBatchResponse constructs a BatchResponse with default values
func NewBatchResponse() *BatchResponse {
	return &BatchResponse{
		StartedAt:     time.Now(),
		ResponseItems: make([]ResponseCollector, 0),
	}
}

func (br *BatchResponse) append(col ResponseCollector) {
	br.lock.Lock()
	defer br.lock.Unlock()
	if col.Status != http.StatusOK {
		br.ErrorCount++
	}
	br.ResponseItems = append(br.ResponseItems, col)
	br.EndedAt = time.Now()
}

// ResponseCollector collects responses for the writer
type ResponseCollector struct {
	Status      int         `json:"status"`
	RequestID   string      `json:"requestID"`
	OperationID string      `json:"operationID"`
	Method      string      `json:"method"`
	URL         string      `json:"url"`
	Body        interface{} `json:"body"`

	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt"`

	headerMap http.Header
}

// NewResponseCollector constructs a ResponseCollector with default values
func NewResponseCollector(op BatchOperation) ResponseCollector {
	return ResponseCollector{
		headerMap:   make(http.Header),
		Method:      op.Method,
		URL:         op.URL,
		OperationID: op.OperationID,
		StartedAt:   time.Now(),
		Status:      http.StatusOK,
	}
}

// WriteHeader sets the status code
func (rec *ResponseCollector) WriteHeader(code int) {
	rec.Status = code
}

// Write is just the collector for BatchResponse
func (rec *ResponseCollector) Write(b []byte) (int, error) {
	var data interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return 0, err
	}

	rec.Body = data

	if header, ok := rec.headerMap["X-Request-Id"]; ok {
		rec.RequestID = header[0]
	}

	return 0, nil
}

// Header returns header map
func (rec *ResponseCollector) Header() http.Header {
	return rec.headerMap
}

// BatchOperation defines a single request within a batch
type BatchOperation struct {
	Method      string                 `json:"method"`
	URL         string                 `json:"url"`
	OperationID string                 `json:"operationID"`
	Body        map[string]interface{} `json:"body"`
	Params      map[string]string      `json:"params"`
	Headers     map[string]string      `json:"headers"`
}

// BatchRequest is the original request that is used for batching
type BatchRequest struct {
	Operations []BatchOperation `json:"operations"`
}

func makeRequest(next http.Handler, op BatchOperation) (ResponseCollector, error) {

	col := NewResponseCollector(op)

	bytesBody, err := json.Marshal(op.Body)
	if err != nil {
		col.Status = http.StatusBadRequest
		return col, fmt.Errorf("cannot convert operation body to bytes for operation id: %s with error: %v", op.OperationID, err)
	}
	reader := bytes.NewReader(bytesBody)
	opReq, err := http.NewRequest(op.Method, op.URL, reader)
	if err != nil {
		col.Status = http.StatusBadRequest
		return col, fmt.Errorf("cannot make a new request for operation id: %s, with error: %v", op.OperationID, err)
	}

	for headerKey, headerValue := range op.Headers {
		opReq.Header.Add(headerKey, headerValue)
		col.headerMap[headerKey] = []string{headerValue}
	}

	for paramKey, paramValue := range op.Params {
		values := opReq.URL.Query()
		values.Add(paramKey, paramValue)
		opReq.URL.RawQuery = values.Encode()
	}

	next.ServeHTTP(&col, opReq)
	col.EndedAt = time.Now()
	return col, nil
}

// BatchRouter intercepts requests for the given url to return a StatusOK.
func BatchRouter(batchRequests config.BatchRequestsConfig) func(http.Handler) http.Handler {

	f := func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && strings.HasSuffix(strings.ToLower(r.URL.Path), "/batch") {

				decoder := json.NewDecoder(r.Body)
				var req BatchRequest
				err := decoder.Decode(&req)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					render.JSON(w, r, `{"error": "cannot decode the operation body"}`)
					return
				}

				if len(req.Operations) > batchRequests.OperationsLimit {
					GetLogger(r).Info().Msg(fmt.Sprintf("Too many operations, setting to the first %d operations", batchRequests.OperationsLimit))
					req.Operations = req.Operations[:batchRequests.OperationsLimit]
				}

				batchRes := NewBatchResponse()
				var eg errgroup.Group

				ch := make(chan struct{}, batchRequests.ParallelRequests)

				for _, op := range req.Operations {
					op := op
					eg.Go(func() error {
						ch <- struct{}{}
						defer func() { <-ch }()

						col, e := makeRequest(next, op)
						// Append response item
						batchRes.append(col)
						return e
					})
				}
				if err := eg.Wait(); err != nil {
					GetLogger(r).Error().Err(err).Msg("Problem with making a request")
				}
				render.JSON(w, r, batchRes)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
	return f
}
