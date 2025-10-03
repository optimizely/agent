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

package pubsub

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/optimizely/agent/pkg/metrics"
)

func setupRedisStreamsWithRetry() *RedisStreams {
	return &RedisStreams{
		Host:          "localhost:6379",
		Password:      "",
		Database:      0,
		MaxLen:        1000,
		ConsumerGroup: "test-group",
		ConsumerName:  "test-consumer",
		BatchSize:     10,
		FlushInterval: 5 * time.Second,
		MaxRetries:    3,
		RetryDelay:    50 * time.Millisecond,
		MaxRetryDelay: 1 * time.Second,
		ConnTimeout:   5 * time.Second,
		// Don't set metricsRegistry by default to avoid conflicts
		metricsRegistry: nil,
	}
}

func TestRedisStreams_RetryConfiguration_Defaults(t *testing.T) {
	rs := &RedisStreams{}

	assert.Equal(t, 3, rs.getMaxRetries())
	assert.Equal(t, 100*time.Millisecond, rs.getRetryDelay())
	assert.Equal(t, 5*time.Second, rs.getMaxRetryDelay())
	assert.Equal(t, 10*time.Second, rs.getConnTimeout())
}

func TestRedisStreams_RetryConfiguration_Custom(t *testing.T) {
	rs := &RedisStreams{
		MaxRetries:    5,
		RetryDelay:    200 * time.Millisecond,
		MaxRetryDelay: 10 * time.Second,
		ConnTimeout:   30 * time.Second,
	}

	assert.Equal(t, 5, rs.getMaxRetries())
	assert.Equal(t, 200*time.Millisecond, rs.getRetryDelay())
	assert.Equal(t, 10*time.Second, rs.getMaxRetryDelay())
	assert.Equal(t, 30*time.Second, rs.getConnTimeout())
}

func TestRedisStreams_IsRetryableError(t *testing.T) {
	rs := setupRedisStreamsWithRetry()

	testCases := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "nil error",
			err:       nil,
			retryable: false,
		},
		{
			name:      "connection refused",
			err:       errors.New("connection refused"),
			retryable: true,
		},
		{
			name:      "connection reset",
			err:       errors.New("connection reset by peer"),
			retryable: true,
		},
		{
			name:      "timeout error",
			err:       errors.New("i/o timeout"),
			retryable: true,
		},
		{
			name:      "network unreachable",
			err:       errors.New("network is unreachable"),
			retryable: true,
		},
		{
			name:      "broken pipe",
			err:       errors.New("broken pipe"),
			retryable: true,
		},
		{
			name:      "EOF error",
			err:       errors.New("EOF"),
			retryable: true,
		},
		{
			name:      "context deadline exceeded",
			err:       errors.New("context deadline exceeded"),
			retryable: true,
		},
		{
			name:      "context canceled",
			err:       errors.New("context canceled"),
			retryable: true,
		},
		{
			name:      "Redis LOADING",
			err:       errors.New("LOADING Redis is loading the dataset in memory"),
			retryable: true,
		},
		{
			name:      "Redis READONLY",
			err:       errors.New("READONLY You can't write against a read only replica."),
			retryable: true,
		},
		{
			name:      "Redis CLUSTERDOWN",
			err:       errors.New("CLUSTERDOWN Hash slot not served"),
			retryable: true,
		},
		{
			name:      "syntax error - not retryable",
			err:       errors.New("ERR syntax error"),
			retryable: false,
		},
		{
			name:      "wrong type error - not retryable",
			err:       errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"),
			retryable: false,
		},
		{
			name:      "authentication error - not retryable",
			err:       errors.New("NOAUTH Authentication required"),
			retryable: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := rs.isRetryableError(tc.err)
			assert.Equal(t, tc.retryable, result, "Error: %v", tc.err)
		})
	}
}

