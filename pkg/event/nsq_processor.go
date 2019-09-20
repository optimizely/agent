package event

import (
	"github.com/optimizely/go-sdk/optimizely/event"
	"time"
)

// NewEventProcessor returns a new instance of QueueingEventProcessor with queueSize and flushInterval
func NewEventProcessorNSQ(queueSize int, flushInterval time.Duration ) event.Processor {
	p := &event.QueueingEventProcessor{MaxQueueSize: queueSize, FlushInterval:flushInterval, Q:NewNSQueueDefault(), EventDispatcher:&event.HTTPEventDispatcher{}}
	p.BatchSize = 10
	p.StartTicker()
	return p
}

