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

	"github.com/rs/zerolog/log"
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
			EventURL:            "https://logx.optimizely.com/v1/events",
			// https://github.com/google/re2/wiki/Syntax
			SdkKeyRegex: "^\\w+(:\\w+)?$",
		},
		Runtime: RuntimeConfig{
			BlockProfileRate:     0, // 0 is disabled
			MutexProfileFraction: 0, // 0 is disabled
		},

		Server: ServerConfig{
			AllowedHosts:    []string{"localhost"},
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			HealthCheckPath: "/health",
			CertFile:        "",
			KeyFile:         "",
			DisabledCiphers: make([]string, 0),
			Host:            "127.0.0.1",
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
	Runtime RuntimeConfig `json:"runtime"`
	Server  ServerConfig  `json:"server"`
	Webhook WebhookConfig `json:"webhook"`
}

// HTTPSDisabledWarning is logged when keyfile and certfile are not provided in server configuration
var HTTPSDisabledWarning = "keyfile and certfile not available, so server will use HTTP. For production deployments, it is recommended to either set keyfile and certfile for HTTPS, or run Agent behind a load balancer/reverse proxy that uses HTTPS."

// AuthDisabledWarningTemplate is used to log a warning when auth is disabled for API or Admin endpoints
var AuthDisabledWarningTemplate = "Authorization not enabled for %v endpoint. For production deployments, authorization is recommended."

// LogConfigWarnings checks this configuration and logs any relevant warnings.
func (ac *AgentConfig) LogConfigWarnings() {
	if !ac.Server.isHTTPSEnabled() {
		log.Warn().Msg(HTTPSDisabledWarning)
	}

	if !ac.API.Auth.isAuthorizationEnabled() {
		log.Warn().Msgf(AuthDisabledWarningTemplate, "API")
	}

	if !ac.Admin.Auth.isAuthorizationEnabled() {
		log.Warn().Msgf(AuthDisabledWarningTemplate, "Admin")
	}
}

// ClientConfig holds the configuration options for the Optimizely Client.
type ClientConfig struct {
	PollingInterval     time.Duration `json:"pollingInterval"`
	BatchSize           int           `json:"batchSize" default:"10"`
	QueueSize           int           `json:"queueSize" default:"1000"`
	FlushInterval       time.Duration `json:"flushInterval" default:"30s"`
	DatafileURLTemplate string        `json:"datafileURLTemplate"`
	EventURL            string        `json:"eventURL"`
	SdkKeyRegex         string        `json:"sdkKeyRegex"`
}

// LogConfig holds the log configuration
type LogConfig struct {
	Pretty bool   `json:"pretty"`
	Level  string `json:"level"`
}

// ServerConfig holds the global http server configs
type ServerConfig struct {
	AllowedHosts    []string      `json:"allowedHosts"`
	ReadTimeout     time.Duration `json:"readTimeout"`
	WriteTimeout    time.Duration `json:"writeTimeout"`
	CertFile        string        `json:"certFile"`
	KeyFile         string        `json:"keyFile"`
	DisabledCiphers []string      `json:"disabledCiphers"`
	HealthCheckPath string        `json:"healthCheckPath"`
	Host            string        `json:"host"`
}

func (sc *ServerConfig) isHTTPSEnabled() bool {
	return sc.KeyFile != "" && sc.CertFile != ""
}

// GetAllowedHosts returns the allowed hosts for this server. Requests whose host is not found in this slice should be
// rejected by the server.
func (sc *ServerConfig) GetAllowedHosts() []string {
	return append([]string{sc.Host}, sc.AllowedHosts...)
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

func (sc *ServiceAuthConfig) isAuthorizationEnabled() bool {
	return len(sc.HMACSecrets) > 0 || sc.JwksURL != ""
}

// RuntimeConfig holds any configuration related to the native runtime package
// These should only be configured when debugging in a non-production environment.
type RuntimeConfig struct {
	// SetBlockProfileRate controls the fraction of goroutine blocking events
	// that are reported in the blocking profile. The profiler aims to sample
	// an average of one blocking event per rate nanoseconds spent blocked.
	//
	// To include every blocking event in the profile, pass rate = 1.
	// To turn off profiling entirely, pass rate <= 0.
	BlockProfileRate int `json:"blockProfileRate"`

	// mutexProfileFraction controls the fraction of mutex contention events
	// that are reported in the mutex profile. On average 1/rate events are
	// reported. The previous rate is returned.
	//
	// To turn off profiling entirely, pass rate 0.
	// To just read the current rate, pass rate < 0.
	// (For n>1 the details of sampling may change.)
	MutexProfileFraction int `json:"mutexProfileFraction"`
}
