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
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	viper.SetConfigFile("./testdata/default.yaml")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	assert.NoError(t, err)

	conf := AgentConfig{}
	err = viper.Unmarshal(&conf)
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

	assert.True(t, conf.Api.Enabled)
	assert.Equal(t, 100, conf.Api.MaxConns)
	assert.Equal(t, "3000", conf.Api.Port)

	assert.True(t, conf.Webhook.Enabled)
	assert.Equal(t, "3001", conf.Webhook.Port)
	assert.Equal(t, "secret-10000", conf.Webhook.Projects[10000].Secret)
	assert.Equal(t, []string{"aaa", "bbb", "ccc"}, conf.Webhook.Projects[10000].SDKKeys)
	assert.True(t, conf.Webhook.Projects[10000].SkipSignatureCheck)
	assert.Equal(t, "secret-20000", conf.Webhook.Projects[20000].Secret)
	assert.Equal(t, []string{"xxx", "yyy", "zzz"}, conf.Webhook.Projects[20000].SDKKeys)
	assert.False(t, conf.Webhook.Projects[20000].SkipSignatureCheck)

	viper.Reset()
}
