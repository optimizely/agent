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

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/syncer/pubsub"
)

func TestNewPubSub(t *testing.T) {
	type args struct {
		conf config.SyncConfig
		flag SycnFeatureFlag
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
				flag: SyncFeatureFlagNotificaiton,
			},
			want: &pubsub.Redis{
				Host:     "localhost:6379",
				Password: "",
				Database: 0,
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
				flag: SycnFeatureFlagDatafile,
			},
			want: &pubsub.Redis{
				Host:     "localhost:6379",
				Password: "",
				Database: 0,
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
				flag: SyncFeatureFlagNotificaiton,
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
				flag: SyncFeatureFlagNotificaiton,
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
				flag: SyncFeatureFlagNotificaiton,
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
				flag: SyncFeatureFlagNotificaiton,
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
				flag: SyncFeatureFlagNotificaiton,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test with invalid redis config without password",
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
				flag: SyncFeatureFlagNotificaiton,
			},
			want:    nil,
			wantErr: true,
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
				flag: SyncFeatureFlagNotificaiton,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test with invalid redis config with invalid password",
			args: args{
				conf: config.SyncConfig{
					Pubsub: map[string]interface{}{
						"redis": map[string]interface{}{
							"host":     "localhost:6379",
							"password": 1234,
							"database": 0,
						},
					},
					Notification: config.FeatureSyncConfig{
						Default: "redis",
						Enable:  true,
					},
				},
				flag: SyncFeatureFlagNotificaiton,
			},
			want:    nil,
			wantErr: true,
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
				flag: SyncFeatureFlagNotificaiton,
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
