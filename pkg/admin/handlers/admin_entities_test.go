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
	"encoding/json"
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
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	a := NewAdmin("1", "2", "3", []HealthChecker{})
	a.Health(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code, "Status code differs")

	expected := string(`{"status":"error", "reasons": ["no services"]}`)
	assert.JSONEq(t, expected, rec.Body.String(), "Response body differs")

}

func TestHealthHandlerBothServicesStarted(t *testing.T) {

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	srvc1 := &MockActiveService{}
	srvc2 := &MockActiveService{}

	a := NewAdmin("1", "2", "3", []HealthChecker{srvc1, srvc2})
	a.Health(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Status code differs")

	expected := string(`{"status":"ok"}`)
	assert.JSONEq(t, expected, rec.Body.String(), "Response body differs")
}

func TestHealthHandlerOneServiceNotStarted(t *testing.T) {

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	srvc1 := &MockActiveService{}
	srvc2 := &MockActiveService{}
	srvc3 := &MockInactiveService{}

	a := NewAdmin("1", "2", "3", []HealthChecker{srvc1, srvc2, srvc3})
	a.Health(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code, "Status code differs")

	expected := string(`{"status":"error", "reasons": ["not healthy"]}`)
	assert.JSONEq(t, expected, rec.Body.String(), "Response body differs")
}

func TestHealthHandlerTwoServiceNotStarted(t *testing.T) {

	req, _ := http.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	srvc1 := &MockActiveService{}
	srvc2 := &MockInactiveService{}
	srvc3 := &MockInactiveService{}

	a := NewAdmin("1", "2", "3", []HealthChecker{srvc1, srvc2, srvc3})
	a.Health(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code, "Status code differs")

	expected := string(`{"status":"error", "reasons": ["not healthy", "not healthy"]}`)
	assert.JSONEq(t, expected, rec.Body.String(), "Response body differs")
}

func TestAppInfoHandler(t *testing.T) {

	req := httptest.NewRequest("GET", "/info", nil)
	rec := httptest.NewRecorder()

	a := NewAdmin("1", "2", "3", nil)
	a.AppInfo(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Status code differs")

	expected := string(`{"app_name":"3", "version":"1", "author":"2"}`)
	assert.JSONEq(t, expected, rec.Body.String(), "Response body differs")
}

func TestAppInfoHeaderHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/info", nil)
	rec := httptest.NewRecorder()

	a := NewAdmin("1", "2", "3", []HealthChecker{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	a.AppInfoHeader(handler).ServeHTTP(rec, req)

	assert.Equal(t, "1", rec.Header().Get("App-Version"))
	assert.Equal(t, "2", rec.Header().Get("Author"))
	assert.Equal(t, "3", rec.Header().Get("App-Name"))
}

func TestMetrics(t *testing.T) {

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	a := NewAdmin("1", "2", "3", nil)
	a.Metrics(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Status code differs")

	var expVarMap JSON
	err := json.Unmarshal(rec.Body.Bytes(), &expVarMap)
	assert.Nil(t, err)

	memStatsMap, ok := expVarMap["memstats"]
	assert.True(t, ok)

	assert.Contains(t, memStatsMap, "Alloc")
	assert.Contains(t, memStatsMap, "BySize")
	assert.Contains(t, memStatsMap, "BuckHashSys")
}
