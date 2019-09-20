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
	"bytes"
	"encoding/gob"
	"github.com/nsqio/go-nsq"
	"github.com/nsqio/nsq/nsqd"
	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/rs/zerolog/log"
	snsq "github.com/segmentio/nsq-go"
)

const NsqConsumerChannel string = "optimizely"
const NsqListenSpec string = "localhost:4150"
const NsqTopic string = "user_event"

var embeddedNSQD *nsqd.NSQD = nil

var done = make(chan bool)

type NSQQueue struct {
	p *nsq.Producer
	c *snsq.Consumer
	messages event.Queue
}

// Get returns queue for given count size
func (i *NSQQueue) Get(count int) []interface{} {

	var (
		events = make([]interface{}, count)
	)

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
	userEvent, ok := item.(event.UserEvent)
	if !ok {
		// cannot add non-user events
		return
	}
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(userEvent)
	if err != nil {
		log.Error().Err(err).Msg("Error encoding event")
	}

	if i.p != nil {
		err := i.p.Publish(NsqTopic, buf.Bytes())
		if err != nil {
			log.Error().Err(err).Msg("Error publishing event")
		}
	}
}

func (i *NSQQueue) decodeMessage(body []byte) event.UserEvent {
	reader := bytes.NewReader(body)
	dec := gob.NewDecoder(reader)
	userEvent := event.UserEvent{}
	err := dec.Decode(&userEvent)
	if err != nil {
		log.Error().Err(err).Msg("Error decoding event")
	}

	return userEvent
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
	if embeddedNSQD == nil && startDaemon {
		go func() {
			// running an NSQD with all of the default options
			// (as if you ran it from the command line with no flags)
			// is literally these three lines of code. the nsqd
			// binary mainly wraps up the handling of command
			// line args and does something similar
			if embeddedNSQD == nil {
				opts := nsqd.NewOptions()
				var err error
				embeddedNSQD, err = nsqd.New(opts)
				if err == nil {
					embeddedNSQD.Main()
					// wait until we are told to continue and exit
					<-done
					embeddedNSQD.Exit()
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
			Topic:       NsqTopic,
			Channel:     NsqConsumerChannel,
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
	return NewNSQueue(100, NsqListenSpec, true, true, true)
}

