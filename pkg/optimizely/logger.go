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
	"os"
	"github.com/rs/zerolog"

	"github.com/optimizely/go-sdk/optimizely/logging"
)

var levelMap = make(map[logging.LogLevel]zerolog.Level)

// init overrides the Optimizely SDK logger with a logrus implementation.
func init() {
	levelMap[logging.LogLevelDebug]   = zerolog.DebugLevel
	levelMap[logging.LogLevelInfo]    = zerolog.InfoLevel
	levelMap[logging.LogLevelWarning] = zerolog.WarnLevel
	levelMap[logging.LogLevelError]   = zerolog.ErrorLevel

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	logConsumer := &LogConsumer{
		logger: &logger,
	}

	logging.SetLogger(logConsumer)
}

// LogConsumer is an implementation of the OptimizelyLogConsumer that wraps a zerolog logger
type LogConsumer struct {
	logger *zerolog.Logger
}

// Log logs the message if it's log level is higher than or equal to the logger's set level
func (l *LogConsumer) Log(level logging.LogLevel, message string) {
	l.logger.WithLevel(levelMap[level]).Msg(message)
}

// SetLogLevel changes the log level to the given level
func (l *LogConsumer) SetLogLevel(level logging.LogLevel) {
	childLogger := l.logger.Level(levelMap[level])
	l.logger = &childLogger
}
