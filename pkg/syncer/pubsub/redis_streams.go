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
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"

	"github.com/optimizely/agent/pkg/metrics"
)

// RedisStreams implements persistent message delivery using Redis Streams
type RedisStreams struct {
	Host     string
	Password string
	Database int
	// Stream configuration
	MaxLen        int64
	ConsumerGroup string
	ConsumerName  string
	// Batching configuration
	BatchSize     int
	FlushInterval time.Duration
	// Retry configuration
	MaxRetries    int
	RetryDelay    time.Duration
	MaxRetryDelay time.Duration
	// Connection timeout
	ConnTimeout time.Duration
	// Metrics registry
	metricsRegistry *metrics.Registry
}

func (r *RedisStreams) Publish(ctx context.Context, channel string, message interface{}) error {
	streamName := r.getStreamName(channel)

	// Convert message to string for consistent handling
	var messageStr string
	switch v := message.(type) {
	case []byte:
		messageStr = string(v)
	case string:
		messageStr = v
	default:
		// For other types, marshal to JSON
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
		messageStr = string(jsonBytes)
	}

	// Add message to stream with automatic ID generation
	args := &redis.XAddArgs{
		Stream: streamName,
		Values: map[string]interface{}{
			"data":      messageStr,
			"timestamp": time.Now().Unix(),
		},
	}

	// Apply max length trimming if configured
	if r.MaxLen > 0 {
		args.MaxLen = r.MaxLen
		args.Approx = true // Use approximate trimming for better performance
	}

	return r.executeWithRetry(ctx, func(client *redis.Client) error {
		return client.XAdd(ctx, args).Err()
	})
}

func (r *RedisStreams) Subscribe(ctx context.Context, channel string) (chan string, error) {
	streamName := r.getStreamName(channel)
	consumerGroup := r.getConsumerGroup()
	consumerName := r.getConsumerName()

	ch := make(chan string)
	ready := make(chan error, 1) // Signal when consumer group is ready
	stop := make(chan struct{})   // Signal to stop goroutine

	go func() {
		defer close(ch)
		defer close(ready) // Ensure ready is closed

		batchSize := r.getBatchSize()
		flushTicker := time.NewTicker(r.getFlushInterval())
		defer flushTicker.Stop()

		var batch []string
		var client *redis.Client
		var lastReconnect time.Time
		reconnectDelay := 1 * time.Second
		maxReconnectDelay := 30 * time.Second

		// Initialize connection
		client = r.createClient()
		defer client.Close()

		// Create consumer group with retry
		if err := r.createConsumerGroupWithRetry(ctx, client, streamName, consumerGroup); err != nil {
			log.Error().Err(err).Str("stream", streamName).Str("group", consumerGroup).Msg("Failed to create consumer group")
			select {
			case ready <- err: // Signal initialization failure
			case <-stop: // Main function signaled stop
			}
			return
		}

		// Signal that consumer group is ready
		select {
		case ready <- nil:
		case <-stop: // Main function signaled stop
			return
		}

		for {
			select {
			case <-stop:
				// Main function requested stop
				return
			case <-ctx.Done():
				// Send any remaining batch before closing
				if len(batch) > 0 {
					r.sendBatch(ch, batch, ctx)
				}
				return
			case <-flushTicker.C:
				// Flush interval reached - send current batch
				if len(batch) > 0 {
					r.incrementCounter("batch.flush_interval")
					r.sendBatch(ch, batch, ctx)
					batch = nil
				}
			default:
				// Read messages from the stream using consumer group
				streams, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
					Group:    consumerGroup,
					Consumer: consumerName,
					Streams:  []string{streamName, ">"},
					Count:    int64(batchSize - len(batch)), // Read up to remaining batch size
					Block:    100 * time.Millisecond,        // Short block to allow flush checking
				}).Result()

				if err != nil {
					if err == redis.Nil {
						continue // No messages, continue polling
					}

					// Handle connection errors with exponential backoff reconnection
					if r.isConnectionError(err) {
						r.incrementCounter("connection.error")
						log.Warn().Err(err).Msg("Redis connection error, attempting reconnection")

						// Apply exponential backoff for reconnection
						if time.Since(lastReconnect) > reconnectDelay {
							r.incrementCounter("connection.reconnect_attempt")
							client.Close()
							client = r.createClient()
							lastReconnect = time.Now()

							// Recreate consumer group after reconnection
							if groupErr := r.createConsumerGroupWithRetry(ctx, client, streamName, consumerGroup); groupErr != nil {
								r.incrementCounter("connection.group_recreate_error")
								log.Error().Err(groupErr).Msg("Failed to recreate consumer group after reconnection")
							} else {
								r.incrementCounter("connection.reconnect_success")
							}

							// Increase reconnect delay with exponential backoff
							reconnectDelay = time.Duration(math.Min(float64(reconnectDelay*2), float64(maxReconnectDelay)))
						} else {
							// Wait before next retry
							time.Sleep(100 * time.Millisecond)
						}
					} else {
						// Log other errors but continue processing
						r.incrementCounter("read.error")
						log.Debug().Err(err).Msg("Redis streams read error")
					}
					continue
				}

				// Reset reconnect delay on successful read
				reconnectDelay = 1 * time.Second

				// Process messages from streams
				messageCount := 0
				for _, stream := range streams {
					for _, message := range stream.Messages {
						// Extract the data field
						if data, ok := message.Values["data"].(string); ok {
							batch = append(batch, data)
							messageCount++

							// Acknowledge the message with retry
							if ackErr := r.acknowledgeMessage(ctx, client, streamName, consumerGroup, message.ID); ackErr != nil {
								log.Warn().Err(ackErr).Str("messageID", message.ID).Msg("Failed to acknowledge message")
							}

							// Send batch if it's full
							if len(batch) >= batchSize {
								r.incrementCounter("batch.sent")
								r.sendBatch(ch, batch, ctx)
								batch = nil
								// Continue processing more messages
							}
						}
					}
				}

				// Track successful message reads
				if messageCount > 0 {
					r.incrementCounter("messages.read")
				}
			}
		}
	}()

	// Wait for consumer group initialization before returning
	select {
	case err := <-ready:
		if err != nil {
			close(stop) // Signal goroutine to stop on initialization error
			return nil, err
		}
		// Success - goroutine continues running
		return ch, nil
	case <-ctx.Done():
		close(stop) // Signal goroutine to stop on context cancellation
		return nil, ctx.Err()
	}
}

