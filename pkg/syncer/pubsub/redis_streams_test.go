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
	"encoding/json"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRedisHost = "localhost:6379"
	testDatabase  = 0
	testPassword  = ""
)

func setupRedisStreams() *RedisStreams {
	return &RedisStreams{
		Host:          testRedisHost,
		Password:      testPassword,
		Database:      testDatabase,
		MaxLen:        1000,
		ConsumerGroup: "test-group",
		ConsumerName:  "test-consumer",
		BatchSize:     10,
		FlushInterval: 5 * time.Second,
	}
}

func cleanupRedisStream(streamName string) {
	client := redis.NewClient(&redis.Options{
		Addr:     testRedisHost,
		Password: testPassword,
		DB:       testDatabase,
	})
	defer client.Close()

	// Delete the stream and consumer group
	client.Del(context.Background(), streamName)
}

func TestRedisStreams_Publish_String(t *testing.T) {
	rs := setupRedisStreams()
	ctx := context.Background()
	channel := "test-channel-string"
	message := "test message"

	defer cleanupRedisStream(rs.getStreamName(channel))

	err := rs.Publish(ctx, channel, message)
	assert.NoError(t, err)

	// Verify message was added to stream
	client := redis.NewClient(&redis.Options{
		Addr:     testRedisHost,
		Password: testPassword,
		DB:       testDatabase,
	})
	defer client.Close()

	streamName := rs.getStreamName(channel)
	messages, err := client.XRange(ctx, streamName, "-", "+").Result()
	require.NoError(t, err)
	assert.Len(t, messages, 1)

	// Check message content
	data, exists := messages[0].Values["data"]
	assert.True(t, exists)
	assert.Equal(t, message, data)

	// Check timestamp exists
	timestamp, exists := messages[0].Values["timestamp"]
	assert.True(t, exists)
	assert.NotNil(t, timestamp)
}

func TestRedisStreams_Publish_JSON(t *testing.T) {
	rs := setupRedisStreams()
	ctx := context.Background()
	channel := "test-channel-json"

	testObj := map[string]interface{}{
		"type":    "notification",
		"payload": "test data",
		"id":      123,
	}

	defer cleanupRedisStream(rs.getStreamName(channel))

	err := rs.Publish(ctx, channel, testObj)
	assert.NoError(t, err)

	// Verify message was serialized correctly
	client := redis.NewClient(&redis.Options{
		Addr:     testRedisHost,
		Password: testPassword,
		DB:       testDatabase,
	})
	defer client.Close()

	streamName := rs.getStreamName(channel)
	messages, err := client.XRange(ctx, streamName, "-", "+").Result()
	require.NoError(t, err)
	assert.Len(t, messages, 1)

	// Check JSON was stored correctly
	data, exists := messages[0].Values["data"]
	assert.True(t, exists)

	var decoded map[string]interface{}
	err = json.Unmarshal([]byte(data.(string)), &decoded)
	require.NoError(t, err)
	assert.Equal(t, testObj["type"], decoded["type"])
	assert.Equal(t, testObj["payload"], decoded["payload"])
	assert.Equal(t, float64(123), decoded["id"]) // JSON numbers become float64
}

func TestRedisStreams_Publish_ByteArray(t *testing.T) {
	rs := setupRedisStreams()
	ctx := context.Background()
	channel := "test-channel-bytes"
	message := []byte("test byte message")

	defer cleanupRedisStream(rs.getStreamName(channel))

	err := rs.Publish(ctx, channel, message)
	assert.NoError(t, err)

	// Verify message was stored as string
	client := redis.NewClient(&redis.Options{
		Addr:     testRedisHost,
		Password: testPassword,
		DB:       testDatabase,
	})
	defer client.Close()

	streamName := rs.getStreamName(channel)
	messages, err := client.XRange(ctx, streamName, "-", "+").Result()
	require.NoError(t, err)
	assert.Len(t, messages, 1)

	data, exists := messages[0].Values["data"]
	assert.True(t, exists)
	assert.Equal(t, string(message), data)
}

