/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/syncer"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/registry"
)

// A MessageChan is a channel of bytes
// Each http handler call creates a new channel and pumps decision service messages onto it.
type MessageChan chan []byte

// types of notifications supported.
var types = map[string]notification.Type{
	string(notification.Decision):            notification.Decision,
	string(notification.Track):               notification.Track,
	string(notification.ProjectConfigUpdate): notification.ProjectConfigUpdate,
}

func getFilter(filters []string) map[string]notification.Type {
	notificationsToAdd := map[string]notification.Type{}
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
			if _, ok := types[split]; ok {
				notificationsToAdd[split] = notification.Type(split)
			}
		}
	}

	return notificationsToAdd
}

func NotificationEventStreamHandler(syncConfig *config.SyncConfig) http.HandlerFunc {
	if !syncConfig.Notification.Enable {
		return NotificationEventSteamMonolithHandler
	}
	return NotificationEventSteamSyncHandler(syncConfig)
}

// NotificationEventSteamHandler implements the http.Handler interface.
func NotificationEventSteamMonolithHandler(w http.ResponseWriter, r *http.Request) {
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

	// Each connection registers its own message channel with the NotificationHandler's connections registry
	messageChan := make(MessageChan)
	// Each connection also adds listeners
	sdkKey := r.Header.Get(middleware.OptlySDKHeader)
	nc := registry.GetNotificationCenter(sdkKey)

	// Parse the form.
	_ = r.ParseForm()

	filters := r.Form["filter"]

	// Parse out the any filters that were added
	notificationsToAdd := getFilter(filters)

	ids := []struct {
		int
		notification.Type
	}{}

	for _, value := range notificationsToAdd {
		id, e := nc.AddHandler(value, func(n interface{}) {
			jsonEvent, err := json.Marshal(n)
			if err != nil {
				middleware.GetLogger(r).Error().Msg("encoding notification to json")
			} else {
				messageChan <- jsonEvent
			}
		})
		if e != nil {
			RenderError(e, http.StatusUnprocessableEntity, w, r)
			return
		}

		// do defer outside the loop.
		ids = append(ids, struct {
			int
			notification.Type
		}{id, value})
	}

	// Remove the decision listener if we exited.
	defer func() {
		for _, id := range ids {
			err := nc.RemoveHandler(id.int, id.Type)
			if err != nil {
				middleware.GetLogger(r).Error().AnErr("removing notification", err)
			}
		}
	}()

	// "raw" query string option
	// If provided, send raw JSON lines instead of SSE-compliant strings.
	raw := len(r.Form["raw"]) > 0

	// Listen to connection close and un-register messageChan
	notify := r.Context().Done()
	// block waiting or messages broadcast on this connection's messageChan
	for {
		select {
		// Write to the ResponseWriter
		case msg := <-messageChan:
			if raw {
				// Raw JSON events, one per line
				_, _ = fmt.Fprintf(w, "%s\n", msg)
			} else {
				// Server Sent Events compatible
				_, _ = fmt.Fprintf(w, "data: %s\n\n", msg)
			}
			// Flush the data immediately instead of buffering it for later.
			// The flush will fail if the connection is closed.  That will cause the handler to exit.
			flusher.Flush()
		case <-notify:
			middleware.GetLogger(r).Debug().Msg("received close on the request.  So, we are shutting down this handler")
			return
		}
	}

}

// NotificationEventSteamHandler implements the http.Handler interface.
func NotificationEventSteamSyncHandler(conf *config.SyncConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Make sure that the writer supports flushing.
		flusher, ok := w.(http.Flusher)

		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		// Set the headers related to event streaming.
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		redisSyncer, err := syncer.NewRedisPubSubSyncer(middleware.GetLogger(r), conf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		client := redis.NewClient(&redis.Options{
			Addr:     redisSyncer.Host,
			Password: redisSyncer.Password,
			DB:       redisSyncer.Database,
		})
		defer client.Close()

		// Subscribe to a Redis channel
		pubsub := client.Subscribe(r.Context(), redisSyncer.Channel)
		defer pubsub.Close()

		// "raw" query string option
		// If provided, send raw JSON lines instead of SSE-compliant strings.
		raw := len(r.Form["raw"]) > 0

		// Listen to connection close and un-register messageChan
		notify := r.Context().Done()

		for {
			select {
			case <-notify:
				middleware.GetLogger(r).Debug().Msg("received close on the request.  So, we are shutting down this handler")
				return
			default:
				log.Println("looking for redis message")
				msg, err := pubsub.ReceiveMessage(r.Context())
				if err != nil {
					log.Println("Error receiving message:", err)
					return
				}
				if raw {
					// Raw JSON events, one per line
					_, _ = fmt.Fprintf(w, "%s\n", msg.Payload)
				} else {
					// Server Sent Events compatible
					_, _ = fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
				}
				// Flush the data immediately instead of buffering it for later.
				// The flush will fail if the connection is closed.  That will cause the handler to exit.
				flusher.Flush()
			}
		}
	}
}
