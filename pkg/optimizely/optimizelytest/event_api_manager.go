// Package optimizelytest //
package optimizelytest

import (
	"errors"
	"sync"

	"github.com/optimizely/go-sdk/v2/pkg/odp/event"
)

// TestEventAPIManager implements Event API Manager to aid in testing
type TestEventAPIManager struct {
	events             []event.Event
	mutex              sync.Mutex
	expectedEventCount int
}

// SendOdpEvents appends an event to a slice of events and returns a boolean false that retrying didn't take place,
// meaning that event was added to the events slice
func (e *TestEventAPIManager) SendOdpEvents(apiKey, apiHost string, events []event.Event) (canRetry bool, err error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.events = append(e.events, events...)

	return false, errors.New("failed to send odp event")
}

// GetEvents returns a copy of the events received by the TestEventApiManager
// This method will wait until events are greater or equal to the number of expected events.
func (e *TestEventAPIManager) GetEvents() []event.Event {
	for {
		e.mutex.Lock()
		if len(e.events) >= e.expectedEventCount {
			break
		}
		e.mutex.Unlock()
	}

	c := make([]event.Event, len(e.events))
	copy(c, e.events)
	e.mutex.Unlock()

	return c
}

// SetExpectedNumberEvents sets the expected number of events that we send. Used in expected test result.
func (e *TestEventAPIManager) SetExpectedNumberEvents(num int) {
	e.expectedEventCount = num
}
