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
	"context"
	"encoding/json"
	"expvar"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type JSON map[string]interface{}

var getTestMetrics = func() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Need to sleep to register response duration
		time.Sleep(1 * time.Millisecond)
	})
}

type RequestMetrics struct {
	suite.Suite
	rw      http.ResponseWriter
	req     *http.Request
	handler http.Handler
}

func (rm *RequestMetrics) SetupRoute(key string) {

	rm.rw = httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	rm.req = r.WithContext(context.WithValue(r.Context(), responseTime, time.Now()))
	rm.handler = http.Handler(Metricize(key)(getTestMetrics()))

}

func (rm RequestMetrics) serveRoute() {
	rm.handler.ServeHTTP(rm.rw, rm.req)
}

func (rm RequestMetrics) serveSetTimehHandler() {
	http.Handler(SetTime(getTestMetrics())).ServeHTTP(rm.rw, rm.req)
}

func (rm RequestMetrics) serveExpvarRoute() {
	expvar.Handler().ServeHTTP(rm.rw, rm.req)
}

func (rm RequestMetrics) getMetricsMap() JSON {
	var expVarMap JSON
	err := json.Unmarshal(rm.rw.(*httptest.ResponseRecorder).Body.Bytes(), &expVarMap)
	rm.Suite.Nil(err)

	return expVarMap
}

func (rm RequestMetrics) getCode() int {
	return rm.rw.(*httptest.ResponseRecorder).Code
}

var sufixList = []string{".counts", ".responseTime", ".responseTimeHist.p50", ".responseTimeHist.p90", ".responseTimeHist.p95", ".responseTimeHist.p99"}

func (suite *RequestMetrics) TestUpdateMetricsHitOnce() {

	suite.SetupRoute("some_key")

	suite.serveRoute()

	suite.Equal(http.StatusOK, suite.getCode(), "Status code differs")
	suite.serveExpvarRoute()

	expVarMap := suite.getMetricsMap()
	for _, item := range sufixList {
		expectedKey := metricPrefix + "some_key" + item
		value, ok := expVarMap[expectedKey]
		suite.True(ok)

		suite.NotEqual(0.0, value)
	}
}

func (suite *RequestMetrics) TestUpdateMetricsHitMultiple() {

	const hitNumber = 10.0

	suite.SetupRoute("different_key")

	for i := 0; i < hitNumber; i++ {
		suite.serveRoute()
	}

	suite.Equal(http.StatusOK, suite.getCode(), "Status code differs")

	suite.serveExpvarRoute()

	expVarMap := suite.getMetricsMap()

	expectedKey := metricPrefix + "different_key.counts"
	value, ok := expVarMap[expectedKey]
	suite.True(ok)

	suite.NotEqual(0.0, value)
}

func TestRequestMetrics(t *testing.T) {
	suite.Run(t, new(RequestMetrics))
}
