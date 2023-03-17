/****************************************************************************
 * Copyright 2023, Optimizely, Inc. and contributors                        *
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

// Package utils //
package utils

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testDurationStruct struct {
	D Duration `json:"duration"`
}

func TestValidValues(t *testing.T) {

	testStruct := testDurationStruct{}
	testJSON := `{"duration": "5s"}`
	e := json.Unmarshal([]byte(testJSON), &testStruct)
	assert.NoError(t, e)
	assert.Equal(t, 5*time.Second, testStruct.D.Duration)

	testJSON = `{"duration": "5m"}`
	e = json.Unmarshal([]byte(testJSON), &testStruct)
	assert.NoError(t, e)
	assert.Equal(t, 5*time.Minute, testStruct.D.Duration)

	testJSON = `{"duration": 5}`
	e = json.Unmarshal([]byte(testJSON), &testStruct)
	assert.NoError(t, e)
	assert.Equal(t, 5*time.Nanosecond, testStruct.D.Duration)

	testJSON = `{}`
	testStruct = testDurationStruct{}
	e = json.Unmarshal([]byte(testJSON), &testStruct)
	assert.NoError(t, e)
	assert.Equal(t, time.Duration(0), testStruct.D.Duration)
}

func TestInvalidValues(t *testing.T) {

	// Time without unit
	testStruct := testDurationStruct{}
	testJSON := `{"duration": "5"}`
	e := json.Unmarshal([]byte(testJSON), &testStruct)
	assert.Error(t, e)
}
