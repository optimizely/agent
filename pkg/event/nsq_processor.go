package event

import (
	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/optimizely/go-sdk/optimizely/utils"
	"time"
)

// NewEventProcessor returns a new instance of QueueingEventProcessor with queueSize and flushInterval
func NewEventProcessorNSQ(exeCtx utils.ExecutionCtx, queueSize int, flushInterval time.Duration ) event.Processor {
	p := &event.QueueingEventProcessor{MaxQueueSize: queueSize, FlushInterval:flushInterval,
		Q:NewNSQueueDefault(), EventDispatcher:&event.HTTPEventDispatcher{}, }
	// can't set the wg since it is private (lowercase).
	p.BatchSize = 10
	p.StartTicker(exeCtx.GetContext())
	return p
}

