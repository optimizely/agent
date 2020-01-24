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

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	conf := NewDefaultConfig()

	assert.Equal(t, "", conf.Version)
	assert.Equal(t, "Optimizely Inc.", conf.Author)
	assert.Equal(t, "optimizely", conf.Name)

	assert.Equal(t, 5*time.Second, conf.Server.ReadTimeout)
	assert.Equal(t, 10*time.Second, conf.Server.WriteTimeout)

	assert.False(t, conf.Log.Pretty)
	assert.Equal(t, "info", conf.Log.Level)

	assert.Equal(t, "8088", conf.Admin.Port)

	assert.Equal(t, 0, conf.API.MaxConns)
	assert.Equal(t, "8080", conf.API.Port)

	assert.Equal(t, "8085", conf.Webhook.Port)
	assert.Empty(t, conf.Webhook.Projects)

	assert.Equal(t, 10, conf.Processor.BatchSize)
	assert.Equal(t, 1000, conf.Processor.QueueSize)
	assert.Equal(t, 30*time.Second, conf.Processor.FlushInterval)
}
