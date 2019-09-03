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
	"testing"
	log "github.com/sirupsen/logrus"

	"github.com/optimizely/go-sdk/optimizely/logging"
	"github.com/stretchr/testify/assert"
)

type TestLogger struct {
	filterLevel  log.Level
	messageLevel log.Level
	messages     []interface{}
}

func (l *TestLogger) Log(level log.Level, messages...interface{}) {
	l.messageLevel = level
	l.messages = messages
}

func (l *TestLogger) SetLevel(level log.Level) {
	l.filterLevel = level
}

func TestLog(t *testing.T) {
	logger := new(TestLogger)
	logrusLogConsumer := &LogrusLogConsumer{logger: logger}

	logrusLogConsumer.Log(logging.LogLevelDebug, "debug")
	assert.Equal(t, logger.messageLevel, log.DebugLevel)
	assert.Equal(t, logger.messages[0], "debug")

	logrusLogConsumer.Log(logging.LogLevelInfo, "info")
	assert.Equal(t, logger.messageLevel, log.InfoLevel)
	assert.Equal(t, logger.messages[0], "info")

	logrusLogConsumer.Log(logging.LogLevelWarning, "warn")
	assert.Equal(t, logger.messageLevel, log.WarnLevel)
	assert.Equal(t, logger.messages[0], "warn")

	logrusLogConsumer.Log(logging.LogLevelError, "error")
	assert.Equal(t, logger.messageLevel, log.ErrorLevel)
	assert.Equal(t, logger.messages[0], "error")
}

func TestSetLevel(t *testing.T) {
	logger := new(TestLogger)
	logrusLogConsumer := &LogrusLogConsumer{logger: logger}

	logrusLogConsumer.SetLogLevel(logging.LogLevelDebug)
	assert.Equal(t, logger.filterLevel, log.DebugLevel)

	logrusLogConsumer.SetLogLevel(logging.LogLevelInfo)
	assert.Equal(t, logger.filterLevel, log.InfoLevel)

	logrusLogConsumer.SetLogLevel(logging.LogLevelWarning)
	assert.Equal(t, logger.filterLevel, log.WarnLevel)

	logrusLogConsumer.SetLogLevel(logging.LogLevelError)
	assert.Equal(t, logger.filterLevel, log.ErrorLevel)
}