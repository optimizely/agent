/****************************************************************************
 * Copyright 2020,2023 Optimizely, Inc. and contributors                    *
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

// Package handlers //
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/syncer"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/registry"
	"github.com/rs/zerolog"
)

const (
	LoggerKey = "notification-logger"
	SDKKey    = "context-sdk-key"
)

// A MessageChan is a channel of bytes
// Each http handler call creates a new channel and pumps decision service messages onto it.
type MessageChan chan []byte

type NotificationReceiverFunc func(context.Context) (<-chan syncer.Event, error)

// types of notifications supported.
var types = map[notification.Type]string{
	notification.Decision:            string(notification.Decision),
	notification.Track:               string(notification.Track),
	notification.ProjectConfigUpdate: string(notification.ProjectConfigUpdate),
}

func getFilter(filters []string) map[notification.Type]string {
	notificationsToAdd := make(map[notification.Type]string)
	// Parse out the any filters that were added
	if len(filters) == 0 {
		notificationsToAdd = types
	}
	// iterate through any filter query parameter included.  There may be more than one
	for _, filter := range filters {
		// split it in case it is comma separated list
		splits := strings.Split(filter, ",")
		for _, split := range splits {
			// if the string is a valid type
			if _, ok := types[notification.Type(split)]; ok {
				notificationsToAdd[notification.Type(split)] = split
			}
		}
	}

	return notificationsToAdd
}

func NotificationEventStreamHandler(notificationReceiverFn NotificationReceiverFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Make sure that the writer supports flushing.
		flusher, ok := w.(http.Flusher)

		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		_, err := middleware.GetOptlyClient(r)

		if err != nil {
			RenderError(err, http.StatusUnprocessableEntity, w, r)
			return
		}

		// Set the headers related to event streaming.
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// "raw" query string option
		// If provided, send raw JSON lines instead of SSE-compliant strings.
		raw := len(r.URL.Query()["raw"]) > 0

		// Parse the form.
		_ = r.ParseForm()

		filters := r.Form["filter"]

		// Parse out the any filters that were added
		notificationsToAdd := getFilter(filters)

		// Listen to connection close and un-register messageChan
		notify := r.Context().Done()

		sdkKey := r.Header.Get(middleware.OptlySDKHeader)
		ctx := context.WithValue(r.Context(), SDKKey, sdkKey)

		dataChan, err := notificationReceiverFn(context.WithValue(ctx, LoggerKey, middleware.GetLogger(r)))
		if err != nil {
			middleware.GetLogger(r).Err(err).Msg("error from receiver")
			http.Error(w, "Error from data receiver!", http.StatusInternalServerError)
			return
		}

		for {
			select {
			case <-notify:
				middleware.GetLogger(r).Debug().Msg("received close on the request.  So, we are shutting down this handler")
				return
			case event := <-dataChan:
				_, found := notificationsToAdd[event.Type]
				if !found {
					continue
				}

				jsonEvent, err := json.Marshal(event.Message)
				if err != nil {
					middleware.GetLogger(r).Err(err).Msg("failed to marshal notification into json")
					continue
				}

				if raw {
					// Raw JSON events, one per line
					_, _ = fmt.Fprintf(w, "%s\n", string(jsonEvent))
				} else {
					// Server Sent Events compatible
					_, _ = fmt.Fprintf(w, "data: %s\n\n", string(jsonEvent))
				}
				// Flush the data immediately instead of buffering it for later.
				// The flush will fail if the connection is closed.  That will cause the handler to exit.
				flusher.Flush()
			}
		}
	}
}

func DefaultNotificationReceiver(ctx context.Context) (<-chan syncer.Event, error) {
	logger, ok := ctx.Value(LoggerKey).(*zerolog.Logger)
	if !ok {
		logger = &zerolog.Logger{}
	}

	sdkKey, ok := ctx.Value(SDKKey).(string)
	if !ok || sdkKey == "" {
		return nil, errors.New("sdk key not found")
	}

	// Each connection registers its own message channel with the NotificationHandler's connections registry
	messageChan := make(chan syncer.Event)
	nc := registry.GetNotificationCenter(sdkKey)

	// Parse out the any filters that were added
	notificationsToAdd := types

	ids := []struct {
		int
		notification.Type
	}{}

	for notificationType := range notificationsToAdd {
		id, e := nc.AddHandler(notificationType, func(n interface{}) {
			msg := syncer.Event{
				Type:    notificationType,
				Message: n,
			}
			messageChan <- msg
		})
		if e != nil {
			return nil, e
		}

		// do defer outside the loop.
		ids = append(ids, struct {
			int
			notification.Type
		}{id, notificationType})
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				for _, id := range ids {
					err := nc.RemoveHandler(id.int, id.Type)
					if err != nil {
						logger.Err(err).AnErr("error in removing notification handler", err)
					}
				}
				return
			}
		}
	}()

	return messageChan, nil
}

func RedisNotificationReceiver(conf config.SyncConfig) NotificationReceiverFunc {
	return func(ctx context.Context) (<-chan syncer.Event, error) {
		sdkKey, ok := ctx.Value(SDKKey).(string)
		if !ok || sdkKey == "" {
			return nil, errors.New("sdk key not found")
		}

		redisSyncer, err := syncer.NewRedisNotificationSyncer(&zerolog.Logger{}, conf, sdkKey)
		if err != nil {
			return nil, err
		}

		client := redis.NewClient(&redis.Options{
			Addr:     redisSyncer.Host,
			Password: redisSyncer.Password,
			DB:       redisSyncer.Database,
		})

		// Subscribe to a Redis channel
		pubsub := client.Subscribe(ctx, syncer.GetChannelForSDKKey(redisSyncer.Channel, sdkKey))

		dataChan := make(chan syncer.Event)

		logger, ok := ctx.Value(LoggerKey).(*zerolog.Logger)
		if !ok {
			logger = &zerolog.Logger{}
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					client.Close()
					pubsub.Close()
					logger.Debug().Msg("context canceled, redis notification receiver is closed")
					return
				default:
					msg, err := pubsub.ReceiveMessage(ctx)
					if err != nil {
						logger.Err(err).Msg("failed to receive message from redis")
						continue
					}

					var event syncer.Event
					if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
						logger.Err(err).Msg("failed to unmarshal redis message")
						continue
					}
					dataChan <- event
				}
			}
		}()

		return dataChan, nil
	}
}
