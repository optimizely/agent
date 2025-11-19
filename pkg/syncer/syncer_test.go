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
	"context"
	"reflect"
	"testing"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/syncer/pubsub"
	"github.com/optimizely/go-sdk/v2/pkg/notification"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

type testPubSub struct {
	publishCalled   bool
	subscribeCalled bool
}

func (r *testPubSub) Publish(ctx context.Context, channel string, message interface{}) error {
	r.publishCalled = true
	return nil
}

func (r *testPubSub) Subscribe(ctx context.Context, channel string) (chan string, error) {
	r.subscribeCalled = true
	return nil, nil
}

func TestNewSyncedNotificationCenter(t *testing.T) {
	type args struct {
		ctx    context.Context
		sdkKey string
		conf   config.SyncConfig
	}
	tests := []struct {
		name    string
		args    args
		want    NotificationSyncer
		wantErr bool
	}{
		{
			name: "Test with valid config",
			args: args{
				ctx:    context.Background(),
				sdkKey: "123",
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": map[string]interface{}{
							"host":                 "localhost:6379",
							"password":             "",
							"database":             0,
							"force_implementation": "pubsub", // Force Pub/Sub for deterministic test
						},
					},
					Notification: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
			},
			want: &SyncedNotificationCenter{
				ctx:    context.Background(),
				logger: &log.Logger,
				sdkKey: "123",
				pubsub: &pubsub.Redis{
					Host:     "localhost:6379",
					Password: "",
					Database: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "Test with invalid sync config",
			args: args{
				ctx:    context.Background(),
				sdkKey: "1234",
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"not-redis": map[string]interface{}{
							"host": "invalid host",
						},
					},
					Notification: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test with empty sync config",
			args: args{
				ctx:    context.Background(),
				sdkKey: "1234",
				conf:   config.SyncConfig{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSyncedNotificationCenter(tt.args.ctx, tt.args.sdkKey, tt.args.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSyncedNotificationCenter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSyncedNotificationCenter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDatafileSyncer(t *testing.T) {
	type args struct {
		conf config.SyncConfig
	}
	tests := []struct {
		name    string
		args    args
		want    *DatafileSyncer
		wantErr bool
	}{
		{
			name: "Test with valid config",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": map[string]interface{}{
							"host":                 "localhost:6379",
							"password":             "",
							"database":             0,
							"force_implementation": "pubsub", // Force Pub/Sub for deterministic test
						},
					},
					Datafile: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
			},
			want: &DatafileSyncer{
				pubsub: &pubsub.Redis{
					Host:     "localhost:6379",
					Password: "",
					Database: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "Test with invalid sync config",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"not-redis": map[string]interface{}{
							"host": "invalid host",
						},
					},
					Datafile: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDatafileSyncer(tt.args.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDatafileSyncer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDatafileSyncer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatafileSyncer_Sync(t *testing.T) {
	type fields struct {
		pubsub PubSub
	}
	type args struct {
		ctx     context.Context
		channel string
		sdkKey  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test datafile sync",
			fields: fields{
				pubsub: &testPubSub{},
			},
			args: args{
				ctx:     context.Background(),
				channel: "test-ch",
				sdkKey:  "123",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &DatafileSyncer{
				pubsub: tt.fields.pubsub,
			}
			if err := r.Sync(tt.args.ctx, tt.args.channel, tt.args.sdkKey); (err != nil) != tt.wantErr {
				t.Errorf("DatafileSyncer.Sync() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.True(t, tt.fields.pubsub.(*testPubSub).publishCalled)
		})
	}
}

func TestDatafileSyncer_Subscribe(t *testing.T) {
	type fields struct {
		pubsub PubSub
	}
	type args struct {
		ctx     context.Context
		channel string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test datafile sync",
			fields: fields{
				pubsub: &testPubSub{},
			},
			args: args{
				ctx:     context.Background(),
				channel: "test-ch",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &DatafileSyncer{
				pubsub: tt.fields.pubsub,
			}
			_, err := r.Subscribe(tt.args.ctx, tt.args.channel)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatafileSyncer.Subscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.True(t, tt.fields.pubsub.(*testPubSub).subscribeCalled)
		})
	}
}

func TestSyncedNotificationCenter_Send(t *testing.T) {
	type fields struct {
		ctx    context.Context
		logger *zerolog.Logger
		sdkKey string
		pubsub PubSub
	}
	type args struct {
		t notification.Type
		n interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test notification send",
			fields: fields{
				ctx:    context.Background(),
				logger: &log.Logger,
				sdkKey: "123",
				pubsub: &testPubSub{},
			},
			args: args{
				t: notification.Decision,
				n: "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SyncedNotificationCenter{
				ctx:    tt.fields.ctx,
				logger: tt.fields.logger,
				sdkKey: tt.fields.sdkKey,
				pubsub: tt.fields.pubsub,
			}
			if err := r.Send(tt.args.t, tt.args.n); (err != nil) != tt.wantErr {
				t.Errorf("SyncedNotificationCenter.Send() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.True(t, tt.fields.pubsub.(*testPubSub).publishCalled)
		})
	}
}

func TestSyncedNotificationCenter_Subscribe(t *testing.T) {
	type fields struct {
		ctx    context.Context
		logger *zerolog.Logger
		sdkKey string
		pubsub PubSub
	}
	type args struct {
		ctx     context.Context
		channel string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test notification send",
			fields: fields{
				ctx:    context.Background(),
				logger: &log.Logger,
				sdkKey: "123",
				pubsub: &testPubSub{},
			},
			args: args{
				ctx:     context.Background(),
				channel: "test-ch",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SyncedNotificationCenter{
				ctx:    tt.fields.ctx,
				logger: tt.fields.logger,
				sdkKey: tt.fields.sdkKey,
				pubsub: tt.fields.pubsub,
			}
			_, err := r.Subscribe(tt.args.ctx, tt.args.channel)
			if (err != nil) != tt.wantErr {
				t.Errorf("SyncedNotificationCenter.Subscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.True(t, tt.fields.pubsub.(*testPubSub).subscribeCalled)
		})
	}
}

func TestNewSyncedNotificationCenter_CacheHit(t *testing.T) {
	// Clear cache before test
	ncCache = make(map[string]NotificationSyncer)

	conf := config.SyncConfig{
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
	}

	sdkKey := "test-sdk-key"
	ctx := context.Background()

	// First call - should create new instance
	nc1, err := NewSyncedNotificationCenter(ctx, sdkKey, conf)
	assert.NoError(t, err)
	assert.NotNil(t, nc1)

	// Second call with same sdkKey - should return cached instance
	nc2, err := NewSyncedNotificationCenter(ctx, sdkKey, conf)
	assert.NoError(t, err)
	assert.NotNil(t, nc2)

	// Should be the same instance (cache hit)
	assert.Equal(t, nc1, nc2)
}

func TestSyncedNotificationCenter_AddHandler(t *testing.T) {
	nc := &SyncedNotificationCenter{
		ctx:    context.Background(),
		logger: &log.Logger,
		sdkKey: "test",
		pubsub: &testPubSub{},
	}

	id, err := nc.AddHandler(notification.Decision, func(interface{}) {})
	assert.NoError(t, err)
	assert.Equal(t, 0, id)
}

func TestSyncedNotificationCenter_RemoveHandler(t *testing.T) {
	nc := &SyncedNotificationCenter{
		ctx:    context.Background(),
		logger: &log.Logger,
		sdkKey: "test",
		pubsub: &testPubSub{},
	}

	err := nc.RemoveHandler(0, notification.Decision)
	assert.NoError(t, err)
}

func TestSyncedNotificationCenter_Send_MarshalError(t *testing.T) {
	nc := &SyncedNotificationCenter{
		ctx:    context.Background(),
		logger: &log.Logger,
		sdkKey: "test",
		pubsub: &testPubSub{},
	}

	// Pass a channel which cannot be marshaled to JSON
	ch := make(chan int)
	err := nc.Send(notification.Decision, ch)
	assert.Error(t, err)
}

func TestGetDatafileSyncChannel(t *testing.T) {
	result := GetDatafileSyncChannel()
	expected := "optimizely-sync-datafile"
	assert.Equal(t, expected, result)
}

func TestGetChannelForSDKKey(t *testing.T) {
	result := GetChannelForSDKKey("test-channel", "sdk-123")
	expected := "test-channel-sdk-123"
	assert.Equal(t, expected, result)
}
