/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
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

// Package optimizelytest //
package optimizelytest

import (
	"errors"

	"github.com/optimizely/go-sdk/v2/pkg/event"
)

// ErrNotImplemented indicates that the error was returned since the functionality was not implemented
var ErrNotImplemented = errors.New("not implemented")

// TestEventProcessor implements an Optimizely Processor to aid in testing
type TestEventProcessor struct {
	events []event.UserEvent
}

// ProcessEvent appends the events to a slice of UserEvents
func (p *TestEventProcessor) ProcessEvent(e event.UserEvent) bool {
	p.events = append(p.events, e)
	return true
}

// GetEvents returns a copy of the events received by the TestEventProcessor
func (p *TestEventProcessor) GetEvents() []event.UserEvent {
	c := make([]event.UserEvent, len(p.events))
	copy(c, p.events)

	return c
}

// OnEventDispatch is a non-op notification action
func (p *TestEventProcessor) OnEventDispatch(callback func(logEvent event.LogEvent)) (int, error) {
	return 0, ErrNotImplemented
}

// RemoveOnEventDispatch is a non-op notification action
func (p *TestEventProcessor) RemoveOnEventDispatch(id int) error {
	return ErrNotImplemented
}