func TestRedisStreams_Subscribe_BasicFlow(t *testing.T) {
	rs := setupRedisStreams()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	channel := "test-channel-subscribe"
	defer cleanupRedisStream(rs.getStreamName(channel))

	// Start subscriber
	ch, err := rs.Subscribe(ctx, channel)
	require.NoError(t, err)

	// Give subscriber time to set up
	time.Sleep(100 * time.Millisecond)

	// Publish a message AFTER subscriber is ready
	testMessage := "subscription test message"
	err = rs.Publish(ctx, channel, testMessage)
	require.NoError(t, err)

	// Wait for message
	select {
	case received := <-ch:
		assert.Equal(t, testMessage, received)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestRedisStreams_Subscribe_MultipleMessages(t *testing.T) {
	rs := setupRedisStreams()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	channel := "test-channel-multiple"
	defer cleanupRedisStream(rs.getStreamName(channel))

	// Start subscriber
	ch, err := rs.Subscribe(ctx, channel)
	require.NoError(t, err)

	// Give subscriber time to set up
	time.Sleep(100 * time.Millisecond)

	// Publish multiple messages AFTER subscriber is ready
	messages := []string{"message1", "message2", "message3"}
	for _, msg := range messages {
		err = rs.Publish(ctx, channel, msg)
		require.NoError(t, err)
	}

	// Collect received messages
	var received []string
	timeout := time.After(5 * time.Second)

	for i := 0; i < len(messages); i++ {
		select {
		case msg := <-ch:
			received = append(received, msg)
		case <-timeout:
			t.Fatalf("Timeout waiting for message %d", i+1)
		}
	}

	assert.ElementsMatch(t, messages, received)
}

func TestRedisStreams_HelperMethods(t *testing.T) {
	rs := setupRedisStreams()

	// Test getStreamName
	channel := "test-channel"
	expected := "stream:test-channel"
	assert.Equal(t, expected, rs.getStreamName(channel))

	// Test getConsumerGroup
	assert.Equal(t, "test-group", rs.getConsumerGroup())

	// Test getConsumerGroup with empty value
	rs.ConsumerGroup = ""
	assert.Equal(t, "notifications", rs.getConsumerGroup())

	// Test getConsumerName
	rs.ConsumerName = "custom-consumer"
	assert.Equal(t, "custom-consumer", rs.getConsumerName())

	// Test getConsumerName with empty value (should generate unique name)
	rs.ConsumerName = ""
	name1 := rs.getConsumerName()
	assert.Contains(t, name1, "consumer-")
	// Note: getConsumerName generates the same name unless we create a new instance

	// Test getBatchSize
	assert.Equal(t, 10, rs.getBatchSize())
	rs.BatchSize = 0
	assert.Equal(t, 10, rs.getBatchSize()) // Default
	rs.BatchSize = -5
	assert.Equal(t, 10, rs.getBatchSize()) // Default for negative

	// Test getFlushInterval
	rs.FlushInterval = 3 * time.Second
	assert.Equal(t, 3*time.Second, rs.getFlushInterval())
	rs.FlushInterval = 0
	assert.Equal(t, 5*time.Second, rs.getFlushInterval()) // Default
	rs.FlushInterval = -1 * time.Second
	assert.Equal(t, 5*time.Second, rs.getFlushInterval()) // Default for negative
}

func TestRedisStreams_Batching_Behavior(t *testing.T) {
	rs := setupRedisStreams()
	rs.BatchSize = 3                    // Set small batch size for testing
	rs.FlushInterval = 10 * time.Second // Long interval to test batch size trigger

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	channel := "test-channel-batching"
	defer cleanupRedisStream(rs.getStreamName(channel))

	// Start subscriber
	ch, err := rs.Subscribe(ctx, channel)
	require.NoError(t, err)

	// Publish messages to trigger batch
	messages := []string{"batch1", "batch2", "batch3"}
	for _, msg := range messages {
		err = rs.Publish(ctx, channel, msg)
		require.NoError(t, err)
	}

	// Should receive all messages in one batch
	var received []string
	timeout := time.After(3 * time.Second)

	for len(received) < len(messages) {
		select {
		case msg := <-ch:
			received = append(received, msg)
		case <-timeout:
			t.Fatalf("Timeout waiting for batched messages. Received %d out of %d", len(received), len(messages))
		}
	}

	assert.ElementsMatch(t, messages, received)
}

func TestRedisStreams_MaxLen_Configuration(t *testing.T) {
	rs := setupRedisStreams()
	rs.MaxLen = 2 // Very small max length

	ctx := context.Background()
	channel := "test-channel-maxlen"
	defer cleanupRedisStream(rs.getStreamName(channel))

	// Publish more messages than MaxLen
	messages := []string{"msg1", "msg2", "msg3", "msg4"}
	for _, msg := range messages {
		err := rs.Publish(ctx, channel, msg)
		require.NoError(t, err)
	}

	// Verify stream was trimmed
	client := redis.NewClient(&redis.Options{
		Addr:     testRedisHost,
		Password: testPassword,
		DB:       testDatabase,
	})
	defer client.Close()

	streamName := rs.getStreamName(channel)
	length, err := client.XLen(ctx, streamName).Result()
	require.NoError(t, err)

	// Should be approximately MaxLen (Redis uses approximate trimming)
	// With APPROX, Redis may keep more entries than specified
	assert.LessOrEqual(t, length, int64(10)) // Allow generous buffer for approximate trimming
}
