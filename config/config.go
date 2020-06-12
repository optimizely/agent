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
			Auth: ServiceAuthConfig{
				Clients:            make([]OAuthClientCredentials, 0),
				HMACSecrets:        make([]string, 0),
				TTL:                0,
				JwksURL:            "",
				JwksUpdateInterval: 0,
			},
			Port: "8088",
		},
		API: APIConfig{
			Auth: ServiceAuthConfig{
				Clients:            make([]OAuthClientCredentials, 0),
				HMACSecrets:        make([]string, 0),
				TTL:                0,
				JwksURL:            "",
				JwksUpdateInterval: 0,
			},
			CORS: CORSConfig{
				// If AllowedOrigins is nil or empty, value is set to ["*"].
				AllowedOrigins: nil,
				// If AllowedMethods is nil or empty, value is set to (HEAD, GET and POST).
				AllowedMethods: nil,
				// Default value is [] but "Origin" is always appended to the list.
				AllowedHeaders:     []string{},
				ExposedHeaders:     []string{},
				AllowedCredentials: false,
				MaxAge:             300,
			},
			MaxConns:            0,
			Port:                "8080",
			EnableNotifications: false,
			EnableOverrides:     false,
		},
		Log: LogConfig{
			Pretty: false,
			Level:  "info",
		},
		Client: ClientConfig{
			PollingInterval:     1 * time.Minute,
			BatchSize:           10,
			QueueSize:           1000,
			FlushInterval:       30 * time.Second,
			DatafileURLTemplate: "https://cdn.optimizely.com/datafiles/%s.json",
		},
		Server: ServerConfig{
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			HealthCheckPath: "/health",
			CertFile:        "",
			KeyFile:         "",
			DisabledCiphers: make([]string, 0),
		},
		Webhook: WebhookConfig{
			Port: "8085",
		},
	}

	return &config
}

// AgentConfig is the top level configuration struct
type AgentConfig struct {
	Version string `json:"version"`
	Author  string `json:"author"`
	Name    string `json:"name"`

	SDKKeys []string `yaml:"sdkKeys" json:"sdkKeys"`

	Admin   AdminConfig   `json:"admin"`
	API     APIConfig     `json:"api"`
	Log     LogConfig     `json:"log"`
	Client  ClientConfig  `json:"client"`
	Server  ServerConfig  `json:"server"`
	Webhook WebhookConfig `json:"webhook"`
}

// ClientConfig holds the configuration options for the Optimizely Client.
type ClientConfig struct {
	PollingInterval     time.Duration `json:"pollingInterval"`
	BatchSize           int           `json:"batchSize" default:"10"`
	QueueSize           int           `json:"queueSize" default:"1000"`
	FlushInterval       time.Duration `json:"flushInterval" default:"30s"`
	DatafileURLTemplate string        `json:"datafileURLTemplate"`
}

// LogConfig holds the log configuration
type LogConfig struct {
	Pretty bool   `json:"pretty"`
	Level  string `json:"level"`
}

// ServerConfig holds the global http server configs
type ServerConfig struct {
	ReadTimeout     time.Duration `json:"readTimeout"`
	WriteTimeout    time.Duration `json:"writeTimeout"`
	CertFile        string        `json:"certFile"`
	KeyFile         string        `json:"keyFile"`
	DisabledCiphers []string      `json:"disabledCiphers"`
	HealthCheckPath string        `json:"healthCheckPath"`
}

// APIConfig holds the REST API configuration
type APIConfig struct {
	Auth                ServiceAuthConfig `json:"-"`
	CORS                CORSConfig        `json:"cors"`
	MaxConns            int               `json:"maxConns"`
	Port                string            `json:"port"`
	EnableNotifications bool              `json:"enableNotifications"`
	EnableOverrides     bool              `json:"enableOverrides"`
}

// CORSConfig holds the CORS middleware configuration
type CORSConfig struct {
	AllowedOrigins     []string `json:"allowedOrigins"`
	AllowedMethods     []string `json:"allowedMethods"`
	AllowedHeaders     []string `json:"allowedHeaders"`
	ExposedHeaders     []string `json:"exposedHeaders"`
	AllowedCredentials bool     `json:"allowedCredentials"`
	MaxAge             int      `json:"maxAge"`
}

// AdminConfig holds the configuration for the admin web interface
type AdminConfig struct {
	Auth ServiceAuthConfig `json:"-"`
	Port string            `json:"port"`
}

// WebhookConfig holds configuration for Optimizely Webhooks
type WebhookConfig struct {
	Port     string                   `json:"port"`
	Projects map[int64]WebhookProject `json:"projects"`
}

// WebhookProject holds the configuration for a single Project webhook
type WebhookProject struct {
	SDKKeys            []string `json:"sdkKeys"`
	Secret             string   `json:"-"`
	SkipSignatureCheck bool     `json:"skipSignatureCheck" default:"false"`
}

// OAuthClientCredentials are used for issuing access tokens
type OAuthClientCredentials struct {
	ID         string   `yaml:"id"`
	SecretHash string   `yaml:"secretHash"`
	SDKKeys    []string `yaml:"sdkKeys"`
}

// ServiceAuthConfig holds the authentication configuration for a particular service
type ServiceAuthConfig struct {
	Clients            []OAuthClientCredentials `yaml:"clients" json:"-"`
	HMACSecrets        []string                 `yaml:"hmacSecrets" json:"-"`
	TTL                time.Duration            `yaml:"ttl" json:"-"`
	JwksURL            string                   `yaml:"jwksURL"`
	JwksUpdateInterval time.Duration            `yaml:"jwksUpdateInterval"`
}
