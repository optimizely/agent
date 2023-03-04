// Package optimizelytest //
package optimizelytest

import (
	"errors"

	"github.com/optimizely/go-sdk/pkg/odp/event"
)

// ErrNotImplemented indicates that the error was returned since the functionality was not implemented
var ErrSendingOdpEvent = errors.New("Error with sending odp event")

// TestEventProcessor implements an Optimizely Processor to aid in testing
type TestEventApiManager struct {
	events []event.Event
}

func (e *TestEventApiManager) SendOdpEvents(apiKey, apiHost string, events []event.Event) (canRetry bool, err error) {
	e.events = append(e.events, events...)
	return false, ErrSendingOdpEvent
}

// GetEvents returns a copy of the events received by the TestEventApiManager
func (e *TestEventApiManager) GetEvents() []event.Event {
	c := make([]event.Event, len(e.events))
	copy(c, e.events)

	return c
}
