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

// A EventStreamHandler handles in coming connections,
type EventStreamHandler struct {
}

// HandleEventSteam implements the http.Handler interface.
// This allows us to wrap HTTP handlers (see auth_handler.go)
// http://golang.org/pkg/net/http/#Handler
func (esh *EventStreamHandler) HandleEventSteam(rw http.ResponseWriter, req *http.Request) {
	// Make sure that the writer supports flushing.
	//
	flusher, ok := rw.(http.Flusher)

	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	optlyClient, err := middleware.GetOptlyClient(req)

	if err != nil {
		RenderError(err, http.StatusUnprocessableEntity, rw, req)
		return
	}

	if optlyClient == nil {
		err := fmt.Errorf("optlyContext not available")
		RenderError(err, http.StatusUnprocessableEntity, rw, req)
		return
	}

	// Set the headers related to event streaming.
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the EventStreamHandler's connections registry
	messageChan := make(MessageChan)

	// Each connection also adds one decision listener
	id,err := optlyClient.DecisionService.OnDecision(func(decision notification.DecisionNotification) {
		jsonEvent, err := json.Marshal(decision)
		if err != nil {
			middleware.GetLogger(req).Error().Str("decision", string(decision.Type)).Msg("encoding decision to json")
		} else {
			messageChan <- jsonEvent
		}
	})

	if err != nil {
		RenderError(err, http.StatusUnprocessableEntity, rw, req)
		return
	}

	// Remove the decision listener if we exited.
	defer func() {
		optlyClient.DecisionService.RemoveOnDecision(id)
	}()

	// "raw" query string option
	// If provided, send raw JSON lines instead of SSE-compliant strings.
	req.ParseForm()
	raw := len(req.Form["raw"]) > 0

	// Listen to connection close and un-register messageChan
	notify := req.Context().Done()

	// remove the decision listener if the connection is closed
	go func() {
		<-notify
		optlyClient.DecisionService.RemoveOnDecision(id)
	}()

	// block waiting or messages broadcast on this connection's messageChan
	for {
		// Write to the ResponseWriter
		if raw {
			// Raw JSON events, one per line
			fmt.Fprintf(rw, "%s\n", <-messageChan)
		} else {
			// Server Sent Events compatible
			fmt.Fprintf(rw, "data: %s\n\n", <-messageChan)
		}
		// Flush the data immediately instead of buffering it for later.
		// The flush will fail if the connection is closed.  That will cause the handler to exit.
		flusher.Flush()
	}

}

// NewEventStreamHandler is the EventStreamHandler factory
func NewEventStreamHandler() (esh *EventStreamHandler) {
	// Instantiate a esh
	esh = &EventStreamHandler{}

	return
}
