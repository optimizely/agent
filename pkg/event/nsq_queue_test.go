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

func TestNSQQueue_Add_Get_Size_Remove(t *testing.T) {
	pc := make(chan snsq.ProducerRequest, 10)
	cc := make(chan snsq.Message, 10)
	defer func() {
		close(cc)
		close(pc)
	}()

	go func() {
		for pr := range pc {
			pr.Response <- nil
			cc <- *snsq.NewMessage(0, pr.Message, make(chan snsq.Command, 10))
		}
	}()

	q := NewNSQueue(10, "", false, pc, cc)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	q.Add(impression)
	q.Add(impression)
	q.Add(conversion)

	items1 := q.Get(3)
	assert.Eventually(t, func() bool { return len(items1) == 3 }, 5*time.Second, 10*time.Millisecond)

	q.Remove(1)
	items2 := q.Get(3)
	assert.Equal(t, 2, len(items2))

	allItems := q.Remove(3)
	assert.Equal(t, 2, len(allItems))
	assert.Equal(t, 0, q.Size())
}
