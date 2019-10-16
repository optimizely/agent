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

import snsq "github.com/segmentio/nsq-go"

// NSQProducer and NSQConsumer are abstractions of snsq.Producer and snsq.Consumer.
// These allow us to test our code in isolation from real NSQ producers & consumers

// NSQProducer is an abstraction of snsq.Producer
type NSQProducer interface {
	Requests() chan<- snsq.ProducerRequest
}

// NSQConsumer is an abstraction of snsq.Consumer
type NSQConsumer interface {
	Messages() <-chan snsq.Message
}
