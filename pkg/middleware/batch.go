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
	"github.com/go-chi/render"
	"net/http"
	"strings"
	"time"
)

// BatchResposeItem holds the structure for each item
type BatchResposeItem struct {
	Status    int         `json:"status"`
	RequestID string      `json:"requestID"`
	Method    string      `json:"method"`
	URL       string      `json:"url"`
	Body      interface{} `json:"body"`
}

// BatchResponse has the structure for the final response
type BatchResponse struct {
	StartedAt     time.Time          `json:"startedAt"`
	EndedAt       time.Time          `json:"endedAt"`
	ErrorCount    int                `json:"errorCount"`
	ResponseItems []BatchResposeItem `json:"response"`
}

// BatchWriter implements http.ResponseWriter
type BatchWriter struct {
	BResponse BatchResponse
}

func (br *BatchWriter) append(rec *ResponseCollector) {
	br.BResponse.EndedAt = time.Now()
	if rec.BResponse.Status != http.StatusOK {
		br.BResponse.ErrorCount++
	}

	br.BResponse.ResponseItems = append(br.BResponse.ResponseItems, rec.BResponse)
	br.BResponse.EndedAt = time.Now()
}

// ResponseCollector collects responses for the writer
type ResponseCollector struct {
	headerMap http.Header

	BResponse BatchResposeItem
}

// WriteHeader sets the status code
func (rec *ResponseCollector) WriteHeader(code int) {
	rec.BResponse.Status = code
}

// Write is just the collector for BatchResponse
func (rec *ResponseCollector) Write(b []byte) (int, error) {

	var data interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return 0, err
	}

	requestID := ""
	if header, ok := rec.headerMap["X-Request-Id"]; ok {
		requestID = header[0]

	}
	rec.BResponse.Body = data
	rec.BResponse.RequestID = requestID
	return 0, nil
}

// Header returns header map
func (rec *ResponseCollector) Header() http.Header {
	return rec.headerMap
}

// BatchRequest is the original request that is used for batching
type BatchRequest struct {
	Operations []struct {
		Method      string                 `json:"method"`
		URL         string                 `json:"url"`
		OperationID string                 `json:"operationID"`
		Body        map[string]interface{} `json:"body"`
		Params      map[string]string      `json:"params"`
		Headers     map[string]string      `json:"headers"`
	} `json:"operations"`
}

// BatchRouter intercepts requests for the given url to return a StatusOK.
func BatchRouter(next http.Handler) http.Handler {
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
			batchWriter := BatchWriter{BatchResponse{StartedAt: time.Now(), ResponseItems: []BatchResposeItem{}}}

			for _, op := range req.Operations {

				bytesBody, e := json.Marshal(op.Body)
				if err != nil {
					GetLogger(r).Error().Err(e).Msg("cannot convert operation body to bytes for operation id " + op.OperationID)
					continue
				}
				reader := bytes.NewReader(bytesBody)
				opReq, e := http.NewRequest(op.Method, op.URL, reader)
				if e != nil {
					GetLogger(r).Error().Err(e).Msg("cannot make a new request for operation id: " + op.OperationID)
					continue
				}

				col := ResponseCollector{make(http.Header), BatchResposeItem{Status: 200, Method: op.Method, URL: op.URL}}

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
				// Append response item
				batchWriter.append(&col)
			}

			render.JSON(w, r, batchWriter.BResponse)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
