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

// Package event //
package event

import (
	"time"

	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/optimizely/go-sdk/optimizely/utils"
)

// NewEventProcessorNSQ returns a new instance of QueueingEventProcessor with a backing NSQ queue
func NewEventProcessorNSQ(exeCtx utils.ExecutionCtx, queueSize int, flushInterval time.Duration) (*event.QueueingEventProcessor, error) {
	q, err := NewNSQueueDefault()
	if err != nil {
		return nil, err
	}

	// can't set the wg since it is private (lowercase).
	p := event.NewEventProcessor(exeCtx, event.BatchSize(event.DefaultBatchSize), event.QueueSize(queueSize),
		event.FlushInterval(flushInterval), event.PQ(q), event.PDispatcher(&event.HTTPEventDispatcher{}))

	go func() {
		<-exeCtx.GetContext().Done()
		// if there is an embedded nsqd, tell it to shutdown.
		done <- true
	}()
	return p, nil
}
