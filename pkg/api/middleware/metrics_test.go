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

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/suite"
)

type JSON map[string]interface{}

var getTestMetrics = func() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
	})
}

type RequestMetrics struct {
	suite.Suite
	rw      http.ResponseWriter
	req     *http.Request
	handler http.Handler
}

func (rm *RequestMetrics) SetupTest() {

	metricsMap := NewMetricsCollection()
	rm.rw = httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	rctx := chi.NewRouteContext()
	rctx.RoutePatterns = []string{"/item/{set_item}"}

	rm.req = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	rm.handler = http.Handler(UpdateRouteMetrics(metricsMap)(getTestMetrics()))

}

func (rm RequestMetrics) serveRoute() {
	rm.handler.ServeHTTP(rm.rw, rm.req)
}

func (rm RequestMetrics) serveExpvarRoute() {
	expvar.Handler().ServeHTTP(rm.rw, rm.req)
}

func (rm RequestMetrics) serveSetTimehHandler() {
	http.Handler(SetTime(getTestMetrics())).ServeHTTP(rm.rw, rm.req)
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

	suite.serveSetTimehHandler()
	suite.serveRoute()

	suite.Equal(http.StatusOK, suite.getCode(), "Status code differs")

	suite.serveExpvarRoute()

	expVarMap := suite.getMetricsMap()
	for _, item := range sufixList {
		expectedKey := metricPrefix + "GET__item_{set_item}" + item
		value, ok := expVarMap[expectedKey]
		suite.True(ok)

		suite.NotEqual(0.0, value)
	}
}

func (suite *RequestMetrics) TestUpdateMetricsHitMultiple() {

	const hitNumber = 10.0

	for i := 0; i < hitNumber; i++ {
		suite.serveRoute()
	}

	suite.Equal(http.StatusOK, suite.getCode(), "Status code differs")

	suite.serveExpvarRoute()

	expVarMap := suite.getMetricsMap()

	expectedKey := metricPrefix + "GET__item_{set_item}.counts"
	value, ok := expVarMap[expectedKey]
	suite.True(ok)

	suite.NotEqual(0.0, value)
}

func (suite *RequestMetrics) TestGetMetrics() {
	metricsColl := NewMetricsCollection()
	suite.NotNil(metricsColl.MetricMap)
	suite.Empty(metricsColl.MetricMap)

	metricsColl.getMetrics("some_key")
	suite.NotNil(metricsColl.MetricMap)
	suite.NotEmpty(metricsColl.MetricMap)

	_, ok := metricsColl.MetricMap["some_key"]

	suite.True(ok)

}

func TestRequestMetrics(t *testing.T) {
	suite.Run(t, new(RequestMetrics))
}
