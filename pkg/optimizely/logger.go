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
	log "github.com/sirupsen/logrus"

	"github.com/optimizely/go-sdk/optimizely/logging"
)

var levelMap = make(map[logging.LogLevel]log.Level)

func init() {
	levelMap[logging.LogLevelDebug]   = log.DebugLevel
	levelMap[logging.LogLevelInfo]    = log.InfoLevel
	levelMap[logging.LogLevelWarning] = log.WarnLevel
	levelMap[logging.LogLevelError]   = log.ErrorLevel

	logger := log.New()
	logger.SetLevel(log.InfoLevel)

	logConsumer := &LogrusLogConsumer{
		logger: logger,
	}

	logging.SetLogger(logConsumer)
}

// Logger interface is primarily used to fascilitate testing
type Logger interface {
	Log(log.Level, ...interface{})
	SetLevel(log.Level)
}

// LogrusLogConsumer is an implementation of the OptimizelyLogConsumer that wraps a logrus logger
type LogrusLogConsumer struct {
	logger Logger
}

// Log logs the message if it's log level is higher than or equal to the logger's set level
func (l *LogrusLogConsumer) Log(level logging.LogLevel, message string) {
	l.logger.Log(levelMap[level], message)
}

// SetLogLevel changes the log level to the given level
func (l *LogrusLogConsumer) SetLogLevel(level logging.LogLevel) {
	l.logger.SetLevel(levelMap[level])
}
