package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/render"
	"net/http"
	"strings"
	"time"
)

// BatchResponseItem holds the structure for each item
type BatchResposeItem struct {
	Status       int         `json:"status"`
	RequestID    string      `json:"requestID"`
	Method       string      `json:"method"`
	URL          string      `json:"url"`
	OrigResponse interface{} `json:"origResponse"`
}

// BatchResponse has the structure for the final response
type BatchResponse struct {
	StartedAt  time.Time          `json:"startedAt"`
	EndedAt    time.Time          `json:"endedAt"`
	ErrorCount int                `json:"errorCount"`
	Response   []BatchResposeItem `json:"response"`
}

// BatchWriter implements http.ResponseWriter
type BatchWriter struct {
	statusCode int
	headerMap  http.Header
	method     string
	url        string

	BResponse BatchResponse
}

// WriteHeader sets the status code
func (rec *BatchWriter) WriteHeader(code int) {
	rec.statusCode = code
}

// Write us just the collector for BatchResponse
func (rec *BatchWriter) Write(b []byte) (int, error) {

	var data interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return 0, err
	}
	if rec.statusCode != http.StatusOK {
		rec.BResponse.ErrorCount++
	}
	bri := BatchResposeItem{Status: rec.statusCode, OrigResponse: data, Method: rec.method, URL: rec.url, RequestID: rec.headerMap["X-Request-Id"][0]}
	rec.BResponse.Response = append(rec.BResponse.Response, bri)
	rec.BResponse.EndedAt = time.Now()
	return 0, nil
}

// Header returns header map
func (rec *BatchWriter) Header() http.Header {
	return rec.headerMap
}

type Request struct {
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
			var req Request
			err := decoder.Decode(&req)
			if err != nil {
				render.JSON(w, r, `{"error": "cannot decode the operation body"}`)
				return
			}
			batchWriter := BatchWriter{200, make(http.Header), "", "",
				BatchResponse{StartedAt: time.Now(), Response: []BatchResposeItem{}}}

			for _, op := range req.Operations {
				bytesBody, e := json.Marshal(op.Body)
				if err != nil {
					GetLogger(r).Error().Err(e).Msg("cannot convert operation body to bytes for operation id " + op.OperationID)
					continue
				}
				reader := bytes.NewReader(bytesBody)
				newReq, e := http.NewRequest(op.Method, op.URL, reader)
				if e != nil {
					GetLogger(r).Error().Err(e).Msg("cannot make a new request for operation id: " + op.OperationID)
					continue
				}

				for headerKey, headerValue := range op.Headers {
					newReq.Header.Add(headerKey, headerValue)
					batchWriter.headerMap[headerKey] = []string{headerValue}
				}

				for paramKey, paramValue := range op.Params {
					values := newReq.URL.Query()
					values.Add(paramKey, paramValue)
					newReq.URL.RawQuery = values.Encode()
				}

				batchWriter.method = op.Method
				batchWriter.url = op.URL

				next.ServeHTTP(&batchWriter, newReq)
			}

			render.JSON(w, r, batchWriter.BResponse)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
