package event

import (
	"bytes"
	"encoding/gob"
	"github.com/nsqio/go-nsq"
	"github.com/nsqio/nsq/nsqd"
	"github.com/optimizely/go-sdk/optimizely/event"
	snsq "github.com/segmentio/nsq-go"
)

const NSQ_CONSUMER_CHANNEL string = "optimizely"
const NSQ_LISTEN_SPEC string = "localhost:4150"
const NSQ_TOPIC string = "user_event"

var embedded_nsqd *nsqd.NSQD = nil

var done = make(chan bool)

type NSQQueue struct {
	p *nsq.Producer
	c *snsq.Consumer
	messages event.Queue
}

// Get returns queue for given count size
func (i *NSQQueue) Get(count int) []interface{} {

	events := []interface{}{}

	messages := i.messages.Get(count)
	for _, message := range messages {
		mess, ok := message.(snsq.Message)
		if !ok {
			continue
		}
		events = append(events, i.decodeMessage(mess.Body))
	}

	return events
}

// Add appends item to queue
func (i *NSQQueue) Add(item interface{}) {
	event, ok := item.(event.UserEvent)
	if !ok {
		// cannot add non-user events
		return
	}
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	enc.Encode(event)
	if i.p != nil {
		i.p.Publish(NSQ_TOPIC, buf.Bytes())
	}
}

func (i *NSQQueue) decodeMessage(body []byte) event.UserEvent {
	reader := bytes.NewReader(body)
	dec := gob.NewDecoder(reader)
	event := event.UserEvent{}
	dec.Decode(&event)

	return event
}

// Remove removes item from queue and returns elements slice
func (i *NSQQueue) Remove(count int) []interface{} {
	userEvents := make([]interface{},0, count)
	events := i.messages.Remove(count)
	for _,message := range events {
		mess, ok := message.(snsq.Message)
		if !ok {
			continue
		}
		userEvent := i.decodeMessage(mess.Body)
		mess.Finish()
		userEvents = append(userEvents, userEvent)
	}
	return userEvents
}

// Size returns size of queue
func (i *NSQQueue) Size() int {
	return i.messages.Size()
}

// NewNSQueue returns new NSQ based queue with given queueSize
func NewNSQueue(queueSize int, address string, startDaemon bool, startProducer bool, startConsumer bool) event.Queue {

	// Run nsqd embedded
	if embedded_nsqd == nil && startDaemon {
		go func() {
			// running an nsqd with all of the default options
			// (as if you ran it from the command line with no flags)
			// is literally these three lines of code. the nsqd
			// binary mainly wraps up the handling of command
			// line args and does something similar
			if embedded_nsqd == nil {
				opts := nsqd.NewOptions()
				embedded_nsqd, err := nsqd.New(opts)
				if err == nil {
					embedded_nsqd.Main()
					// wait until we are told to continue and exit
					<-done
					embedded_nsqd.Exit()
				}
			}
		}()
	}

	var p *nsq.Producer = nil
	var err error = nil
	nsqConfig := nsq.NewConfig()

	if startProducer {
		p, err = nsq.NewProducer(address, nsqConfig)
		if err != nil {
			//log.Fatal(err)
		}
	}

	var consumer *snsq.Consumer = nil
	if startConsumer {
		consumer, _ = snsq.StartConsumer(snsq.ConsumerConfig{
			Topic:       NSQ_TOPIC,
			Channel:     NSQ_CONSUMER_CHANNEL,
			Address:     address,
			MaxInFlight: queueSize,
		})

	}

	i := &NSQQueue{p: p, c: consumer, messages: event.NewInMemoryQueue(queueSize)}

	if startConsumer {
		go func() {
			for message := range i.c.Messages() {
				i.messages.Add(message)
			}
		}()
	}

	return i
}

//NewNSQueueDefault returns a default implementation of the NSQueue
func NewNSQueueDefault() event.Queue {
	return NewNSQueue(100, NSQ_LISTEN_SPEC, true, true, true)
}

