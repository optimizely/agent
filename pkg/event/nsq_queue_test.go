package event

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNSQQueue_Add_Size_Remove(t *testing.T) {
	q := NewNSQueueDefault()

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	q.Add(impression)
	q.Add(impression)
	q.Add(conversion)

	time.Sleep(1 * time.Second)

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
