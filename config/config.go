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
	"time"
)

type AgentConfig struct {
	Log     LogConfig     `yaml:"log"`
	Api     ApiConfig     `yaml:"api"`
	Admin   AdminConfig   `yaml:"admin"`
	Webhook WebhookConfig `yaml:"webhook"`
	Server  ServerConfig  `yaml:"server"`
}

type LogConfig struct {
	Pretty bool   `yaml:"pretty"`
	Level  string `yaml:"level"`
}

type ServerConfig struct {
	ReadTimeout  time.Duration `yaml:"readtimeout"`
	WriteTimeout time.Duration `yaml:"writetimeout"`
}

type ApiConfig struct {
	MaxConns int    `yaml:"maxconns"`
	Enabled  bool   `yaml:"enabled"`
	Port     string `yaml:"port"`
}

type AdminConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    string `yaml:"port"`
}

// WebhookConfig represents configuration of a single Optimizely webhook
type WebhookConfig struct {
	Enabled  bool                     `yaml:"enabled"`
	Port     string                   `yaml:"port"`
	Projects map[int64]WebhookProject `mapstructure:"projects"`
}

type WebhookProject struct {
	SDKKeys            []string `yaml:"sdkKeys"`
	Secret             string   `yaml:"secret"`
	SkipSignatureCheck bool     `yaml:"skipSignatureCheck" default:"false"`
}
