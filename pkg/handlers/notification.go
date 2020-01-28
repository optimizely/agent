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
	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/registry"
	"net/http"
)

// A MessageChan is a channel of bytes
// Each http handler call creates a new channel and pumps decision service messages onto it.
type MessageChan chan []byte

// A NotificationHandler handles in coming connections as server side event streams streaming notifications
// per SDK Key (defined in the header)
type NotificationHandler struct {
}

func contains(arr []string, element string) bool {
	for _, e := range arr {
		if e == element {
			return true
		}
	}

	return false
}

func sendNotificationToChannel(n interface{}, messChan *MessageChan, r *http.Request) {
	switch v := n.(type) {
	case notification.DecisionNotification, notification.TrackNotification, notification.ProjectConfigUpdateNotification:
		jsonEvent, err := json.Marshal(v)
		if err != nil {
			middleware.GetLogger(r).Error().Msg("encoding notification to json")
		} else {
			*messChan <- jsonEvent
		}
	}
}

// types of notifications supported.
var types = []notification.Type{notification.Decision, notification.Track, notification.ProjectConfigUpdate}

// HandleEventSteam implements the http.Handler interface.
// This allows us to wrap HTTP handlers (see auth_handler.go)
// http://golang.org/pkg/net/http/#Handler
func (nh *NotificationHandler) HandleEventSteam(w http.ResponseWriter, r *http.Request) {
	// Make sure that the writer supports flushing.
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	optlyClient, err := middleware.GetOptlyClient(r)

	if err != nil {
		RenderError(err, http.StatusUnprocessableEntity, w, r)
		return
	}

	if optlyClient == nil {
		e := fmt.Errorf("optlyContext not available")
		RenderError(e, http.StatusUnprocessableEntity, w, r)
		return
	}

	// Parse the form.
	_ = r.ParseForm()

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the NotificationHandler's connections registry
	messageChan := make(MessageChan)
	// Each connection also adds one decision listener
	sdkKey := r.Header.Get(middleware.OptlySDKHeader)
	nc := registry.GetNotificationCenter(sdkKey)

	filters := r.Form["filter"]

	var ids []int

	for _, notificationType := range types {
		if !contains(filters, string(notificationType)) {
			id,e := nc.AddHandler(notificationType, func (n interface{}) {
				sendNotificationToChannel(n, &messageChan, r)
			})
			if e != nil {
				RenderError(e, http.StatusUnprocessableEntity, w, r)
				return
			}
			ids = append(ids, id)
		} else {
			ids = append(ids, 0)
		}
	}

	// Remove the decision listener if we exited.
	defer func() {
		for i, id := range ids {
			if id == 0 {
				continue
			}
			err := nc.RemoveHandler(id, types[i])
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

// NewEventStreamHandler is the NotificationHandler factory
func NewEventStreamHandler() (nh *NotificationHandler) {
	// Instantiate a nh
	nh = &NotificationHandler{}

	return
}
