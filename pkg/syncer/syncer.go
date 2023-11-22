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
	"encoding/json"
	"fmt"
	"sync"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/rs/zerolog"
)

var (
	ncCache   = make(map[string]*SyncedNotificationCenter)
	mutexLock = &sync.Mutex{}
)

type NotificationSyncer interface {
	notification.Center
	Subscribe(ctx context.Context, channel string) (chan string, error)
}

// RedisSyncer defines Redis pubsub configuration
type SyncedNotificationCenter struct {
	ctx    context.Context
	logger *zerolog.Logger
	sdkKey string
	pubsub PubSub
}

// Event holds the notification event with it's type
type Event struct {
	Type    notification.Type `json:"type"`
	Message interface{}       `json:"message"`
}

func NewSyncedNotificationCenter(ctx context.Context, logger *zerolog.Logger, sdkKey string, conf config.SyncConfig) (NotificationSyncer, error) {
	mutexLock.Lock()
	defer mutexLock.Unlock()

	if nc, ok := ncCache[sdkKey]; ok {
		return nc, nil
	}

	pubsub, err := NewPubSub(conf)
	if err != nil {
		return nil, err
	}

	nc := &SyncedNotificationCenter{
		ctx:    ctx,
		logger: logger,
		sdkKey: sdkKey,
		pubsub: pubsub,
	}
	ncCache[sdkKey] = nc
	return nc, nil
}

// AddHandler is empty but needed to implement notification.Center interface
func (r *SyncedNotificationCenter) AddHandler(_ notification.Type, _ func(interface{})) (int, error) {
	return 0, nil
}

// RemoveHandler is empty but needed to implement notification.Center interface
func (r *SyncedNotificationCenter) RemoveHandler(_ int, t notification.Type) error {
	return nil
}

// Send will send the notification to the specified channel in the Redis pubsub
func (r *SyncedNotificationCenter) Send(t notification.Type, n interface{}) error {
	event := Event{
		Type:    t,
		Message: n,
	}

	jsonEvent, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return r.pubsub.Publish(r.ctx, GetChannelForSDKKey(PubSubDefaultChan, r.sdkKey), jsonEvent)
}

func (r *SyncedNotificationCenter) Subscribe(ctx context.Context, channel string) (chan string, error) {
	return r.pubsub.Subscribe(ctx, channel)
}

func GetDatafileSyncChannel() string {
	return fmt.Sprintf("%s-datafile", PubSubDefaultChan)
}

func GetChannelForSDKKey(channel, key string) string {
	return fmt.Sprintf("%s-%s", channel, key)
}

type DatafileSyncer struct {
	pubsub PubSub
}

func NewDatafileSyncer(conf config.SyncConfig) (*DatafileSyncer, error) {
	pubsub, err := NewPubSub(conf)
	if err != nil {
		return nil, err
	}

	return &DatafileSyncer{
		pubsub: pubsub,
	}, nil
}

func (r *DatafileSyncer) Sync(ctx context.Context, channel string, sdkKey string) error {
	return r.pubsub.Publish(ctx, channel, sdkKey)
}

func (r *DatafileSyncer) Subscribe(ctx context.Context, channel string) (chan string, error) {
	return r.pubsub.Subscribe(ctx, channel)
}
