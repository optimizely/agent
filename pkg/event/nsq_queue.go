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
	"errors"
	"fmt"
	"time"

	// this is the nsq demon
	"github.com/nsqio/nsq/nsqd"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/rs/zerolog/log"

	// we use the Consumer from segmentio as it mentions in the segmentio documents, they
	// found several problems with the nsq Consumer and ended up creating their own wrapper.
	// we use that wrapper.
	snsq "github.com/segmentio/nsq-go"
)

// NsqConsumerChannel is the default Consumer channel
const NsqConsumerChannel string = "optimizely"

// NsqListenSpec is the default NSQD address
const NsqListenSpec string = "localhost:4150"

// NsqTopic is the default NSQ topic
const NsqTopic string = "user_event"

const deadline = 5 * time.Second

var embeddedNSQD *nsqd.NSQD

// the done channel is used by the embeddedNSQD.  The processor implementation waits for a context.Done()
// and then calls done to shutdown the embeddedNSQD if it is running.
var done = make(chan bool)

// NSQQueue is a implementation of Queue interface for use with NSQ
// When constructed with NewNSQueue, it can be used in three modes:
// - Consumer-only
// - Producer-only
// - both Producer & Consumer
// In Consumer-only mode, producerWriteChan is nil, and the Add method has no effect.
// In Producer-only and "both" modes, Add writes messages to producerWriteChan.
// In Consumer-only and "both" modes, NewNSQueue pulls messages from the Consumer channel and
// adds them to consumerMessages. The other Queue interface methods interact with consumerMessages.
type NSQQueue struct {
	Producer         NSQProducer
	Consumer         NSQConsumer
	consumerMessages event.Queue
}

// Get returns queue for given count size
func (q *NSQQueue) Get(count int) []interface{} {

	var (
		events = make([]interface{}, 0)
	)

	messages := q.consumerMessages.Get(count)
	for _, message := range messages {
		mess, ok := message.(snsq.Message)
		if !ok {
			continue
		}
		events = append(events, q.decodeMessage(mess.Body))
	}

	return events
}

// Add appends item to queue
func (q *NSQQueue) Add(item interface{}) {
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

	// If Producer is nil, this queue is in Consumer-only mode
	if q.Producer != nil {
		response := make(chan error, 1)
		deadline := time.Now().Add(deadline)

		// Attempts to queue the request so one of the active connections can pick
		// it up.
		q.Producer.Requests() <- snsq.ProducerRequest{
			Topic:    NsqTopic,
			Message:  buf.Bytes(),
			Response: response,
			Deadline: deadline,
		}
		// This will always trigger, either if the connection was lost or if a
		// response was successfully sent.
		err = <-response

		if err != nil {
			log.Error().Err(err).Msg("Error publishing event")
		}
	} else {
		log.Error().Msg("No publisher present")
	}
}

func (q *NSQQueue) decodeMessage(body []byte) event.UserEvent {
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
func (q *NSQQueue) Remove(count int) []interface{} {
	userEvents := make([]interface{}, 0, count)
	events := q.consumerMessages.Remove(count)
	for _, message := range events {
		mess, ok := message.(snsq.Message)
		if !ok {
			log.Error().Msg("Message found in consumerMessages was not a snsq.Message")
			continue
		}
		userEvent := q.decodeMessage(mess.Body)
		mess.Finish()
		userEvents = append(userEvents, userEvent)
	}
	return userEvents
}

// Size returns size of queue
func (q *NSQQueue) Size() int {
	return q.consumerMessages.Size()
}

// NewNSQueue returns a NSQQueue connected to the argument Producer & Consumer
// NSQQueue can operate in 3 modes:
// - Producer-only
// - Consumer-only
// - both Consumer & Producer
// To create an instance in Consumer-only mode, pass nil for the Producer argument.
// To create an instance in Producer-only mode, pass nil for the Consumer argument.
// For "both" mode, provide both Producer and Consumer
func NewNSQueue(queueSize int, startDaemon bool, producer NSQProducer, consumer NSQConsumer) (event.Queue, error) {
	if producer == nil && consumer == nil {
		return nil, errors.New("invalid arguments: must provide at least one of Producer or Consumer")
	}

	// Run NSQD embedded
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
					err = embeddedNSQD.Main()
					if err != nil {
						log.Error().Err(err).Msg("error calling Main() on embeddedNSQD")
					}
					// wait until we are told to continue and exit
					<-done
					embeddedNSQD.Exit()
					embeddedNSQD = nil
				}
			}
		}()
	}

	i := &NSQQueue{Producer: producer, Consumer: consumer, consumerMessages: event.NewInMemoryQueue(queueSize)}

	if consumer != nil {
		go func() {
			for message := range i.Consumer.Messages() {
				i.consumerMessages.Add(message)
			}
		}()
	}

	return i, nil
}

// NewNSQueueDefault returns a default implementation of the NSQueue
func NewNSQueueDefault() (event.Queue, error) {
	nsqConfig := snsq.ProducerConfig{Address: NsqListenSpec, Topic: NsqTopic}
	producer, err := snsq.NewProducer(nsqConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating default NSQQueue: %v", err)
	}
	producer.Start()

	consumer, err := snsq.StartConsumer(snsq.ConsumerConfig{
		Topic:       NsqTopic,
		Channel:     NsqConsumerChannel,
		Address:     NsqListenSpec,
		MaxInFlight: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating default NSQQueue: %v", err)
	}

	q, err := NewNSQueue(100, true, producer, consumer)
	if err != nil {
		return nil, fmt.Errorf("error creating default NSQQueue: %v", err)
	}

	return q, nil
}
