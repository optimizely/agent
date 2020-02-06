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
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/optimizely/go-sdk/pkg/logging"
)

var levelMap = map[logging.LogLevel]zerolog.Level{
	logging.LogLevelDebug:   zerolog.DebugLevel,
	logging.LogLevelInfo:    zerolog.InfoLevel,
	logging.LogLevelWarning: zerolog.WarnLevel,
	logging.LogLevelError:   zerolog.ErrorLevel,
}

// init overrides the Optimizely SDK logger with the default zerolog logger.
func init() {
	SetLogger(&log.Logger)
}

var LogManager = notification.NewAtomicManager()

type LogNotification struct {
	Level   string
	Message string
	Fields  map[string]interface{}
}

// SetLogger explicitly overwrites the zerolog used by the SDK with the provided zerolog logger.
func SetLogger(logger *zerolog.Logger) {
	logConsumer := &LogConsumer{
		logger: logger,
	}

	logging.SetLogger(logConsumer)
}

// LogConsumer is an implementation of the OptimizelyLogConsumer that wraps a zerolog logger
type LogConsumer struct {
	logger  *zerolog.Logger
	manager notification.Manager
}

// Log logs the message if it's log level is higher than or equal to the logger's set level
func (l *LogConsumer) Log(level logging.LogLevel, message string, fields map[string]interface{}) {
	// intercept levels > something. trigger another webhook
	if level > logging.LogLevelInfo {
		LogManager.Send(&LogNotification{
			Level:   level.String(),
			Message: message,
			Fields:  fields,
		})
	}

	l.logger.WithLevel(levelMap[level]).Fields(fields).Msg(message)
}

// SetLogLevel changes the log level to the given level
func (l *LogConsumer) SetLogLevel(level logging.LogLevel) {
	childLogger := l.logger.Level(levelMap[level])
	l.logger = &childLogger
}