// Helper method to send batch to channel
func (r *RedisStreams) sendBatch(ch chan string, batch []string, ctx context.Context) {
	for _, msg := range batch {
		select {
		case ch <- msg:
			// Message sent successfully
		case <-ctx.Done():
			return
		}
	}
}

// Helper methods
func (r *RedisStreams) getStreamName(channel string) string {
	return fmt.Sprintf("stream:%s", channel)
}

func (r *RedisStreams) getConsumerGroup() string {
	if r.ConsumerGroup == "" {
		return "notifications"
	}
	return r.ConsumerGroup
}

func (r *RedisStreams) getConsumerName() string {
	if r.ConsumerName == "" {
		hostname, _ := os.Hostname()
		pid := os.Getpid()
		return fmt.Sprintf("consumer-%s-%d-%d", hostname, pid, time.Now().Unix())
	}
	return r.ConsumerName
}

func (r *RedisStreams) getBatchSize() int {
	if r.BatchSize <= 0 {
		return 10 // Default batch size
	}
	return r.BatchSize
}

func (r *RedisStreams) getFlushInterval() time.Duration {
	if r.FlushInterval <= 0 {
		return 5 * time.Second // Default flush interval
	}
	return r.FlushInterval
}

func (r *RedisStreams) getMaxRetries() int {
	if r.MaxRetries <= 0 {
		return 3 // Default max retries
	}
	return r.MaxRetries
}

func (r *RedisStreams) getRetryDelay() time.Duration {
	if r.RetryDelay <= 0 {
		return 100 * time.Millisecond // Default retry delay
	}
	return r.RetryDelay
}

func (r *RedisStreams) getMaxRetryDelay() time.Duration {
	if r.MaxRetryDelay <= 0 {
		return 5 * time.Second // Default max retry delay
	}
	return r.MaxRetryDelay
}

func (r *RedisStreams) getConnTimeout() time.Duration {
	if r.ConnTimeout <= 0 {
		return 10 * time.Second // Default connection timeout
	}
	return r.ConnTimeout
}

// createClient creates a new Redis client with configured timeouts
func (r *RedisStreams) createClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         r.Host,
		Password:     r.Password,
		DB:           r.Database,
		DialTimeout:  r.getConnTimeout(),
		ReadTimeout:  r.getConnTimeout(),
		WriteTimeout: r.getConnTimeout(),
		PoolTimeout:  r.getConnTimeout(),
	})
}

// executeWithRetry executes a Redis operation with retry logic
func (r *RedisStreams) executeWithRetry(ctx context.Context, operation func(client *redis.Client) error) error {
	start := time.Now()
	maxRetries := r.getMaxRetries()
	retryDelay := r.getRetryDelay()
	maxRetryDelay := r.getMaxRetryDelay()

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		client := r.createClient()
		var err error
		func() {
			defer client.Close() // Always executes, even on panic
			err = operation(client)
		}()

		if err == nil {
			// Record successful operation metrics
			r.incrementCounter("operations.success")
			r.recordTimer("operations.duration", time.Since(start).Seconds())
			if attempt > 0 {
				r.incrementCounter("retries.success")
			}
			return nil // Success
		}

		lastErr = err
		r.incrementCounter("operations.error")

		// Don't retry on non-recoverable errors
		if !r.isRetryableError(err) {
			r.incrementCounter("errors.non_retryable")
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// Don't sleep after the last attempt
		if attempt < maxRetries {
			r.incrementCounter("retries.attempt")
			// Calculate delay with exponential backoff
			delay := time.Duration(math.Min(float64(retryDelay)*math.Pow(2, float64(attempt)), float64(maxRetryDelay)))

			select {
			case <-ctx.Done():
				r.incrementCounter("operations.canceled")
				return ctx.Err()
			case <-time.After(delay):
				// Continue to next retry
			}
		}
	}

	r.incrementCounter("retries.exhausted")
	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}

