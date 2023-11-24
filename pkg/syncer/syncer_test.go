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
			},
			want: &SyncedNotificationCenter{
				ctx:    context.Background(),
				logger: &log.Logger,
				sdkKey: "123",
				pubsub: &pubsubRedis{
					host:     "localhost:6379",
					password: "",
					database: 0,
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
			},
			want: &DatafileSyncer{
				pubsub: &pubsubRedis{
					host:     "localhost:6379",
					password: "",
					database: 0,
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
					Notification: config.FeatureSyncConfig{
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
