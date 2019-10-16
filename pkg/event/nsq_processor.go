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
	"fmt"
	"time"

	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/utils"
)

// NewEventProcessorNSQ returns a new instance of BatchEventProcessor with a backing NSQ queue
func NewEventProcessorNSQ(exeCtx utils.ExecutionCtx, queueSize int, flushInterval time.Duration) (*event.BatchEventProcessor, error) {
	q, err := NewNSQueueDefault()
	if err != nil {
		return nil, fmt.Errorf("error creating NSQ event processor: %v", err)
	}

	p := event.NewBatchEventProcessor(event.WithBatchSize(event.DefaultBatchSize), event.WithQueueSize(queueSize),
		event.WithFlushInterval(flushInterval), event.WithQueue(q), event.WithEventDispatcher(&event.HTTPEventDispatcher{}))
	p.Start(exeCtx)

	go func() {
		<-exeCtx.GetContext().Done()
		// if there is an embedded nsqd, tell it to shutdown.
		done <- true
	}()
	return p, nil
}