func TestRedisStreams_IsConnectionError(t *testing.T) {
	rs := setupRedisStreamsWithRetry()

	testCases := []struct {
		name         string
		err          error
		isConnection bool
	}{
		{
			name:         "nil error",
			err:          nil,
			isConnection: false,
		},
		{
			name:         "connection refused",
			err:          errors.New("connection refused"),
			isConnection: true,
		},
		{
			name:         "connection reset",
			err:          errors.New("connection reset by peer"),
			isConnection: true,
		},
		{
			name:         "network unreachable",
			err:          errors.New("network is unreachable"),
			isConnection: true,
		},
		{
			name:         "broken pipe",
			err:          errors.New("broken pipe"),
			isConnection: true,
		},
		{
			name:         "EOF error",
			err:          errors.New("EOF"),
			isConnection: true,
		},
		{
			name:         "syntax error - not connection",
			err:          errors.New("ERR syntax error"),
			isConnection: false,
		},
		{
			name:         "timeout - not connection",
			err:          errors.New("i/o timeout"),
			isConnection: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := rs.isConnectionError(tc.err)
			assert.Equal(t, tc.isConnection, result, "Error: %v", tc.err)
		})
	}
}

func TestRedisStreams_Publish_WithInvalidHost_ShouldRetry(t *testing.T) {
	rs := setupRedisStreamsWithRetry()
	rs.Host = "invalid-host:6379" // Use invalid host to trigger connection errors
	rs.MaxRetries = 2             // Limit retries for faster test
	rs.RetryDelay = 10 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := rs.Publish(ctx, "test-channel", "test message")

	// Should fail with either retry exhaustion or non-retryable error (DNS lookup can fail differently in CI)
	assert.Error(t, err)
	errMsg := err.Error()
	assert.True(t,
		strings.Contains(errMsg, "operation failed after") ||
			strings.Contains(errMsg, "non-retryable error") ||
			strings.Contains(errMsg, "lookup invalid-host"),
		"Expected retry or DNS error, got: %s", errMsg)
}

func TestRedisStreams_Publish_WithCanceledContext(t *testing.T) {
	rs := setupRedisStreamsWithRetry()
	rs.Host = "invalid-host:6379" // Use invalid host to trigger retries
	rs.MaxRetries = 5
	rs.RetryDelay = 100 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel context immediately to test cancellation handling
	cancel()

	err := rs.Publish(ctx, "test-channel", "test message")

	// Should fail with context canceled error
	assert.Error(t, err)
	// Could be either context canceled directly or wrapped in retry error
	assert.True(t, strings.Contains(err.Error(), "context canceled") ||
		strings.Contains(err.Error(), "operation failed after"))
}

func TestRedisStreams_MetricsIntegration(t *testing.T) {
	rs := setupRedisStreamsWithRetry()

	// Test that metrics registry can be set and retrieved
	registry := metrics.NewRegistry("metrics_integration_test")
	rs.SetMetricsRegistry(registry)

	assert.Equal(t, registry, rs.metricsRegistry)
}

func TestRedisStreams_MetricsTracking_SafeWithNilRegistry(t *testing.T) {
	rs := setupRedisStreamsWithRetry()
	rs.metricsRegistry = nil

	// These should not panic with nil registry
	rs.incrementCounter("test.counter")
	rs.recordTimer("test.timer", 1.5)
}

func TestRedisStreams_CreateClient_WithTimeouts(t *testing.T) {
	rs := setupRedisStreamsWithRetry()
	rs.ConnTimeout = 2 * time.Second

	client := rs.createClient()
	defer client.Close()

	assert.NotNil(t, client)
	// Note: go-redis client options are not easily inspectable,
	// but we can verify the client was created without error
}

func TestRedisStreams_AcknowledgeMessage_WithRetry(t *testing.T) {
	// This test requires a running Redis instance
	rs := setupRedisStreamsWithRetry()
	ctx := context.Background()

	// Create a client to set up test data
	client := redis.NewClient(&redis.Options{
		Addr:     rs.Host,
		Password: rs.Password,
		DB:       rs.Database,
	})
	defer client.Close()

	streamName := "test-ack-stream"
	consumerGroup := "test-ack-group"

	// Clean up
	defer func() {
		client.Del(ctx, streamName)
	}()

	// Add a message to the stream
	msgID, err := client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: map[string]interface{}{
			"data": "test message",
		},
	}).Result()
	require.NoError(t, err)

	// Create consumer group
	client.XGroupCreateMkStream(ctx, streamName, consumerGroup, "0")

	// Test acknowledge with valid message ID (should succeed)
	err = rs.acknowledgeMessage(ctx, client, streamName, consumerGroup, msgID)
	assert.NoError(t, err)

	// Test acknowledge with invalid message ID (should fail but not crash)
	err = rs.acknowledgeMessage(ctx, client, streamName, consumerGroup, "invalid-id")
	assert.Error(t, err)
}

