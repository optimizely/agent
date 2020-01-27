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
	"net/http"
)

// A MessageChan is a channel of bytes
// Each http handler call creates a new channel and pumps decision service messages onto it.
type MessageChan chan []byte

// A NotificationHandler handles in coming connections as server side event streams streaming notifications
// per SDK Key (defined in the header)
type NotificationHandler struct {
}

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

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the NotificationHandler's connections registry
	messageChan := make(MessageChan)
	// Each connection also adds one decision listener
	id, err2 := optlyClient.DecisionService.OnDecision(func(decision notification.DecisionNotification) {
		jsonEvent, err := json.Marshal(decision)
		if err != nil {
			middleware.GetLogger(r).Error().Str("decision", string(decision.Type)).Msg("encoding decision to json")
		} else {
			messageChan <- jsonEvent
		}
	})

	if err2 != nil {
		RenderError(err2, http.StatusUnprocessableEntity, w, r)
		return
	}

	// Remove the decision listener if we exited.
	defer func() {
		err := optlyClient.DecisionService.RemoveOnDecision(id)
		if err != nil {
			middleware.GetLogger(r).Error().AnErr("removingOnDecision", err)
		}
	}()

	// "raw" query string option
	// If provided, send raw JSON lines instead of SSE-compliant strings.
	_ = r.ParseForm()
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
			fmt.Println("received message", msg)
		// Remove the decision listener if the connection is closed and exit
		case sig := <-notify:
			err := optlyClient.DecisionService.RemoveOnDecision(id)
			if err != nil {
				middleware.GetLogger(r).Error().AnErr("removingOnDecision", err)
			}
			fmt.Println("received close on the request.  So, we are shutting down this handler", sig)
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
