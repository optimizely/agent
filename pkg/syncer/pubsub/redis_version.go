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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

const (
	// MinRedisVersionForStreams is the minimum Redis version that supports Streams
	MinRedisVersionForStreams = 5
	// RedisVersionCheckTimeout is the timeout for version detection
	RedisVersionCheckTimeout = 5 * time.Second
)

// detectRedisMajorVersion attempts to detect the major version of Redis
// Returns the major version number (e.g., 6 for Redis 6.2.5) or 0 on error
func detectRedisMajorVersion(client *redis.Client) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), RedisVersionCheckTimeout)
	defer cancel()

	// Try to get server info
	info, err := client.Info(ctx, "server").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to execute INFO server command: %w", err)
	}

	// Parse version from "redis_version:x.y.z"
	// Expected format: "redis_version:6.2.5" or "redis_version:7.0.0"
	for _, line := range strings.Split(info, "\r\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "redis_version:") {
			version := strings.TrimPrefix(line, "redis_version:")
			version = strings.TrimSpace(version)

			// Split on dots: "6.2.5" -> ["6", "2", "5"]
			parts := strings.Split(version, ".")
			if len(parts) < 1 {
				return 0, fmt.Errorf("invalid version format: %s", version)
			}

			// Parse major version (first part)
			major, err := strconv.Atoi(parts[0])
			if err != nil {
				return 0, fmt.Errorf("failed to parse major version from %s: %w", version, err)
			}

			if major < 2 {
				return 0, fmt.Errorf("unsupported Redis version: %d (minimum supported: 2)", major)
			}

			return major, nil
		}
	}

	return 0, fmt.Errorf("redis_version field not found in INFO server output")
}

// SupportsRedisStreams checks if the connected Redis instance supports Streams
// Returns true if Redis >= 5.0, false otherwise
// Falls back to false on any error (safe default)
func SupportsRedisStreams(client *redis.Client) bool {
	majorVersion, err := detectRedisMajorVersion(client)
	if err != nil {
		// Detection failed - log warning and return false (safe fallback)
		log.Warn().Err(err).Msg("Could not detect Redis version - will use Pub/Sub as safe fallback")
		return false
	}

	supported := majorVersion >= MinRedisVersionForStreams
	if supported {
		log.Info().
			Int("redis_version", majorVersion).
			Msg("Redis Streams supported - will use Streams for notifications")
	} else {
		log.Info().
			Int("redis_version", majorVersion).
			Int("min_required", MinRedisVersionForStreams).
			Msg("Redis Streams not supported - will use Pub/Sub for notifications")
	}

	return supported
}
