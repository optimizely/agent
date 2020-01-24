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

// Package config contains all the configuration attributes for running Optimizely Agent
package config

import (
	"time"
)

// NewDefaultConfig returns the default configuration for Optimizely Agent
func NewDefaultConfig() *AgentConfig {

	config := AgentConfig{
		Version: "",
		Author:  "Optimizely Inc.",
		Name:    "optimizely",

		Admin: AdminConfig{
			Port: "8088",
		},
		API: APIConfig{
			MaxConns: 0,
			Port:     "8080",
		},
		Log: LogConfig{
			Pretty: false,
			Level:  "info",
		},
		Processor: ProcessorConfig{
			BatchSize:     10,
			QueueSize:     1000,
			FlushInterval: 30 * time.Second,
		},
		Server: ServerConfig{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		Webhook: WebhookConfig{
			Port: "8085",
		},
	}

	return &config
}

// AgentConfig is the top level configuration struct
type AgentConfig struct {
	Version string `yaml:"version" json:"version"`
	Author  string `yaml:"author" json:"author"`
	Name    string `yaml:"name" json:"name"`

	SDKKeys []string `yaml:"sdkkeys" json:"sdkkeys"`

	Admin     AdminConfig     `yaml:"admin" json:"admin"`
	API       APIConfig       `yaml:"api" json:"api"`
	Log       LogConfig       `yaml:"log" json:"log"`
	Processor ProcessorConfig `yaml:"processor" json:"processor"`
	Server    ServerConfig    `yaml:"server" json:"server"`
	Webhook   WebhookConfig   `yaml:"webhook" json:"webhook"`
}

// OptlyConfig holds the set of SDK keys to bootstrap during initialization
type OptlyConfig struct {
	Processor ProcessorConfig `yaml:"processor" json:"processor"`
}

// ProcessorConfig holds the configuration options for the Optimizely Event Processor.
type ProcessorConfig struct {
	BatchSize     int           `yaml:"batchSize" json:"batchSize" default:"10"`
	QueueSize     int           `yaml:"queueSize" json:"queueSize" default:"1000"`
	FlushInterval time.Duration `yaml:"flushInterval" json:"flushInterval" default:"30s"`
}

// LogConfig holds the log configuration
type LogConfig struct {
	Pretty bool   `yaml:"pretty" json:"pretty"`
	Level  string `yaml:"level" json:"level"`
}

// ServerConfig holds the global http server configs
type ServerConfig struct {
	ReadTimeout  time.Duration `yaml:"readtimeout" json:"readtimeout"`
	WriteTimeout time.Duration `yaml:"writetimeout" json:"writetimeout"`
}

// APIConfig holds the REST API configuration
type APIConfig struct {
	MaxConns int    `yaml:"maxconns" json:"maxconns"`
	Port     string `yaml:"port" json:"port"`
}

// AdminConfig holds the configuration for the admin web interface
type AdminConfig struct {
	Port string `yaml:"port" json:"port"`
}

// WebhookConfig holds configuration for Optimizely Webhooks
type WebhookConfig struct {
	Port     string                   `yaml:"port" json:"port"`
	Projects map[int64]WebhookProject `mapstructure:"projects" json:"projects"`
}

// WebhookProject holds the configuration for a single Project webhook
type WebhookProject struct {
	SDKKeys            []string `yaml:"sdkKeys" json:"sdkKeys"`
	Secret             string   `yaml:"secret" json:"-"`
	SkipSignatureCheck bool     `yaml:"skipSignatureCheck" json:"skipSignatureCheck" default:"false"`
}
