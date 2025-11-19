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

// Package pubsub provides pubsub functionality for the agent syncer
package pubsub

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
)

func TestDetectRedisMajorVersion(t *testing.T) {
	tests := []struct {
		name        string
		infoOutput  string
		infoError   error
		wantVersion int
		wantErr     bool
	}{
		{
			name:        "Redis 6.2.5",
			infoOutput:  "# Server\r\nredis_version:6.2.5\r\nredis_git_sha1:00000000\r\nredis_git_dirty:0\r\nredis_build_id:1234567890\r\nredis_mode:standalone\r\nos:Linux 5.10.0-8-amd64 x86_64\r\narch_bits:64",
			wantVersion: 6,
			wantErr:     false,
		},
		{
			name:        "Redis 7.0.0",
			infoOutput:  "# Server\r\nredis_version:7.0.0\r\nredis_git_sha1:00000000",
			wantVersion: 7,
			wantErr:     false,
		},
		{
			name:        "Redis 5.0.14",
			infoOutput:  "# Server\r\nredis_version:5.0.14\r\nredis_git_sha1:00000000",
			wantVersion: 5,
			wantErr:     false,
		},
		{
			name:        "Redis 4.0.14",
			infoOutput:  "# Server\r\nredis_version:4.0.14\r\nredis_git_sha1:00000000",
			wantVersion: 4,
			wantErr:     false,
		},
		{
			name:        "Redis 3.2.12",
			infoOutput:  "# Server\r\nredis_version:3.2.12\r\nredis_git_sha1:00000000",
			wantVersion: 3,
			wantErr:     false,
		},
		{
			name:        "Redis 2.8.24",
			infoOutput:  "# Server\r\nredis_version:2.8.24\r\nredis_git_sha1:00000000",
			wantVersion: 2,
			wantErr:     false,
		},
		{
			name:        "Version with extra whitespace",
			infoOutput:  "# Server\r\nredis_version:  6.2.5\r\nredis_git_sha1:00000000",
			wantVersion: 6,
			wantErr:     false,
		},
		{
			name:        "Version field missing",
			infoOutput:  "# Server\r\nredis_git_sha1:00000000\r\nredis_git_dirty:0",
			wantVersion: 0,
			wantErr:     true,
		},
		{
			name:        "Invalid version format (no dots)",
			infoOutput:  "# Server\r\nredis_version:6\r\nredis_git_sha1:00000000",
			wantVersion: 6,
			wantErr:     false,
		},
		{
			name:        "Invalid version format (non-numeric)",
			infoOutput:  "# Server\r\nredis_version:abc.def.ghi\r\nredis_git_sha1:00000000",
			wantVersion: 0,
			wantErr:     true,
		},
		{
			name:        "Unsupported Redis 1.x",
			infoOutput:  "# Server\r\nredis_version:1.2.6\r\nredis_git_sha1:00000000",
			wantVersion: 0,
			wantErr:     true,
		},
		{
			name:        "INFO command blocked (NOPERM)",
			infoOutput:  "",
			infoError:   redis.NewCmd(context.Background()).Err(),
			wantVersion: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Redis client
			db, mock := redismock.NewClientMock()
			defer db.Close()

			// Setup mock expectation
			if tt.infoError != nil {
				mock.ExpectInfo("server").SetErr(tt.infoError)
			} else {
				mock.ExpectInfo("server").SetVal(tt.infoOutput)
			}

			// Run test
			got, err := detectRedisMajorVersion(db)

			// Check expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Redis mock expectations not met: %v", err)
			}

			// Verify results
			if (err != nil) != tt.wantErr {
				t.Errorf("detectRedisMajorVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantVersion {
				t.Errorf("detectRedisMajorVersion() = %v, want %v", got, tt.wantVersion)
			}
		})
	}
}

func TestSupportsRedisStreams(t *testing.T) {
	tests := []struct {
		name        string
		infoOutput  string
		infoError   error
		wantSupport bool
	}{
		{
			name:        "Redis 7.x supports Streams",
			infoOutput:  "# Server\r\nredis_version:7.0.0\r\nredis_git_sha1:00000000",
			wantSupport: true,
		},
		{
			name:        "Redis 6.x supports Streams",
			infoOutput:  "# Server\r\nredis_version:6.2.5\r\nredis_git_sha1:00000000",
			wantSupport: true,
		},
		{
			name:        "Redis 5.x supports Streams",
			infoOutput:  "# Server\r\nredis_version:5.0.14\r\nredis_git_sha1:00000000",
			wantSupport: true,
		},
		{
			name:        "Redis 4.x does not support Streams",
			infoOutput:  "# Server\r\nredis_version:4.0.14\r\nredis_git_sha1:00000000",
			wantSupport: false,
		},
		{
			name:        "Redis 3.x does not support Streams",
			infoOutput:  "# Server\r\nredis_version:3.2.12\r\nredis_git_sha1:00000000",
			wantSupport: false,
		},
		{
			name:        "Redis 2.x does not support Streams",
			infoOutput:  "# Server\r\nredis_version:2.8.24\r\nredis_git_sha1:00000000",
			wantSupport: false,
		},
		{
			name:        "Detection failure falls back to false (safe default)",
			infoOutput:  "",
			infoError:   redis.NewCmd(context.Background()).Err(),
			wantSupport: false,
		},
		{
			name:        "Invalid version falls back to false",
			infoOutput:  "# Server\r\nredis_version:invalid\r\nredis_git_sha1:00000000",
			wantSupport: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Redis client
			db, mock := redismock.NewClientMock()
			defer db.Close()

			// Setup mock expectation
			if tt.infoError != nil {
				mock.ExpectInfo("server").SetErr(tt.infoError)
			} else {
				mock.ExpectInfo("server").SetVal(tt.infoOutput)
			}

			// Run test
			got := SupportsRedisStreams(db)

			// Check expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Redis mock expectations not met: %v", err)
			}

			// Verify results
			if got != tt.wantSupport {
				t.Errorf("SupportsRedisStreams() = %v, want %v", got, tt.wantSupport)
			}
		})
	}
}

// TestSupportsRedisStreams_NoPanic ensures the function never panics
func TestSupportsRedisStreams_NoPanic(t *testing.T) {
	// Test various error conditions to ensure no panic
	testCases := []struct {
		name       string
		infoOutput string
		wantResult bool
	}{
		{
			name:       "Empty response",
			infoOutput: "",
			wantResult: false,
		},
		{
			name:       "Invalid format",
			infoOutput: "invalid data",
			wantResult: false,
		},
		{
			name:       "Empty version",
			infoOutput: "# Server\r\nredis_version:\r\nredis_git_sha1:00000000",
			wantResult: false,
		},
		{
			name:       "Version with just newline",
			infoOutput: "# Server\r\nredis_version:\r\nredis_git_sha1:00000000",
			wantResult: false,
		},
		{
			name:       "Invalid dots",
			infoOutput: "# Server\r\nredis_version:.....\r\nredis_git_sha1:00000000",
			wantResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := redismock.NewClientMock()
			defer db.Close()

			mock.ExpectInfo("server").SetVal(tc.infoOutput)

			// This should not panic
			result := SupportsRedisStreams(db)

			// Verify expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Redis mock expectations not met: %v", err)
			}

			// Should always fall back to false on any error
			if result != tc.wantResult {
				t.Errorf("Expected %v for malformed input, got %v", tc.wantResult, result)
			}
		})
	}
}
