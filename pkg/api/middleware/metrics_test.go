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

func (rm *RequestMetrics) setRoute(metricsKey string) {

	metricsMap := expvar.NewMap(metricsKey)
	rm.rw = httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	rctx := chi.NewRouteContext()
	rctx.RoutePatterns = []string{"/item/{set_item}"}

	rm.req = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	rm.handler = http.Handler(UpdateMetrics(metricsMap)(getTestMetrics()))

}

func (rm RequestMetrics) serveRoute() {
	rm.handler.ServeHTTP(rm.rw, rm.req)
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

func (suite *RequestMetrics) TestUpdateMetricsHitOnce() {

	var metricsKey = "counter"

	suite.setRoute(metricsKey)
	suite.serveRoute()

	suite.Equal(http.StatusOK, suite.getCode(), "Status code differs")

	suite.serveExpvarRoute()

	expVarMap := suite.getMetricsMap()

	counterMap, ok := expVarMap[metricsKey]
	suite.True(ok)

	suite.Contains(counterMap, "GET__item_{set_item}")

	m := counterMap.(map[string]interface{})

	suite.Equal(1.0, m["GET__item_{set_item}"])

}

func (suite *RequestMetrics) TestUpdateMetricsHitMultiple() {

	var metricsKey = "counter1"
	const hitNumber = 10.0

	suite.setRoute(metricsKey)

	for i := 0; i < hitNumber; i++ {
		suite.serveRoute()
	}

	suite.Equal(http.StatusOK, suite.getCode(), "Status code differs")

	suite.serveExpvarRoute()

	expVarMap := suite.getMetricsMap()

	counterMap, ok := expVarMap[metricsKey]
	suite.True(ok)

	suite.Contains(counterMap, "GET__item_{set_item}")

	m := counterMap.(map[string]interface{})

	suite.Equal(hitNumber, m["GET__item_{set_item}"])

}

func TestRequestMetrics(t *testing.T) {
	suite.Run(t, new(RequestMetrics))
}