// createConsumerGroupWithRetry creates a consumer group with retry logic
func (r *RedisStreams) createConsumerGroupWithRetry(ctx context.Context, client *redis.Client, streamName, consumerGroup string) error {
	maxRetries := r.getMaxRetries()
	retryDelay := r.getRetryDelay()
	maxRetryDelay := r.getMaxRetryDelay()

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		_, err := client.XGroupCreateMkStream(ctx, streamName, consumerGroup, "$").Result()
		if err == nil || err.Error() == "BUSYGROUP Consumer Group name already exists" {
			return nil // Success
		}

		lastErr = err

		// Don't retry on non-recoverable errors
		if !r.isRetryableError(err) {
			return fmt.Errorf("non-retryable error creating consumer group: %w", err)
		}

		// Don't sleep after the last attempt
		if attempt < maxRetries {
			// Calculate delay with exponential backoff
			delay := time.Duration(math.Min(float64(retryDelay)*math.Pow(2, float64(attempt)), float64(maxRetryDelay)))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue to next retry
			}
		}
	}

	return fmt.Errorf("failed to create consumer group after %d retries: %w", maxRetries, lastErr)
}

// acknowledgeMessage acknowledges a message with retry logic
func (r *RedisStreams) acknowledgeMessage(ctx context.Context, client *redis.Client, streamName, consumerGroup, messageID string) error {
	maxRetries := 2 // Fewer retries for ACK operations
	retryDelay := 50 * time.Millisecond

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := client.XAck(ctx, streamName, consumerGroup, messageID).Err()
		if err == nil {
			r.incrementCounter("ack.success")
			if attempt > 0 {
				r.incrementCounter("ack.retry_success")
			}
			return nil // Success
		}

		lastErr = err
		r.incrementCounter("ack.error")

		// Don't retry on non-recoverable errors
		if !r.isRetryableError(err) {
			r.incrementCounter("ack.non_retryable_error")
			return fmt.Errorf("non-retryable ACK error: %w", err)
		}

		// Don't sleep after the last attempt
		if attempt < maxRetries {
			r.incrementCounter("ack.retry_attempt")
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
				// Continue to next retry
			}
		}
	}

	r.incrementCounter("ack.retry_exhausted")
	return fmt.Errorf("ACK failed after %d retries: %w", maxRetries, lastErr)
}

// isRetryableError determines if an error is retryable
func (r *RedisStreams) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Network/connection errors that are retryable
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"network is unreachable",
		"broken pipe",
		"eof",
		"i/o timeout",
		"connection pool exhausted",
		"context deadline exceeded",
		"context canceled", // Handle graceful shutdowns
		"no such host",     // DNS lookup failures
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(strings.ToLower(errStr), retryable) {
			return true
		}
	}

	// Redis-specific retryable errors
	if strings.Contains(errStr, "LOADING") || // Redis is loading data
		strings.Contains(errStr, "READONLY") || // Redis is in read-only mode
		strings.Contains(errStr, "CLUSTERDOWN") { // Redis cluster is down
		return true
	}

	return false
}

// isConnectionError determines if an error is a connection error
func (r *RedisStreams) isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"network is unreachable",
		"broken pipe",
		"eof",
		"connection pool exhausted",
	}

	for _, connErr := range connectionErrors {
		if strings.Contains(strings.ToLower(errStr), connErr) {
			return true
		}
	}

	return false
}

// SetMetricsRegistry sets the metrics registry for tracking statistics
func (r *RedisStreams) SetMetricsRegistry(registry *metrics.Registry) {
	r.metricsRegistry = registry
}

// incrementCounter safely increments a metrics counter if registry is available
func (r *RedisStreams) incrementCounter(key string) {
	if r.metricsRegistry != nil {
		if counter := r.metricsRegistry.GetCounter("redis_streams." + key); counter != nil {
			counter.Add(1)
		}
	}
}

// recordTimer safely records a timer metric if registry is available
func (r *RedisStreams) recordTimer(key string, duration float64) {
	if r.metricsRegistry != nil {
		if timer := r.metricsRegistry.NewTimer("redis_streams." + key); timer != nil {
			timer.Update(duration)
		}
	}
}
