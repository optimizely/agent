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
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestUnmarshalWithPasswordExtraction(t *testing.T) {
	// Test struct to unmarshal into
	type TestRedisConfig struct {
		Host     string `json:"host"`
		Password string `json:"password"`
		Database int    `json:"database"`
	}

	tests := []struct {
		name         string
		jsonData     string
		envVar       string
		envValue     string
		wantPassword string
		wantErr      bool
		wantHost     string
		wantDatabase int
	}{
		{
			name:         "unmarshal with auth_token field",
			jsonData:     `{"host":"localhost:6379","auth_token":"token123","database":0}`,
			envVar:       "TEST_REDIS_PASSWORD",
			wantPassword: "token123",
			wantErr:      false,
			wantHost:     "localhost:6379",
			wantDatabase: 0,
		},
		{
			name:         "unmarshal with redis_secret field",
			jsonData:     `{"host":"localhost:6379","redis_secret":"secret456","database":1}`,
			envVar:       "TEST_REDIS_PASSWORD",
			wantPassword: "secret456",
			wantErr:      false,
			wantHost:     "localhost:6379",
			wantDatabase: 1,
		},
		{
			name:         "unmarshal with password field",
			jsonData:     `{"host":"localhost:6379","password":"pass789","database":2}`,
			envVar:       "TEST_REDIS_PASSWORD",
			wantPassword: "pass789",
			wantErr:      false,
			wantHost:     "localhost:6379",
			wantDatabase: 2,
		},
		{
			name:         "unmarshal with env var fallback",
			jsonData:     `{"host":"localhost:6379","database":0}`,
			envVar:       "TEST_REDIS_PASSWORD",
			envValue:     "env_password",
			wantPassword: "env_password",
			wantErr:      false,
			wantHost:     "localhost:6379",
			wantDatabase: 0,
		},
		{
			name:         "unmarshal with no password configured",
			jsonData:     `{"host":"localhost:6379","database":0}`,
			envVar:       "TEST_REDIS_PASSWORD",
			wantPassword: "",
			wantErr:      false,
			wantHost:     "localhost:6379",
			wantDatabase: 0,
		},
		{
			name:         "priority: auth_token over redis_secret",
			jsonData:     `{"host":"localhost:6379","auth_token":"token","redis_secret":"secret","password":"pass","database":0}`,
			envVar:       "TEST_REDIS_PASSWORD",
			wantPassword: "token",
			wantErr:      false,
			wantHost:     "localhost:6379",
			wantDatabase: 0,
		},
		{
			name:         "invalid JSON returns error",
			jsonData:     `{"host":"localhost:6379","invalid json`,
			envVar:       "TEST_REDIS_PASSWORD",
			wantPassword: "",
			wantErr:      true,
		},
		{
			name:         "empty JSON object",
			jsonData:     `{}`,
			envVar:       "TEST_REDIS_PASSWORD",
			wantPassword: "",
			wantErr:      false,
			wantHost:     "",
			wantDatabase: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if needed
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			} else {
				os.Unsetenv(tt.envVar)
			}

			// Create target struct
			var config TestRedisConfig

			// Call the function
			password, err := UnmarshalWithPasswordExtraction([]byte(tt.jsonData), &config, tt.envVar)

			// Check error
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify password extraction
			assert.Equal(t, tt.wantPassword, password)

			// Verify other fields were unmarshaled correctly
			assert.Equal(t, tt.wantHost, config.Host)
			assert.Equal(t, tt.wantDatabase, config.Database)
		})
	}
}

func TestUnmarshalWithPasswordExtraction_AliasPattern(t *testing.T) {
	// This test verifies that the alias pattern works correctly
	// and prevents infinite recursion

	type RedisConfig struct {
		Host     string `json:"host"`
		Password string `json:"-"` // Not unmarshaled from JSON
		Database int    `json:"database"`
	}

	// Simulate the alias pattern used in actual implementations
	jsonData := `{"host":"localhost:6379","auth_token":"secret123","database":1}`

	var config RedisConfig
	type Alias RedisConfig
	alias := (*Alias)(&config)

	// This should not cause infinite recursion
	password, err := UnmarshalWithPasswordExtraction([]byte(jsonData), alias, "TEST_PASSWORD")

	require.NoError(t, err)
	assert.Equal(t, "secret123", password)
	assert.Equal(t, "localhost:6379", config.Host)
	assert.Equal(t, 1, config.Database)
}

func TestUnmarshalWithPasswordExtraction_StructWithNestedFields(t *testing.T) {
	// Test with a more complex struct to ensure general unmarshaling works

	type ComplexRedisConfig struct {
		Host     string         `json:"host"`
		Password string         `json:"password"`
		Database int            `json:"database"`
		Options  map[string]int `json:"options"`
	}

	jsonData := `{
		"host":"redis.example.com:6379",
		"auth_token":"complex_token",
		"database":5,
		"options":{"maxRetries":3,"timeout":30}
	}`

	var config ComplexRedisConfig
	password, err := UnmarshalWithPasswordExtraction([]byte(jsonData), &config, "")

	require.NoError(t, err)
	assert.Equal(t, "complex_token", password)
	assert.Equal(t, "redis.example.com:6379", config.Host)
	assert.Equal(t, 5, config.Database)
	assert.Equal(t, 3, config.Options["maxRetries"])
	assert.Equal(t, 30, config.Options["timeout"])
}

func TestUnmarshalWithPasswordExtraction_MalformedJSON(t *testing.T) {
	type RedisConfig struct {
		Host     string `json:"host"`
		Database int    `json:"database"`
	}

	tests := []struct {
		name     string
		jsonData string
	}{
		{
			name:     "incomplete JSON",
			jsonData: `{"host":"localhost:6379"`,
		},
		{
			name:     "invalid syntax",
			jsonData: `{host:localhost}`,
		},
		{
			name:     "trailing comma",
			jsonData: `{"host":"localhost:6379",}`,
		},
		{
			name:     "empty string",
			jsonData: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config RedisConfig
			_, err := UnmarshalWithPasswordExtraction([]byte(tt.jsonData), &config, "TEST_PASSWORD")
			assert.Error(t, err, "Expected error for malformed JSON")

			// Verify it's a JSON unmarshal error
			var jsonErr *json.SyntaxError
			var jsonTypeErr *json.UnmarshalTypeError
			isJSONError := assert.ErrorAs(t, err, &jsonErr) || assert.ErrorAs(t, err, &jsonTypeErr) || err.Error() == "unexpected end of JSON input"
			assert.True(t, isJSONError, "Error should be a JSON unmarshal error")
		})
	}
}
