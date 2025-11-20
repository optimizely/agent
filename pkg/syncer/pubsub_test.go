/****************************************************************************
 * Copyright 2023 Optimizely, Inc. and contributors                         *
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

// Package syncer provides synchronization across Agent nodes
package syncer

import (
	"reflect"
	"testing"
	"time"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/syncer/pubsub"
)

func TestNewPubSub(t *testing.T) {
	type args struct {
		conf config.SyncConfig
		flag SyncFeatureFlag
	}
	tests := []struct {
		name    string
		args    args
		want    PubSub
		wantErr bool
	}{
		{
			name: "Test with valid config",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": map[string]interface{}{
							"host":     "localhost:6379",
							"password": "",
							"database": 0,
						},
					},
					Notification: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
				flag: SyncFeatureFlagNotification,
			},
			want: &pubsub.RedisStreams{
				Host:          "localhost:6379",
				Password:      "",
				Database:      0,
				BatchSize:     10,
				FlushInterval: 5 * time.Second,
				MaxRetries:    3,
				RetryDelay:    100 * time.Millisecond,
				MaxRetryDelay: 5 * time.Second,
				ConnTimeout:   10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "Test with valid config for datafile",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": map[string]interface{}{
							"host":     "localhost:6379",
							"password": "",
							"database": 0,
						},
					},
					Datafile: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
				flag: SyncFeatureFlagDatafile,
			},
			want: &pubsub.RedisStreams{
				Host:          "localhost:6379",
				Password:      "",
				Database:      0,
				BatchSize:     10,
				FlushInterval: 5 * time.Second,
				MaxRetries:    3,
				RetryDelay:    100 * time.Millisecond,
				MaxRetryDelay: 5 * time.Second,
				ConnTimeout:   10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "Test with invalid config",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"nopt-redis": map[string]interface{}{},
					},
					Notification: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
				flag: SyncFeatureFlagNotification,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test with nil config",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": nil,
					},
					Notification: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
				flag: SyncFeatureFlagNotification,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test with empty config",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": nil,
					},
					Notification: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
				flag: SyncFeatureFlagNotification,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test with invalid redis config",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": map[string]interface{}{
							"host":     123,
							"password": "",
							"database": "invalid-db",
						},
					},
					Notification: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
				flag: SyncFeatureFlagNotification,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test with invalid redis config without host",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": map[string]interface{}{
							"password": "",
							"database": 0,
						},
					},
					Notification: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
				flag: SyncFeatureFlagNotification,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test with valid redis config without password",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": map[string]interface{}{
							"host":     "localhost:6379",
							"database": 0,
						},
					},
					Notification: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
				flag: SyncFeatureFlagNotification,
			},
			want: &pubsub.RedisStreams{
				Host:          "localhost:6379",
				Password:      "", // Empty password is valid (no auth required)
				Database:      0,
				BatchSize:     10,
				FlushInterval: 5 * time.Second,
				MaxRetries:    3,
				RetryDelay:    100 * time.Millisecond,
				MaxRetryDelay: 5 * time.Second,
				ConnTimeout:   10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "Test with invalid redis config without db",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": map[string]interface{}{
							"host":     "localhost:6379",
							"password": "",
						},
					},
					Notification: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
				flag: SyncFeatureFlagNotification,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test with redis config with invalid password type (ignored)",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": map[string]interface{}{
							"host":     "localhost:6379",
							"password": 1234, // Invalid type, will be ignored
							"database": 0,
						},
					},
					Notification: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
				flag: SyncFeatureFlagNotification,
			},
			want: &pubsub.RedisStreams{
				Host:          "localhost:6379",
				Password:      "", // Invalid type ignored, falls back to empty string
				Database:      0,
				BatchSize:     10,
				FlushInterval: 5 * time.Second,
				MaxRetries:    3,
				RetryDelay:    100 * time.Millisecond,
				MaxRetryDelay: 5 * time.Second,
				ConnTimeout:   10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "Test with invalid redis config with invalid database",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": map[string]interface{}{
							"host":     "localhost:6379",
							"password": "",
							"database": "invalid-db",
						},
					},
					Notification: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
				flag: SyncFeatureFlagNotification,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test with auto-detected redis-streams with custom config",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": map[string]interface{}{
							"host":               "localhost:6379",
							"password":           "",
							"database":           0,
							"batch_size":         20,
							"flush_interval":     "10s",
							"max_retries":        5,
							"retry_delay":        "200ms",
							"max_retry_delay":    "10s",
							"connection_timeout": "15s",
						},
					},
					Notification: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
				flag: SyncFeatureFlagNotification,
			},
			want: &pubsub.RedisStreams{
				Host:          "localhost:6379",
				Password:      "",
				Database:      0,
				BatchSize:     20,
				FlushInterval: 10000000000, // 10s in nanoseconds
				MaxRetries:    5,
				RetryDelay:    200000000,   // 200ms in nanoseconds
				MaxRetryDelay: 10000000000, // 10s in nanoseconds
				ConnTimeout:   15000000000, // 15s in nanoseconds
			},
			wantErr: false,
		},
		{
			name: "Test with auto-detected redis-streams for datafile",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": map[string]interface{}{
							"host":     "localhost:6379",
							"password": "",
							"database": 0,
						},
					},
					Datafile: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
				flag: SyncFeatureFlagDatafile,
			},
			want: &pubsub.RedisStreams{
				Host:          "localhost:6379",
				Password:      "",
				Database:      0,
				BatchSize:     10,          // default
				FlushInterval: 5000000000,  // 5s default in nanoseconds
				MaxRetries:    3,           // default
				RetryDelay:    100000000,   // 100ms default in nanoseconds
				MaxRetryDelay: 5000000000,  // 5s default in nanoseconds
				ConnTimeout:   10000000000, // 10s default in nanoseconds
			},
			wantErr: false,
		},
		{
			name: "Test with unsupported pubsub type",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": map[string]interface{}{
							"host":     "localhost:6379",
							"password": "",
							"database": 0,
						},
					},
					Notification: config.FeatureSyncConfig{
						Default: "unsupported-type",
						Enable:  true,
					},
				},
				flag: SyncFeatureFlagNotification,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test with invalid feature flag",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": map[string]interface{}{
							"host":     "localhost:6379",
							"password": "",
							"database": 0,
						},
					},
					Notification: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
				flag: "invalid-flag",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newPubSub(tt.args.conf, tt.args.flag)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPubSub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPubSub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetIntFromConfig(t *testing.T) {
	tests := []struct {
		name         string
		config       map[string]interface{}
		key          string
		defaultValue int
		want         int
	}{
		{
			name: "Valid int value",
			config: map[string]interface{}{
				"test_key": 42,
			},
			key:          "test_key",
			defaultValue: 10,
			want:         42,
		},
		{
			name: "Missing key returns default",
			config: map[string]interface{}{
				"other_key": 42,
			},
			key:          "test_key",
			defaultValue: 10,
			want:         10,
		},
		{
			name: "Invalid type returns default",
			config: map[string]interface{}{
				"test_key": "not an int",
			},
			key:          "test_key",
			defaultValue: 10,
			want:         10,
		},
		{
			name: "Nil value returns default",
			config: map[string]interface{}{
				"test_key": nil,
			},
			key:          "test_key",
			defaultValue: 10,
			want:         10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getIntFromConfig(tt.config, tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getIntFromConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDurationFromConfig(t *testing.T) {
	tests := []struct {
		name         string
		config       map[string]interface{}
		key          string
		defaultValue time.Duration
		want         time.Duration
	}{
		{
			name: "Valid duration string",
			config: map[string]interface{}{
				"test_key": "5s",
			},
			key:          "test_key",
			defaultValue: 1 * time.Second,
			want:         5 * time.Second,
		},
		{
			name: "Valid millisecond duration",
			config: map[string]interface{}{
				"test_key": "100ms",
			},
			key:          "test_key",
			defaultValue: 1 * time.Second,
			want:         100 * time.Millisecond,
		},
		{
			name: "Missing key returns default",
			config: map[string]interface{}{
				"other_key": "5s",
			},
			key:          "test_key",
			defaultValue: 1 * time.Second,
			want:         1 * time.Second,
		},
		{
			name: "Invalid duration string returns default",
			config: map[string]interface{}{
				"test_key": "invalid duration",
			},
			key:          "test_key",
			defaultValue: 1 * time.Second,
			want:         1 * time.Second,
		},
		{
			name: "Non-string value returns default",
			config: map[string]interface{}{
				"test_key": 123,
			},
			key:          "test_key",
			defaultValue: 1 * time.Second,
			want:         1 * time.Second,
		},
		{
			name: "Nil value returns default",
			config: map[string]interface{}{
				"test_key": nil,
			},
			key:          "test_key",
			defaultValue: 1 * time.Second,
			want:         1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDurationFromConfig(tt.config, tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getDurationFromConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPubSub_DatabaseTypeConversion(t *testing.T) {
	tests := []struct {
		name     string
		database interface{}
		wantErr  bool
	}{
		{
			name:     "database as int",
			database: 0,
			wantErr:  false,
		},
		{
			name:     "database as float64 (from YAML/JSON)",
			database: float64(0),
			wantErr:  false,
		},
		{
			name:     "database as float64 non-zero",
			database: float64(1),
			wantErr:  false,
		},
		{
			name:     "database as string - should fail",
			database: "0",
			wantErr:  true,
		},
		{
			name:     "database as nil - should fail",
			database: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := config.SyncConfig{
				Pubsub: map[string]interface{}{
					"redis": map[string]interface{}{
						"host":     "localhost:6379",
						"password": "",
						"database": tt.database,
					},
				},
				Notification: config.FeatureSyncConfig{
					Default: "redis",
					Enable:  true,
				},
			}

			_, err := newPubSub(conf, SyncFeatureFlagNotification)
			if (err != nil) != tt.wantErr {
				t.Errorf("newPubSub() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetPubSubRedisStreams_ErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		conf    config.SyncConfig
		wantErr bool
	}{
		{
			name: "redis config not found",
			conf: config.SyncConfig{
				Pubsub: map[string]interface{}{
					"not-redis": map[string]interface{}{},
				},
				Notification: config.FeatureSyncConfig{
					Default: "redis",
					Enable:  true,
				},
			},
			wantErr: true,
		},
		{
			name: "redis config not valid (not a map)",
			conf: config.SyncConfig{
				Pubsub: map[string]interface{}{
					"redis": "invalid-config",
				},
				Notification: config.FeatureSyncConfig{
					Default: "redis",
					Enable:  true,
				},
			},
			wantErr: true,
		},
		{
			name: "redis host not found",
			conf: config.SyncConfig{
				Pubsub: map[string]interface{}{
					"redis": map[string]interface{}{
						"password": "",
						"database": 0,
					},
				},
				Notification: config.FeatureSyncConfig{
					Default: "redis",
					Enable:  true,
				},
			},
			wantErr: true,
		},
		{
			name: "redis host not valid (not a string)",
			conf: config.SyncConfig{
				Pubsub: map[string]interface{}{
					"redis": map[string]interface{}{
						"host":     123,
						"password": "",
						"database": 0,
					},
				},
				Notification: config.FeatureSyncConfig{
					Default: "redis",
					Enable:  true,
				},
			},
			wantErr: true,
		},
		{
			name: "redis database not found",
			conf: config.SyncConfig{
				Pubsub: map[string]interface{}{
					"redis": map[string]interface{}{
						"host":     "localhost:6379",
						"password": "",
					},
				},
				Notification: config.FeatureSyncConfig{
					Default: "redis",
					Enable:  true,
				},
			},
			wantErr: true,
		},
		{
			name: "redis database as float64 (valid)",
			conf: config.SyncConfig{
				Pubsub: map[string]interface{}{
					"redis": map[string]interface{}{
						"host":     "localhost:6379",
						"password": "",
						"database": float64(1),
					},
				},
				Notification: config.FeatureSyncConfig{
					Default: "redis",
					Enable:  true,
				},
			},
			wantErr: false,
		},
		{
			name: "redis database invalid type",
			conf: config.SyncConfig{
				Pubsub: map[string]interface{}{
					"redis": map[string]interface{}{
						"host":     "localhost:6379",
						"password": "",
						"database": "invalid",
					},
				},
				Notification: config.FeatureSyncConfig{
					Default: "redis",
					Enable:  true,
				},
			},
			wantErr: true,
		},
		{
			name: "datafile with unsupported pubsub type",
			conf: config.SyncConfig{
				Pubsub: map[string]interface{}{
					"redis": map[string]interface{}{
						"host":     "localhost:6379",
						"password": "",
						"database": 0,
					},
				},
				Datafile: config.FeatureSyncConfig{
					Default: "unsupported-type",
					Enable:  true,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.conf.Notification.Default != "" {
				_, err = newPubSub(tt.conf, SyncFeatureFlagNotification)
			} else {
				_, err = newPubSub(tt.conf, SyncFeatureFlagDatafile)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("newPubSub() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
