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

package event

import (
	"bytes"
	"encoding/gob"
	"testing"
	"time"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
	snsq "github.com/segmentio/nsq-go"
	"github.com/stretchr/testify/assert"
)

func BuildTestImpressionEvent() event.UserEvent {
	config := TestConfig{}

	experiment := entities.Experiment{}
	experiment.Key = "background_experiment"
	experiment.LayerID = "15399420423"
	experiment.ID = "15402980349"

	variation := entities.Variation{}
	variation.Key = "variation_a"
	variation.ID = "15410990633"

	impressionUserEvent := event.CreateImpressionUserEvent(config, experiment, variation, userContext)

	return impressionUserEvent
}

func BuildTestConversionEvent() event.UserEvent {
	config := TestConfig{}
	conversionUserEvent := event.CreateConversionUserEvent(config, entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, userContext, make(map[string]interface{}))

	return conversionUserEvent
}

type MockProducer struct {
	ProducerChannel chan snsq.ProducerRequest
}

func (p *MockProducer) Requests() chan<- snsq.ProducerRequest {
	return p.ProducerChannel
}

type MockConsumer struct {
	ConsumerChannel chan snsq.Message
}

func (c *MockConsumer) Messages() <-chan snsq.Message {
	return c.ConsumerChannel
}

func NewMockProducerAndConsumer() (*MockProducer, *MockConsumer) {
	pc := make(chan snsq.ProducerRequest, 10)
	cc := make(chan snsq.Message, 10)
	go func() {
		for pr := range pc {
			pr.Response <- nil
			cc <- *snsq.NewMessage(0, pr.Message, make(chan snsq.Command, 10))
		}
	}()
	mp := &MockProducer{ProducerChannel: pc}
	mc := &MockConsumer{ConsumerChannel: cc}
	return mp, mc
}

func TestNSQQueue_Add_Get_Size_Remove(t *testing.T) {
	mp, mc := NewMockProducerAndConsumer()
	defer func() {
		close(mp.ProducerChannel)
		close(mc.ConsumerChannel)
	}()

	q, err := NewNSQueue(10, "", false, mp, mc)
	assert.NoError(t, err)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	q.Add(impression)
	q.Add(impression)
	q.Add(conversion)

	assert.Eventually(t, func() bool { return len(q.Get(3)) == 3 }, 5*time.Second, 10*time.Millisecond)

	q.Remove(1)
	items2 := q.Get(3)
	assert.Equal(t, 2, len(items2))

	allItems := q.Remove(3)
	assert.Equal(t, 2, len(allItems))
	assert.Equal(t, 0, q.Size())
}

func TestNSQQueue_Consumer_Only(t *testing.T) {
	mp, mc := NewMockProducerAndConsumer()
	defer func() {
		close(mp.ProducerChannel)
		close(mc.ConsumerChannel)
	}()

	// Pass nil Producer for Consumer-only
	q, err := NewNSQueue(10, "", false, nil, mc)
	assert.NoError(t, err)

	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err = enc.Encode(BuildTestImpressionEvent())
	mp.Requests() <- snsq.ProducerRequest{
		Topic:    NsqTopic,
		Message:  buf.Bytes(),
		Response: make(chan error, 1),
		Deadline: time.Now().Add(2 * time.Second),
	}

	assert.Eventually(t, func() bool { return len(q.Get(1)) == 1 }, 5*time.Second, 10*time.Millisecond)

	q.Add(BuildTestConversionEvent())
}

func TestNSQQueue_Producer_Only(t *testing.T) {
	mp, mc := NewMockProducerAndConsumer()
	defer func() {
		close(mp.ProducerChannel)
		close(mc.ConsumerChannel)
	}()

	// Pass nil Consumer for Producer-only
	q, err := NewNSQueue(10, "", false, mp, nil)
	assert.NoError(t, err)

	q.Add(BuildTestConversionEvent())
	assert.Eventually(t, func() bool {
		select {
		case <-mc.ConsumerChannel:
			return true
		default:
			return false
		}
	}, 5*time.Second, 10*time.Millisecond)
}
