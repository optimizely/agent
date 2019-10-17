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
	server *httptest.Server
}

func (suite *RequestHeader) SetupTest() {
	getTestHandler := func() http.HandlerFunc {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	}
	suite.server = httptest.NewServer(SetRequestID(getTestHandler()))
}

func (suite *RequestHeader) TearDownTest() {
	suite.server.Close()
}

func (suite *RequestHeader) TestSetRequestHeaderWithEmtpyHeader() {

	res, err := http.Get(suite.server.URL + "/some_string")
	suite.NoError(err)
	header := res.Header["X-Request-Id"]
	suite.Equal(len(header), 1)
	suite.Equal(len(header[0]), 36)
}

func (suite *RequestHeader) TestSetRequestHeaderWithRequestHeader() {

	req, _ := http.NewRequest("POST", suite.server.URL+"/some_post_request", nil)

	req.Header.Set("X-Request-Id", "12345")

	client := &http.Client{}
	res, err := client.Do(req)

	suite.NoError(err)
	header := res.Header["X-Request-Id"]
	suite.Equal(len(header), 1)
	suite.Equal(header[0], "12345")
}

func TestRequestHeaderSuite(t *testing.T) {
	suite.Run(t, new(RequestHeader))

}
