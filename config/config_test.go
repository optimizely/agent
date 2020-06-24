/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
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

package config

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	conf := NewDefaultConfig()

	assert.Equal(t, "", conf.Version)
	assert.Equal(t, "Optimizely Inc.", conf.Author)
	assert.Equal(t, "optimizely", conf.Name)

	assert.Equal(t, 5*time.Second, conf.Server.ReadTimeout)
	assert.Equal(t, 10*time.Second, conf.Server.WriteTimeout)
	assert.Equal(t, "/health", conf.Server.HealthCheckPath)
	assert.Equal(t, "", conf.Server.KeyFile)
	assert.Equal(t, "", conf.Server.CertFile)
	assert.Equal(t, []string{}, conf.Server.DisabledCiphers)

	assert.False(t, conf.Log.Pretty)
	assert.Equal(t, "info", conf.Log.Level)

	assert.Equal(t, "8088", conf.Admin.Port)
	assert.Equal(t, make([]OAuthClientCredentials, 0), conf.Admin.Auth.Clients)
	assert.Equal(t, make([]string, 0), conf.Admin.Auth.HMACSecrets)
	assert.Equal(t, time.Duration(0), conf.Admin.Auth.TTL)
	assert.Equal(t, "", conf.Admin.Auth.JwksURL)
	assert.Equal(t, time.Duration(0), conf.Admin.Auth.JwksUpdateInterval)

	assert.Equal(t, 0, conf.API.MaxConns)
	assert.Equal(t, "8080", conf.API.Port)
	assert.Equal(t, make([]OAuthClientCredentials, 0), conf.API.Auth.Clients)
	assert.Equal(t, make([]string, 0), conf.API.Auth.HMACSecrets)
	assert.Equal(t, time.Duration(0), conf.API.Auth.TTL)
	assert.Equal(t, "", conf.API.Auth.JwksURL)
	assert.Equal(t, time.Duration(0), conf.API.Auth.JwksUpdateInterval)
	assert.Equal(t, false, conf.API.EnableOverrides)
	assert.Equal(t, false, conf.API.EnableNotifications)
	assert.Equal(t, []string(nil), conf.API.CORS.AllowedOrigins)
	assert.Equal(t, []string(nil), conf.API.CORS.AllowedMethods)
	assert.Equal(t, make([]string, 0), conf.API.CORS.AllowedHeaders)
	assert.Equal(t, make([]string, 0), conf.API.CORS.ExposedHeaders)
	assert.Equal(t, false, conf.API.CORS.AllowedCredentials)
	assert.Equal(t, 300, conf.API.CORS.MaxAge)

	assert.Equal(t, "8085", conf.Webhook.Port)
	assert.Empty(t, conf.Webhook.Projects)

	assert.Equal(t, 1*time.Minute, conf.Client.PollingInterval)
	assert.Equal(t, 10, conf.Client.BatchSize)
	assert.Equal(t, 1000, conf.Client.QueueSize)
	assert.Equal(t, 30*time.Second, conf.Client.FlushInterval)
	assert.Equal(t, "https://cdn.optimizely.com/datafiles/%s.json", conf.Client.DatafileURLTemplate)
	assert.Equal(t, "https://logx.optimizely.com/v1/events", conf.Client.EventURL)
}

type logObservation struct {
	msg   string
	level zerolog.Level
}

type testLogHook struct {
	logs []*logObservation
}

func (th *testLogHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	th.logs = append(th.logs, &logObservation{msg, level})
}

func (th *testLogHook) messages() []string {
	logMessages := []string{}
	for _, obs := range th.logs {
		logMessages = append(logMessages, obs.msg)
	}
	return logMessages
}

type LogConfigurationWarningsTestSuite struct {
	suite.Suite
	hook         *testLogHook
	globalLogger zerolog.Logger
}

func (s *LogConfigurationWarningsTestSuite) SetupTest() {
	s.hook = &testLogHook{}
	// Replace global logger for this test suite
	s.globalLogger = log.Logger
	log.Logger = log.Hook(s.hook)
}

func (s *LogConfigurationWarningsTestSuite) TearDownTest() {
	// Restore global logger to original state
	log.Logger = s.globalLogger
}

func (s *LogConfigurationWarningsTestSuite) TestLogConfigurationWarningsHTTPSNotSet() {
	conf := NewDefaultConfig()
	conf.Server.KeyFile = ""
	conf.Server.CertFile = ""

	conf.LogConfigurationWarnings()

	s.Contains(s.hook.logs, &logObservation{
		msg:   HTTPSDisabledWarning,
		level: zerolog.WarnLevel,
	})
}

func (s *LogConfigurationWarningsTestSuite) TestLogConfigurationWarningsHTTPSSet() {
	conf := NewDefaultConfig()
	conf.Server.KeyFile = "/path/to/keyfile"
	conf.Server.CertFile = "/path/to/certfile"

	conf.LogConfigurationWarnings()

	s.NotContains(s.hook.messages(), HTTPSDisabledWarning)
}

func (s *LogConfigurationWarningsTestSuite) TestLogConfigurationWarningsAuthNotSet() {
	conf := NewDefaultConfig()
	conf.API.Auth.JwksURL = ""
	conf.API.Auth.HMACSecrets = []string{}

	conf.LogConfigurationWarnings()

	s.Contains(s.hook.logs, &logObservation{
		msg:   fmt.Sprintf(AuthDisabledWarningTemplate, "API"),
		level: zerolog.WarnLevel,
	})
	s.Contains(s.hook.logs, &logObservation{
		msg:   fmt.Sprintf(AuthDisabledWarningTemplate, "Admin"),
		level: zerolog.WarnLevel,
	})
}

func (s *LogConfigurationWarningsTestSuite) TestLogConfigurationWarningsJWKSUrlSetForAPI() {
	conf := NewDefaultConfig()
	conf.API.Auth.JwksURL = "https://YOUR_DOMAIN/.well-known/jwks.json"

	conf.LogConfigurationWarnings()

	s.NotContains(s.hook.messages(), fmt.Sprintf(AuthDisabledWarningTemplate, "API"))
	s.Contains(s.hook.logs, &logObservation{
		msg:   fmt.Sprintf(AuthDisabledWarningTemplate, "Admin"),
		level: zerolog.WarnLevel,
	})
}
func (s *LogConfigurationWarningsTestSuite) TestLogConfigurationWarningsHMACSecretsSetForAdmin() {
	conf := NewDefaultConfig()
	conf.Admin.Auth.HMACSecrets = []string{"abcd123"}

	conf.LogConfigurationWarnings()

	s.Contains(s.hook.logs, &logObservation{
		msg:   fmt.Sprintf(AuthDisabledWarningTemplate, "API"),
		level: zerolog.WarnLevel,
	})
	s.NotContains(s.hook.messages(), fmt.Sprintf(AuthDisabledWarningTemplate, "Admin"))
}

func (s *LogConfigurationWarningsTestSuite) TestLogConfigurationWarningsAuthSetForBoth() {
	conf := NewDefaultConfig()
	conf.API.Auth.HMACSecrets = []string{"abcd123"}
	conf.Admin.Auth.HMACSecrets = []string{"abcd123"}

	conf.LogConfigurationWarnings()

	messages := s.hook.messages()
	s.NotContains(messages, fmt.Sprintf(AuthDisabledWarningTemplate, "API"))
	s.NotContains(messages, fmt.Sprintf(AuthDisabledWarningTemplate, "Admin"))
}

func TestLogConfigurationWarnings(t *testing.T) {
	suite.Run(t, new(LogConfigurationWarningsTestSuite))
}
