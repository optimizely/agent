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

package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestViperUnmarshal(t *testing.T) {
	v := viper.New()

	v.SetConfigFile("./testdata/default.yaml")
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	assert.NoError(t, err)

	conf := AgentConfig{}
	err = v.Unmarshal(&conf)
	assert.NoError(t, err)

	assert.Equal(t, 5*time.Second, conf.Server.ReadTimeout)
	assert.Equal(t, 10*time.Second, conf.Server.WriteTimeout)

	assert.True(t, conf.Log.Pretty)
	assert.Equal(t, "debug", conf.Log.Level)

	assert.False(t, conf.Admin.Enabled)
	assert.Equal(t, "3002", conf.Admin.Port)
	assert.Equal(t, "0.1.0", conf.Admin.Version)
	assert.Equal(t, "Optimizely Inc.", conf.Admin.Author)
	assert.Equal(t, "optimizely", conf.Admin.Name)

	assert.True(t, conf.API.Enabled)
	assert.Equal(t, 100, conf.API.MaxConns)
	assert.Equal(t, "3000", conf.API.Port)

	assert.True(t, conf.Webhook.Enabled)
	assert.Equal(t, "3001", conf.Webhook.Port)
	assert.Equal(t, "secret-10000", conf.Webhook.Projects[10000].Secret)
	assert.Equal(t, []string{"aaa", "bbb", "ccc"}, conf.Webhook.Projects[10000].SDKKeys)
	assert.True(t, conf.Webhook.Projects[10000].SkipSignatureCheck)
	assert.Equal(t, "secret-20000", conf.Webhook.Projects[20000].Secret)
	assert.Equal(t, []string{"xxx", "yyy", "zzz"}, conf.Webhook.Projects[20000].SDKKeys)
	assert.False(t, conf.Webhook.Projects[20000].SkipSignatureCheck)

	assert.Equal(t, []string{"ddd", "eee", "fff"}, conf.Optly.SDKKeys)
}

func TestViperProps(t *testing.T) {
	v := viper.New()

	v.Set("server.readtimeout", 5*time.Second)
	v.Set("server.writetimeout", 10*time.Second)

	v.Set("log.pretty", true)
	v.Set("log.level", "debug")

	v.Set("admin.port", "3002")
	v.Set("admin.version", "0.1.0")
	v.Set("admin.author", "Optimizely Inc.")
	v.Set("admin.name", "optimizely")

	v.Set("api.enabled", true)
	v.Set("api.maxconns", 100)
	v.Set("api.port", "3000")

	v.Set("webhook.enabled", true)
	v.Set("webhook.port", "3001")
	v.Set("webhook.projects.10000.secret", "secret-10000")
	v.Set("webhook.projects.10000.sdkkeys", []string{"aaa", "bbb", "ccc"})
	v.Set("webhook.projects.10000.skipsignaturecheck", true)
	v.Set("webhook.projects.20000.secret", "secret-20000")
	v.Set("webhook.projects.20000.sdkkeys", []string{"xxx", "yyy", "zzz"})
	v.Set("webhook.projects.20000.skipsignaturecheck", false)

	conf := AgentConfig{}
	err := v.Unmarshal(&conf)
	assert.NoError(t, err)

	assert.Equal(t, 5*time.Second, conf.Server.ReadTimeout)
	assert.Equal(t, 10*time.Second, conf.Server.WriteTimeout)

	assert.True(t, conf.Log.Pretty)
	assert.Equal(t, "debug", conf.Log.Level)

	assert.False(t, conf.Admin.Enabled)
	assert.Equal(t, "3002", conf.Admin.Port)
	assert.Equal(t, "0.1.0", conf.Admin.Version)
	assert.Equal(t, "Optimizely Inc.", conf.Admin.Author)
	assert.Equal(t, "optimizely", conf.Admin.Name)

	assert.True(t, conf.API.Enabled)
	assert.Equal(t, 100, conf.API.MaxConns)
	assert.Equal(t, "3000", conf.API.Port)

	assert.True(t, conf.Webhook.Enabled)
	assert.Equal(t, "3001", conf.Webhook.Port)
	assert.Equal(t, "secret-10000", conf.Webhook.Projects[10000].Secret)
	assert.Equal(t, []string{"aaa", "bbb", "ccc"}, conf.Webhook.Projects[10000].SDKKeys)
	assert.True(t, conf.Webhook.Projects[10000].SkipSignatureCheck)
	assert.Equal(t, "secret-20000", conf.Webhook.Projects[20000].Secret)
	assert.Equal(t, []string{"xxx", "yyy", "zzz"}, conf.Webhook.Projects[20000].SDKKeys)
	assert.False(t, conf.Webhook.Projects[20000].SkipSignatureCheck)
}

func TestDefaultConfig(t *testing.T) {
	conf := NewAgentConfig()

	assert.Equal(t, 5*time.Second, conf.Server.ReadTimeout)
	assert.Equal(t, 10*time.Second, conf.Server.WriteTimeout)

	assert.False(t, conf.Log.Pretty)
	assert.Equal(t, "info", conf.Log.Level)

	assert.True(t, conf.Admin.Enabled)
	assert.Equal(t, "8088", conf.Admin.Port)
	assert.Equal(t, "", conf.Admin.Version)
	assert.Equal(t, "Optimizely Inc.", conf.Admin.Author)
	assert.Equal(t, "optimizely", conf.Admin.Name)

	assert.True(t, conf.API.Enabled)
	assert.Equal(t, 0, conf.API.MaxConns)
	assert.Equal(t, "8080", conf.API.Port)

	assert.True(t, conf.Webhook.Enabled)
	assert.Equal(t, "8085", conf.Webhook.Port)
	assert.Empty(t, conf.Webhook.Projects)
}
