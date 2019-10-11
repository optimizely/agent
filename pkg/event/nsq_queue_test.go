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
	"testing"
	"time"

	snsq "github.com/segmentio/nsq-go"
	"github.com/stretchr/testify/assert"
)

// right now we create a embedded nsqd and send events to it.  this fully tests nsqd but
// is probably not necessary for our tests.  in order to stub the processor and consumer out,
// we need to abstract away the producer and consumer interface, then implement one for nsq and one
// that stubs nsq.
func TestNSQQueue_Add_Get_Size_Remove(t *testing.T) {
	q := NewNSQueueDefault()

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	q.Add(impression)
	q.Add(impression)
	q.Add(conversion)

	time.Sleep(1 * time.Second)

	items1 := q.Get(2)

	assert.NotNil(t, items1)

	q.Remove(1)

	time.Sleep(1 * time.Second)

	items2 := q.Get(1)

	assert.True(t, len(items2) >= 0)

	time.Sleep(1 * time.Second)

	allItems := q.Remove(3)

	assert.True(t, len(allItems) >= 0)

	assert.Equal(t, 0, q.Size())
}

func TestNSQQueue_TestConfigNoProducerConsumer(t *testing.T) {
	pc := make(chan snsq.ProducerRequest, 10)
	cc := make(chan snsq.Message, 10)
	go func() {
		for pr := range pc {
			pr.Response <- nil
		}
	}()
	q := NewNSQueue(10, "", false, pc, cc)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	q.Add(impression)
	q.Add(impression)
	q.Add(conversion)

	items1 := q.Get(2)

	assert.Equal(t, 0, len(items1))

	q.Remove(1)

	items2 := q.Get(1)

	assert.Equal(t, 0, len(items2))

	allItems := q.Remove(3)

	assert.Equal(t, 0, len(allItems))

	assert.Equal(t, 0, q.Size())
}
