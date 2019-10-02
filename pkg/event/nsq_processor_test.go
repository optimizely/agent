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
	"testing"
	"time"

	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/optimizely/go-sdk/optimizely/utils"
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

type MockDispatcher struct {
	Events []event.LogEvent
}

func (f *MockDispatcher) DispatchEvent(event event.LogEvent) (bool, error) {
	f.Events = append(f.Events, event)
	return true, nil
}

func TestNSQEventProcessor_ProcessBatch(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtx()
	processor := NewEventProcessorNSQ(exeCtx, 10, 100)
	processor.EventDispatcher = &MockDispatcher{}

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.NotNil(t, processor.Ticker)

	time.Sleep(1 * time.Second)

	assert.Equal(t, 0, processor.EventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 1, len(result.Events))
		evs := result.Events[0]
		assert.True(t, len(evs.Event.Visitors) >= 1)
	}

	exeCtx.CancelFunc()

	time.Sleep(1 * time.Second)
}
