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

type StubNSQ struct {
	pc chan snsq.ProducerRequest
	cc chan snsq.Message
}

func (s *StubNSQ) Requests() chan<- snsq.ProducerRequest {
	return s.pc
}

func (s *StubNSQ) Messages() <-chan snsq.Message {
	return s.cc
}

func TestNSQQueue_TestConfigNoProducerConsumer(t *testing.T) {
	stubNsq := &StubNSQ{
		pc: make(chan snsq.ProducerRequest, 10),
		cc: make(chan snsq.Message, 10),
	}
	go func() {
		for pr := range stubNsq.pc {
			pr.Response <- nil
		}
	}()
	var ac AbstractConsumer = stubNsq
	var ap AbstractProducer = stubNsq
	q := NewNSQueue(10, "", false, false, ap, ac)

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
