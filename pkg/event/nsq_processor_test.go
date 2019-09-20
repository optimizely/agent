package event

import (
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

func TestNSQEventProcessor_ProcessImpression(t *testing.T) {
	processor := NewEventProcessorNSQ(100, 100)

	impression := BuildTestImpressionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)

	result, ok := processor.(*event.QueueingEventProcessor)

	if ok {
		time.Sleep(3000 * time.Millisecond)

		assert.NotNil(t, result.Ticker)

		assert.Equal(t, 0, result.EventsCount())
	} else {
		assert.Equal(t, true, false)
	}

}

type MockDispatcher struct {
	Events []event.LogEvent
}

func (f *MockDispatcher) DispatchEvent(event event.LogEvent, callback func(success bool)) {
	f.Events = append(f.Events, event)
	callback(true)
}

func TestNSQEventProcessor_ProcessBatch(t *testing.T) {
	processor := &event.QueueingEventProcessor{MaxQueueSize: 100, FlushInterval: 100, Q: NewNSQueueDefault(), EventDispatcher: &MockDispatcher{}}
	processor.BatchSize = 10
	processor.StartTicker()

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	time.Sleep(3000 * time.Millisecond)

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.EventsCount())

	time.Sleep(3000 * time.Millisecond)

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 1, len(result.Events))
		//evs := result.Events[0]
		//assert.True(t, len(evs.event.Visitors) >= 1)
	}
}
