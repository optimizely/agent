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

func (s MockActiveService) IsHealthy() (bool, string) {
	return true, ""
}

type MockInactiveService struct {
}

func (s MockInactiveService) IsHealthy() (bool, string) {
	return false, "not healthy"
}
func TestHealthHandlerNoServicesStarted(t *testing.T) {

	req, _ := http.NewRequest("GET", "/health", nil)

	rr := httptest.NewRecorder()

	a := NewAdmin("1", "2", "3", []HealthChecker{})
	http.HandlerFunc(a.Health).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	expected := string(`{"status":"error", "reasons": ["no services"]}`)

	assert.JSONEq(t, expected, rr.Body.String(), "Response body differs")

}

func TestHealthHandlerBothServicesStarted(t *testing.T) {

	req, _ := http.NewRequest("GET", "/health", nil)

	rr := httptest.NewRecorder()

	srvc1 := &MockActiveService{}
	srvc2 := &MockActiveService{}

	a := NewAdmin("1", "2", "3", []HealthChecker{srvc1, srvc2})
	http.HandlerFunc(a.Health).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	expected := string(`{"status":"ok"}`)

	assert.JSONEq(t, expected, rr.Body.String(), "Response body differs")

}

func TestHealthHandlerOneServiceNotStarted(t *testing.T) {

	req, _ := http.NewRequest("GET", "/health", nil)

	rr := httptest.NewRecorder()

	srvc1 := &MockActiveService{}
	srvc2 := &MockActiveService{}
	srvc3 := &MockInactiveService{}

	a := NewAdmin("1", "2", "3", []HealthChecker{srvc1, srvc2, srvc3})
	http.HandlerFunc(a.Health).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	expected := string(`{"status":"error", "reasons": ["not healthy"]}`)

	assert.JSONEq(t, expected, rr.Body.String(), "Response body differs")
}

func TestHealthHandlerTwoServiceNotStarted(t *testing.T) {

	req, _ := http.NewRequest("GET", "/health", nil)

	rr := httptest.NewRecorder()

	srvc1 := &MockActiveService{}
	srvc2 := &MockInactiveService{}
	srvc3 := &MockInactiveService{}

	a := NewAdmin("1", "2", "3", []HealthChecker{srvc1, srvc2, srvc3})
	http.HandlerFunc(a.Health).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	expected := string(`{"status":"error", "reasons": ["not healthy", "not healthy"]}`)

	assert.JSONEq(t, expected, rr.Body.String(), "Response body differs")
}

func TestAppInfoHandler(t *testing.T) {

	req, _ := http.NewRequest("GET", "/info", nil)

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

	a := NewAdmin("1", "2", "3", []HealthChecker{})
	ts := httptest.NewServer(a.AppInfoHeader(getTestHandler()))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/info")
	assert.NoError(t, err)

	assert.Equal(t, res.Header["App-Version"], []string{"1"})
	assert.Equal(t, res.Header["Author"], []string{"2"})
	assert.Equal(t, res.Header["App-Name"], []string{"3"})
}
