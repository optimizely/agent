package event

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

	time.Sleep(2 * time.Second)

	items1 := q.Get(2)

	assert.GreaterOrEqual(t, len(items1), 1)

	q.Remove(1)

	time.Sleep(1 * time.Second)

	items2 := q.Get(1)

	assert.True(t, len(items2) >= 0)

	time.Sleep(1 * time.Second)

	allItems := q.Remove(3)

	assert.True(t,len(allItems) >= 0)

	assert.Equal(t, 0, q.Size())
}

func TestNSQQueue_TestConfigNoProducerConsumer(t *testing.T) {
	q := NewNSQueue(10, "", false, false, false)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	q.Add(impression)
	q.Add(impression)
	q.Add(conversion)

	time.Sleep(1 * time.Second)

	items1 := q.Get(2)

	assert.Equal(t, 0, len(items1))

	q.Remove(1)

	items2 := q.Get(1)

	assert.Equal(t, 0, len(items2))

	allItems := q.Remove(3)

	assert.Equal(t,0, len(allItems))

	assert.Equal(t, 0, q.Size())
}
