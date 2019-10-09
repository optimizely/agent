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

// Package admin //
package admin

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderJSON(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		j := JSON{"key1": 1, "key2": "222"}
		renderJSON(w, r, j)
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/random_string")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	require.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, `{"key1":1,"key2":"222"}`+"\n", string(body))
	assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))
}

func TestHealthHandler(t *testing.T) {

	req, _ := http.NewRequest("GET", "/admin/health", nil)

	rr := httptest.NewRecorder()
	a := NewAdmin("1", "2", "3")
	http.HandlerFunc(a.Health).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	expected := string(`{"status":"ok"}`)

	assert.JSONEq(t, expected, rr.Body.String(), "Response body differs")
}

func TestAppInfoHandler(t *testing.T) {

	req, _ := http.NewRequest("GET", "/admin/info", nil)

	rr := httptest.NewRecorder()
	a := NewAdmin("1", "2", "3")
	http.HandlerFunc(a.AppInfo).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	expected := string(`{"app_name":"3", "app_version":"1", "author":"2", "host":""}`)

	assert.JSONEq(t, expected, rr.Body.String(), "Response body differs")
}

func TestAppInfoHeaderHandler(t *testing.T) {

	getTestHandler := func() http.HandlerFunc {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	}

	a := NewAdmin("1", "2", "3")
	ts := httptest.NewServer(a.AppInfoHeader(getTestHandler()))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/admin/info")
	assert.NoError(t, err)

	assert.Equal(t, res.Header["App-Version"], []string{"1"})
	assert.Equal(t, res.Header["Author"], []string{"2"})
	assert.Equal(t, res.Header["App-Name"], []string{"3"})
}
