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

	"github.com/stretchr/testify/assert"
)

func TestSetRequestHeaderWithEmtpyHeader(t *testing.T) {

	getTestHandler := func() http.HandlerFunc {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	}

	ts := httptest.NewServer(SetRequestID(getTestHandler()))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/some_string")
	assert.NoError(t, err)
	header := res.Header["X-Request-Id"]
	assert.Equal(t, len(header), 1)
	assert.Equal(t, len(header[0]), 36)
}

func TestSetRequestHeaderWithRequestHeader(t *testing.T) {

	getTestHandler := func() http.HandlerFunc {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	}

	ts := httptest.NewServer(SetRequestID(getTestHandler()))
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/some_post_request", nil)

	req.Header.Set("X-Request-Id", "12345")

	client := &http.Client{}
	res, err := client.Do(req)

	assert.NoError(t, err)
	header := res.Header["X-Request-Id"]
	assert.Equal(t, len(header), 1)
	assert.Equal(t, header[0], "12345")
}
