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
	assert.Equal(t, "health", conf.API.HealthEndPoint)
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
}
