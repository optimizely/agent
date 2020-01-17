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
		Admin: AdminConfig{
			Version: "",
			Author:  "Optimizely Inc.",
			Name:    "optimizely",
			Port:    "8088",
		},
		API: APIConfig{
			MaxConns: 0,
			Port:     "8080",
		},
		Log: LogConfig{
			Pretty: false,
			Level:  "info",
		},
		Optly: OptlyConfig{
			Processor: ProcessorConfig{
				BatchSize:     10,
				QueueSize:     1000,
				FlushInterval: 30 * time.Second,
			},
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
	Admin   AdminConfig   `yaml:"admin"`
	API     APIConfig     `yaml:"api"`
	Log     LogConfig     `yaml:"log"`
	Optly   OptlyConfig   `yaml:"optly"`
	Server  ServerConfig  `yaml:"server"`
	Webhook WebhookConfig `yaml:"webhook"`
}

// OptlyConfig holds the set of SDK keys to bootstrap during initialization
type OptlyConfig struct {
	Processor ProcessorConfig `yaml:"processor"`
	SDKKeys   []string        `yaml:"sdkkeys"`
}

// ProcessorConfig holds the configuration options for the Optimizely Event Processor.
type ProcessorConfig struct {
	BatchSize     int           `yaml:"batchSize" default:"10"`
	QueueSize     int           `yaml:"queueSize" default:"1000"`
	FlushInterval time.Duration `yaml:"flushInterval" default:"30s"`
}

// LogConfig holds the log configuration
type LogConfig struct {
	Pretty bool   `yaml:"pretty"`
	Level  string `yaml:"level"`
}

// ServerConfig holds the global http server configs
type ServerConfig struct {
	ReadTimeout  time.Duration `yaml:"readtimeout"`
	WriteTimeout time.Duration `yaml:"writetimeout"`
}

// APIConfig holds the REST API configuration
type APIConfig struct {
	MaxConns int    `yaml:"maxconns"`
	Port     string `yaml:"port"`
}

// AdminConfig holds the configuration for the admin web interface
type AdminConfig struct {
	Version string `yaml:"version"`
	Author  string `yaml:"author"`
	Name    string `yaml:"name"`
	Port    string `yaml:"port"`
}

// WebhookConfig holds configuration for Optimizely Webhooks
type WebhookConfig struct {
	Port     string                   `yaml:"port"`
	Projects map[int64]WebhookProject `mapstructure:"projects"`
}

// WebhookProject holds the configuration for a single Project webhook
type WebhookProject struct {
	SDKKeys            []string `yaml:"sdkKeys"`
	Secret             string   `yaml:"secret"`
	SkipSignatureCheck bool     `yaml:"skipSignatureCheck" default:"false"`
}
