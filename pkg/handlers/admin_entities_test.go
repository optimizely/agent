// ****************************************************************************
// * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
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

	"github.com/optimizely/agent/config"
)

var testConfig = config.AgentConfig{
	Version: "1",
	Author:  "2",
	Name:    "3",
}

func TestAdminHealthHandlerBothServicesStarted(t *testing.T) {

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	Health(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, "Status code differs")

	expected := string(`{"status":"ok"}`)
	assert.JSONEq(t, expected, rec.Body.String(), "Response body differs")
}

func TestAppInfoHandler(t *testing.T) {

	req := httptest.NewRequest("GET", "/info", nil)
	rec := httptest.NewRecorder()

	a := NewAdmin(testConfig)
	a.AppInfo(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Status code differs")

	actual := &Info{}
	err := json.Unmarshal(rec.Body.Bytes(), actual)
	assert.NoError(t, err)

	assert.Equal(t, "1", actual.Version)
	assert.Equal(t, "2", actual.Author)
	assert.Equal(t, "3", actual.AppName)
	assert.NotEmpty(t, actual.Uptime)
}

func TestAppConfigHandler(t *testing.T) {

	req := httptest.NewRequest("GET", "/config", nil)
	rec := httptest.NewRecorder()

	a := NewAdmin(testConfig)
	a.AppConfig(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Status code differs")

	actual := &config.AgentConfig{}
	err := json.Unmarshal(rec.Body.Bytes(), actual)
	assert.NoError(t, err)

	assert.Equal(t, &testConfig, actual)
}

func TestAppInfoHeaderHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/info", nil)
	rec := httptest.NewRecorder()

	a := NewAdmin(testConfig)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	a.AppInfoHeader(handler).ServeHTTP(rec, req)

	assert.Equal(t, []string{"1"}, rec.HeaderMap["App-Version"])
	assert.Equal(t, []string{"2"}, rec.HeaderMap["Author"])
	assert.Equal(t, []string{"3"}, rec.HeaderMap["App-Name"])
}

func TestMetrics(t *testing.T) {

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	a := NewAdmin(testConfig)
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
