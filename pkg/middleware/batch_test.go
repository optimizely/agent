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

	"github.com/optimizely/agent/config"

	"github.com/stretchr/testify/suite"
)

type RequestBatch struct {
	suite.Suite
}

var batchHandler http.Handler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusForbidden)
	fmt.Fprintln(rw, `{"error":"unable to fetch fresh datafile (consider rechecking SDK key), status code: 403 Forbidden"}`)
})

func (suite *RequestBatch) TestBatchRouter() {

	operations := `{"operations": [
	{
		"method": "GET",
		"url": "/v1/config",
		"operationID": "1",
		"body": {
		},
		"params": {"paramKey": "paramValue"},
		"headers": {
			"X-Optimizely-SDK-Key": "sdk_key",
            "X-Request-Id": "request1"
		}
	},
    {
		"method": "POST",
		"url": "/v1/activate",
		"operationID": "2",
		"body": {
		},
		"params": {"paramKey": "paramValue"},
		"headers": {
			"X-Optimizely-SDK-Key": "sdk_key",
            "X-Request-Id": "request2"
		}
	},
    {
		"method": "bad_request",
		"url": "/v1/#%",
		"operationID": "3",
		"body": {
		},
		"params": {"paramKey": "paramValue"},
		"headers": {
			"X-Optimizely-SDK-Key": "sdk_key",
            "X-Request-Id": "request3"
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
	handler := BatchRouter(config.BatchRequestsConfig{OperationsLimit: 3, ParallelRequests: 1})(batchHandler)

	handler.ServeHTTP(rec, req)

	response := BatchResponse{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	suite.NoError(err)

	suite.Equal(3, response.ErrorCount)
	suite.False(response.StartedAt.IsZero())
	suite.False(response.EndedAt.IsZero())

	for _, responseItem := range response.ResponseItems {
		switch responseItem.URL {
		case "/v1/config":
			suite.Equal("GET", responseItem.Method)
			suite.Equal("request1", responseItem.RequestID)
			suite.Equal("1", responseItem.OperationID)
			suite.Equal(403, responseItem.Status)
			suite.Equal(map[string]interface{}{"error": "unable to fetch fresh datafile (consider rechecking SDK key), status code: 403 Forbidden"}, responseItem.Body)

		case "/v1/activate":
			suite.Equal("POST", responseItem.Method)
			suite.Equal("request2", responseItem.RequestID)
			suite.Equal("2", responseItem.OperationID)
			suite.Equal(403, responseItem.Status)
			suite.Equal(map[string]interface{}{"error": "unable to fetch fresh datafile (consider rechecking SDK key), status code: 403 Forbidden"}, responseItem.Body)

		case "/v1/#%":
			suite.Equal("bad_request", responseItem.Method)
			suite.Equal(400, responseItem.Status)
			suite.Equal("3", responseItem.OperationID)
			suite.Equal(nil, responseItem.Body)

		default:
			suite.Fail("unsupported case")
		}
	}
}

func TestTestBatchRouterSuite(t *testing.T) {
	suite.Run(t, new(RequestBatch))
}
