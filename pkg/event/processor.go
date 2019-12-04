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
	"encoding/json"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/rs/zerolog/log"
	snsq "github.com/segmentio/nsq-go"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

const jsonContentType = "application/json"

const timeout = 5 * time.Second

// OptlyEventProcessorConfig represents configuration of the event processor. Also configuring nsq if used.
type OptlyEventProcessorConfig struct {
	NSQWithProducer     bool                    `yaml:"nsqWithProducer" default:"false"`
	NSQWithConsumer     bool                    `yaml:"nsqWithConsumer" default:"false"`
	NSQEnabled          bool                    `yaml:"nsqEnabled" default:"false"`
	NSQStartEmbedded    bool                    `yaml:"nsqStartEmbedded" default:"false"`
	NSQAddress          string                  `yaml:"nsqAddress" default:"localhost:4150"`
	QueueSize           int                     `yaml:"queueSize" default:"1000"`
	BatchSize           int                     `yaml:"batchSize" default:"10"`
}


// GetOptlyEventProcessor get the optly event processor using viper configuration variables.
func GetOptlyEventProcessor() event.Processor {

	var config OptlyEventProcessorConfig
	if err := viper.UnmarshalKey("optimizely.eventProcessor", &config); err != nil {
		log.Info().Msg("Unable to parse event processor config.")
		return event.NewBatchEventProcessor()
	}

	var q event.Queue

	if config.QueueSize == 0 {
		config.QueueSize = event.DefaultEventQueueSize
	}
	if config.BatchSize == 0 {
		config.BatchSize = event.DefaultBatchSize
	}

	// configure NSQ backed Queue
	if config.NSQEnabled  {
		startEmbedded := config.NSQStartEmbedded
		var nsqAddress string
		if nsqAddress = config.NSQAddress; nsqAddress == "" {
			nsqAddress = NsqListenSpec
		}

		var consumer *snsq.Consumer
		var producer *snsq.Producer

		if config.NSQWithConsumer {
			consumer, _ = snsq.StartConsumer(snsq.ConsumerConfig{
				Topic:       NsqTopic,
				Channel:     NsqConsumerChannel,
				Address:     nsqAddress,
				MaxInFlight: 100,
			})

		}

		if config.NSQWithProducer {
			nsqConfig := snsq.ProducerConfig{Address: nsqAddress, Topic: NsqTopic}
			producer, _ = snsq.NewProducer(nsqConfig)

		}

		q, _ = NewNSQueue(config.QueueSize, startEmbedded, producer, consumer)

	} else {
		// use default in memory queue
		q = event.NewInMemoryQueue(config.QueueSize)
	}

	// return a new batch event processor
	return event.NewBatchEventProcessor(event.WithQueueSize(config.QueueSize), event.WithBatchSize(config.BatchSize), event.WithQueue(q))
}

// SidedoorEventProcessor - sends events to sidedoor API
type SidedoorEventProcessor struct {
	client http.Client
	URL    string
}

// NewSidedoorEventProcessor - Create a SidedoorEventProcessor of the given URL, with a default client that sets a 5 second request timeout
func NewSidedoorEventProcessor(url string) *SidedoorEventProcessor {
	client := http.Client{Timeout: timeout}
	return &SidedoorEventProcessor{
		client: client,
		URL:    url,
	}
}

// ProcessEvent - send event to sidedoor API
func (s *SidedoorEventProcessor) ProcessEvent(userEvent event.UserEvent) error {
	jsonValue, err := json.Marshal(userEvent)
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling event")
		return err
	}

	resp, err := s.client.Post(s.URL, jsonContentType, bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Error().Err(err).Msg("Error sending request")
		return err
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Warn().Str("URL", s.URL).Err(closeErr).Msg("Error closing response body")
		}
	}()

	return err
}
