// ****************************************************************************
// * Copyright 2019, Optimizely, Inc. and contributors                        *
// *                                                                          *
// * Licensed under the Apache License, Version 2.0 (the "License");          *
// * you may not use this file except in compliance with the License.         *
// * You may obtain a copy of the License at                                  *
// *                                                                          *
// *    http://www.apache.org/licenses/LICENSE-2.0                            *
// *                                                                          *
// * Unless required by applicable law or agreed to in writing, software      *
// * distributed under the License is distributed on an "AS IS" BASIS,        *
// * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
// * See the License for the specific language governing permissions and      *
// * limitations under the License.                                           *
// ***************************************************************************/

// Package handlers //
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockActiveService struct {
}

func (s MockActiveService) IsAlive() bool {
	return true
}

type MockInactiveService struct {
}

func (s MockInactiveService) IsAlive() bool {
	return false
}
func TestHealthHandler(t *testing.T) {

	req, _ := http.NewRequest("GET", "/admin/health", nil)

	rr := httptest.NewRecorder()

	/************** No services started ***********/

	a := NewAdmin("1", "2", "3", []AliveChecker{})
	http.HandlerFunc(a.Health).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	expected := string(`{"status":"error"}`)

	assert.JSONEq(t, expected, rr.Body.String(), "Response body differs")

	/************** both services started ***********/

	rr = httptest.NewRecorder()
	srvc1 := &MockActiveService{}
	srvc2 := &MockActiveService{}

	a = NewAdmin("1", "2", "3", []AliveChecker{srvc1, srvc2})
	http.HandlerFunc(a.Health).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	expected = string(`{"status":"ok"}`)

	assert.JSONEq(t, expected, rr.Body.String(), "Response body differs")

	/************** one service not started ***********/

	rr = httptest.NewRecorder()
	srvc3 := &MockInactiveService{}

	a = NewAdmin("1", "2", "3", []AliveChecker{srvc1, srvc2, srvc3})
	http.HandlerFunc(a.Health).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	expected = string(`{"status":"error"}`)

	assert.JSONEq(t, expected, rr.Body.String(), "Response body differs")
}

func TestAppInfoHandler(t *testing.T) {

	req, _ := http.NewRequest("GET", "/admin/info", nil)

	rr := httptest.NewRecorder()
	a := NewAdmin("1", "2", "3", nil)
	http.HandlerFunc(a.AppInfo).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	expected := string(`{"app_name":"3", "version":"1", "author":"2"}`)

	assert.JSONEq(t, expected, rr.Body.String(), "Response body differs")
}

func TestAppInfoHeaderHandler(t *testing.T) {

	getTestHandler := func() http.HandlerFunc {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	}

	a := NewAdmin("1", "2", "3", []AliveChecker{})
	ts := httptest.NewServer(a.AppInfoHeader(getTestHandler()))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/admin/info")
	assert.NoError(t, err)

	assert.Equal(t, res.Header["App-Version"], []string{"1"})
	assert.Equal(t, res.Header["Author"], []string{"2"})
	assert.Equal(t, res.Header["App-Name"], []string{"3"})
}
