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
	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/optimizely/go-sdk/optimizely/utils"
	"time"
)

// NewEventProcessor returns a new instance of QueueingEventProcessor with queueSize and flushInterval
func NewEventProcessorNSQ(exeCtx utils.ExecutionCtx, queueSize int, flushInterval time.Duration ) *event.QueueingEventProcessor {
	// can't set the wg since it is private (lowercase).
	p := event.NewEventProcessor(exeCtx, event.DefaultBatchSize, queueSize, flushInterval)
	p.Q = NewNSQueueDefault()
	p.EventDispatcher = &event.HTTPEventDispatcher{}

	return p
}

