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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RequestHeader struct {
	suite.Suite
}

var getTestHandler = func() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
}

func (suite *RequestHeader) TestSetRequestHeaderWithEmtpyHeader() {

	req := httptest.NewRequest("GET", "/some_string", nil)

	rec := httptest.NewRecorder()
	handler := http.Handler(SetRequestID(getTestHandler()))
	handler.ServeHTTP(rec, req)

	headerMap := rec.Header()
	suite.Equal(1, len(headerMap))

	header := headerMap[OptlyRequestHeader]
	suite.Equal(1, len(header))
	suite.Equal(36, len(header[0]))
}

func (suite *RequestHeader) TestSetRequestHeaderWithRequestHeader() {

	req := httptest.NewRequest("POST", "/some_post_request", nil)

	req.Header.Set(OptlyRequestHeader, "12345")

	rec := httptest.NewRecorder()
	handler := http.Handler(SetRequestID(getTestHandler()))
	handler.ServeHTTP(rec, req)

	headerMap := rec.Header()
	suite.Equal(1, len(headerMap))

	header := headerMap[OptlyRequestHeader]
	suite.Equal(1, len(header))
	suite.Equal("12345", header[0])
}

func TestRequestHeaderSuite(t *testing.T) {
	suite.Run(t, new(RequestHeader))

}
