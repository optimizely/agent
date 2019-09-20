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

