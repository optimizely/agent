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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RequestBatch struct {
	suite.Suite
}

var getBatchHandler = func() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(rw, `{"error":"unable to fetch fresh datafile (consider rechecking SDK key), status code: 403 Forbidden"}`)
	})
}

func (suite *RequestHeader) TestBatchRouter() {

	operations := `{"operations": [
	{
		"method": "GET",
		"url": "/v1/config",
		"operation_id": "1",
		"body": {
		},
		"params": {"paramKey": "paramValue"},
		"headers": {
			"X-Optimizely-SDK-Key": "sdk_key",
            "X-Request-Id": "request1"
		}
	}]}`

	request := BatchRequest{}
	err := json.Unmarshal([]byte(operations), &request)
	suite.NoError(err)

	bytesBody, e := json.Marshal(request)
	suite.NoError(e)
	reader := bytes.NewReader([]byte(bytesBody))

	req := httptest.NewRequest("POST", "/batch", reader)
	rec := httptest.NewRecorder()
	handler := http.Handler(BatchRouter(getBatchHandler()))
	handler.ServeHTTP(rec, req)

	response := BatchResponse{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	suite.NoError(err)

	suite.Equal(1, response.ErrorCount)
	suite.False(response.StartedAt.IsZero())
	suite.False(response.EndedAt.IsZero())

	responseItem := response.ResponseItems[0]
	suite.Equal("/v1/config", responseItem.URL)
	suite.Equal("GET", responseItem.Method)
	suite.Equal("request1", responseItem.RequestID)
	suite.Equal(403, responseItem.Status)
	suite.Equal(map[string]interface{}{"error": "unable to fetch fresh datafile (consider rechecking SDK key), status code: 403 Forbidden"}, responseItem.Body)
}

func TestTestBatchRouterSuite(t *testing.T) {
	suite.Run(t, new(RequestBatch))

}