func TestRedisStreams_ExecuteWithRetry_NonRetryableError(t *testing.T) {
	rs := setupRedisStreamsWithRetry()
	ctx := context.Background()

	// Simulate a non-retryable error
	operation := func(client *redis.Client) error {
		return errors.New("ERR syntax error") // Non-retryable
	}

	err := rs.executeWithRetry(ctx, operation)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-retryable error")
	assert.Contains(t, err.Error(), "ERR syntax error")
}

func TestRedisStreams_ExecuteWithRetry_SuccessAfterRetries(t *testing.T) {
	rs := setupRedisStreamsWithRetry()
	rs.RetryDelay = 1 * time.Millisecond // Fast retries for testing
	// Don't set metrics registry to avoid expvar name conflicts across tests
	// (expvar counters are global and can't be reused even with unique registry names)
	ctx := context.Background()

	attemptCount := 0
	operation := func(client *redis.Client) error {
		attemptCount++
		if attemptCount < 3 {
			return errors.New("connection refused") // Retryable
		}
		return nil // Success on third attempt
	}

	err := rs.executeWithRetry(ctx, operation)

	assert.NoError(t, err)
	assert.Equal(t, 3, attemptCount)
}

func TestRedisStreams_ExecuteWithRetry_ExhaustRetries(t *testing.T) {
	rs := setupRedisStreamsWithRetry()
	rs.MaxRetries = 2
	rs.RetryDelay = 1 * time.Millisecond // Fast retries for testing
	// Don't set metrics registry to avoid expvar name conflicts across tests
	// (expvar counters are global and can't be reused even with unique registry names)
	ctx := context.Background()

	attemptCount := 0
	operation := func(client *redis.Client) error {
		attemptCount++
		return errors.New("connection refused") // Always retryable error
	}

	err := rs.executeWithRetry(ctx, operation)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation failed after 2 retries")
	assert.Equal(t, 3, attemptCount) // 1 initial + 2 retries
}

func TestRedisStreams_CreateConsumerGroupWithRetry_BusyGroupExists(t *testing.T) {
	rs := setupRedisStreamsWithRetry()
	ctx := context.Background()

	// Create a client to set up test data
	client := redis.NewClient(&redis.Options{
		Addr:     rs.Host,
		Password: rs.Password,
		DB:       rs.Database,
	})
	defer client.Close()

	streamName := "test-busy-group-stream"
	consumerGroup := "test-busy-group"

	// Clean up
	defer func() {
		client.Del(ctx, streamName)
	}()

	// First call should succeed
	err := rs.createConsumerGroupWithRetry(ctx, client, streamName, consumerGroup)
	assert.NoError(t, err)

	// Second call should also succeed (BUSYGROUP error is handled)
	err = rs.createConsumerGroupWithRetry(ctx, client, streamName, consumerGroup)
	assert.NoError(t, err)
}

func TestRedisStreams_ErrorHandling_ContextCancellation(t *testing.T) {
	rs := setupRedisStreamsWithRetry()
	rs.RetryDelay = 100 * time.Millisecond
	// Don't set metrics registry to avoid expvar name conflicts across tests
	// (expvar counters are global and can't be reused even with unique registry names)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		// Cancel context after a short delay
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	operation := func(client *redis.Client) error {
		return errors.New("connection refused") // Retryable error
	}

	err := rs.executeWithRetry(ctx, operation)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestRedisStreams_Subscribe_ErrorRecovery_Integration(t *testing.T) {
	// Integration test - requires Redis to be running
	rs := setupRedisStreamsWithRetry()
	rs.MaxRetries = 1 // Limit retries for faster test

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	channel := "test-error-recovery"
	defer cleanupRedisStream(rs.getStreamName(channel))

	// Start subscriber
	ch, err := rs.Subscribe(ctx, channel)
	require.NoError(t, err)

	// Give some time for setup
	time.Sleep(100 * time.Millisecond)

	// Publish a message
	err = rs.Publish(ctx, channel, "test message")
	require.NoError(t, err)

	// Should receive the message despite any internal error recovery
	select {
	case received := <-ch:
		assert.Equal(t, "test message", received)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}
