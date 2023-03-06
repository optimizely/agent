// Package optimizelytest //
package optimizelytest

import (
	"errors"
	"sync"

	"github.com/optimizely/go-sdk/pkg/odp/event"
)

// TestEventAPIManager implements Event API Manager to aid in testing
type TestEventAPIManager struct {
	events []event.Event
	m      sync.Mutex
}

// SendOdpEvents appends an event to a slice of events and returns a boolean false that retrying didn't take place,
// meaning that event was added to the evnts slice
func (e *TestEventAPIManager) SendOdpEvents(apiKey, apiHost string, events []event.Event) (canRetry bool, err error) {
	e.m.Lock()
	defer e.m.Unlock()
	e.events = append(e.events, events...)
	return false, errors.New("failed to send odp event")
}

// GetEvents returns a copy of the events received by the TestEventApiManager
func (e *TestEventAPIManager) GetEvents() []event.Event {
	e.m.Lock()
	defer e.m.Unlock()
	c := make([]event.Event, len(e.events))
	copy(c, e.events)

	return c
}
