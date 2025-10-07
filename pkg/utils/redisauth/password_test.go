/****************************************************************************
 * Copyright 2025 Optimizely, Inc. and contributors                         *
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

package redisauth

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPassword(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		envVar   string
		envValue string
		want     string
	}{
		{
			name: "auth_token has highest priority",
			config: map[string]interface{}{
				"auth_token":   "token123",
				"redis_secret": "secret456",
				"password":     "password789",
			},
			envVar: "TEST_ENV",
			want:   "token123",
		},
		{
			name: "redis_secret used when auth_token missing",
			config: map[string]interface{}{
				"redis_secret": "secret456",
				"password":     "password789",
			},
			envVar: "TEST_ENV",
			want:   "secret456",
		},
		{
			name: "password used when auth_token and redis_secret missing",
			config: map[string]interface{}{
				"password": "password789",
			},
			envVar: "TEST_ENV",
			want:   "password789",
		},
		{
			name: "environment variable used when no config fields present",
			config: map[string]interface{}{
				"host":     "localhost:6379",
				"database": 0,
			},
			envVar:   "TEST_ENV",
			envValue: "env_password",
			want:     "env_password",
		},
		{
			name: "empty string when no password configured",
			config: map[string]interface{}{
				"host":     "localhost:6379",
				"database": 0,
			},
			envVar: "TEST_ENV",
			want:   "",
		},
		{
			name: "empty field values are ignored",
			config: map[string]interface{}{
				"auth_token":   "",
				"redis_secret": "",
				"password":     "password789",
			},
			envVar: "TEST_ENV",
			want:   "password789",
		},
		{
			name: "non-string values are ignored",
			config: map[string]interface{}{
				"auth_token": 12345, // Invalid type
				"password":   "password789",
			},
			envVar: "TEST_ENV",
			want:   "password789",
		},
		{
			name: "config fields take precedence over env var",
			config: map[string]interface{}{
				"password": "config_password",
			},
			envVar:   "TEST_ENV",
			envValue: "env_password",
			want:     "config_password",
		},
		{
			name: "empty env var name is handled gracefully",
			config: map[string]interface{}{
				"host": "localhost:6379",
			},
			envVar: "",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if needed
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			} else {
				// Ensure env var is not set
				os.Unsetenv(tt.envVar)
			}

			got := GetPassword(tt.config, tt.envVar)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetPassword_RealWorldScenarios(t *testing.T) {
	t.Run("Kubernetes secret via env var", func(t *testing.T) {
		os.Setenv("REDIS_PASSWORD", "k8s-secret-value")
		defer os.Unsetenv("REDIS_PASSWORD")

		config := map[string]interface{}{
			"host":     "redis-service:6379",
			"database": 0,
		}

		got := GetPassword(config, "REDIS_PASSWORD")
		assert.Equal(t, "k8s-secret-value", got)
	})

	t.Run("Development config without auth", func(t *testing.T) {
		config := map[string]interface{}{
			"host":     "localhost:6379",
			"database": 0,
		}

		got := GetPassword(config, "REDIS_PASSWORD")
		assert.Equal(t, "", got)
	})

	t.Run("Production config with auth_token", func(t *testing.T) {
		config := map[string]interface{}{
			"host":       "redis.production.example.com:6379",
			"auth_token": "prod-token-12345",
			"database":   1,
		}

		got := GetPassword(config, "REDIS_PASSWORD")
		assert.Equal(t, "prod-token-12345", got)
	})

	t.Run("Legacy config with password field", func(t *testing.T) {
		config := map[string]interface{}{
			"host":     "legacy-redis:6379",
			"password": "legacy-pass",
			"database": 0,
		}

		got := GetPassword(config, "REDIS_PASSWORD")
		assert.Equal(t, "legacy-pass", got)
	})
}
