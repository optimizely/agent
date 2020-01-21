package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/go-sdk/pkg/notification"
	"log"
	"net/http"
)

// A MessageChan is a channel of channels
// Each connection sends a channel of bytes to a global MessageChan
// The main broker listen() loop listens on new connections on MessageChan
// New event messages are broadcast to all registered connection channels
type MessageChan chan []byte

// A EventStreamBroker holds open client connections,
// listens for incoming events on its Notifier channel
// and broadcast event data to all registered connections
type EventStreamBroker struct {

	// Events are pushed to this channel by the registered decision listener
	Notifier chan notification.DecisionNotification

	// New client connections
	newClients chan MessageChan

	// Closed client connections
	closingClients chan MessageChan

	// Client connections registry
	clients map[MessageChan]bool
}

// Listen on different channels and act accordingly
func (broker *EventStreamBroker) listen() {
	for {
		select {
		case s := <-broker.newClients:

			// A new client has connected.
			// Register their message channel
			broker.clients[s] = true
			log.Printf("Client added. %d registered clients", len(broker.clients))
		case s := <-broker.closingClients:

			// A client has dettached and we want to
			// stop sending them messages.
			delete(broker.clients, s)
			log.Printf("Removed client. %d registered clients", len(broker.clients))
		case event := <-broker.Notifier:

			// We got a new event from the outside!
			// Send event to all connected clients
			jsonEvent, err := json.Marshal(event)
			if err != nil {
				log.Println("Error encoding event to JSON")
			} else {
				for clientMessageChan, _ := range broker.clients {
					clientMessageChan <- jsonEvent
				}
				log.Printf("Broadcast message to %d clients", len(broker.clients))
			}
		}
	}

}

// Implement the http.Handler interface.
// This allows us to wrap HTTP handlers (see auth_handler.go)
// http://golang.org/pkg/net/http/#Handler
func (broker *EventStreamBroker) HandleEventSteam(rw http.ResponseWriter, req *http.Request) {
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

	// Set the headers related to event streaming.
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the EventStreamBroker's connections registry
	messageChan := make(MessageChan)

	// Signal the broker that we have a new connection
	broker.newClients <- messageChan

	id,err := optlyClient.DecisionService.OnDecision(func(decision notification.DecisionNotification) {
		broker.Notifier <- decision
	})

	if err != nil {
		RenderError(err, http.StatusUnprocessableEntity, rw, req)
		broker.closingClients <- messageChan
		return
	}

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		optlyClient.DecisionService.RemoveOnDecision(id)
		broker.closingClients <- messageChan
	}()

	// "raw" query string option
	// If provided, send raw JSON lines instead of SSE-compliant strings.
	req.ParseForm()
	raw := len(req.Form["raw"]) > 0

	// Listen to connection close and un-register messageChan
	notify := req.Context().Done()

	go func() {
		<-notify
		optlyClient.DecisionService.RemoveOnDecision(id)
		broker.closingClients <- messageChan
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
		// Flush the data inmediatly instead of buffering it for later.
		flusher.Flush()
	}
}

// EventStreamBroker factory
func NewEventStreamHandler() (broker *EventStreamBroker) {
	// Instantiate a broker
	broker = &EventStreamBroker{
		Notifier:       make(chan notification.DecisionNotification, 1),
		newClients:     make(chan MessageChan),
		closingClients: make(chan MessageChan),
		clients:        make(map[MessageChan]bool),
	}

	// Set it running - listening and broadcasting events
	go broker.listen()

	return
}
