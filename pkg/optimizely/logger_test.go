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

// Package optimizely //
package optimizely

import (
	"bytes"
	"github.com/rs/zerolog"
	"testing"

	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	out := &bytes.Buffer{}
	logger := zerolog.New(out).Level(zerolog.DebugLevel)
	logConsumer := &LogConsumer{logger: &logger}

	logConsumer.Log(logging.LogLevelDebug, "debug")
	assert.Equal(t, "{\"level\":\"debug\",\"message\":\"debug\"}\n", out.String())
	out.Reset()

	logConsumer.Log(logging.LogLevelInfo, "info")
	assert.Equal(t, "{\"level\":\"info\",\"message\":\"info\"}\n", out.String())
	out.Reset()

	logConsumer.Log(logging.LogLevelWarning, "warn")
	assert.Equal(t, "{\"level\":\"warn\",\"message\":\"warn\"}\n", out.String())
	out.Reset()

	logConsumer.Log(logging.LogLevelError, "error")
	assert.Equal(t, "{\"level\":\"error\",\"message\":\"error\"}\n", out.String())
	out.Reset()
}

func TestSetLevel(t *testing.T) {
	out := &bytes.Buffer{}
	logger := zerolog.New(out)
	logConsumer := &LogConsumer{logger: &logger}

	logConsumer.SetLogLevel(logging.LogLevelDebug)
	assert.Equal(t, zerolog.DebugLevel, logConsumer.logger.GetLevel())

	logConsumer.SetLogLevel(logging.LogLevelInfo)
	assert.Equal(t, zerolog.InfoLevel, logConsumer.logger.GetLevel())

	logConsumer.SetLogLevel(logging.LogLevelWarning)
	assert.Equal(t, zerolog.WarnLevel, logConsumer.logger.GetLevel())

	logConsumer.SetLogLevel(logging.LogLevelError)
	assert.Equal(t, zerolog.ErrorLevel, logConsumer.logger.GetLevel())
}

func TestGetLoggerFromReqID(t *testing.T) {
	out := &bytes.Buffer{}
	logger := GetLoggerFromReqID("some_req_id")
	newLogger := logger.Output(out)
	newLogger.Info().Msg("some_message")

	assert.Contains(t, out.String(), `"requestID":"some_req_id"`)

}
