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

package service

import (
	"sync"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	ns := NewService(true, "1", "name", &chi.Mux{}, &sync.WaitGroup{})
	assert.NotNil(t, ns)

	assert.Equal(t, ns.port, "1")
	assert.Equal(t, ns.enabled, true)
	assert.Equal(t, ns.name, "name")
}

func TestUpdateState(t *testing.T) {
	ns := NewService(true, "1", "name", &chi.Mux{}, &sync.WaitGroup{})

	ns.updateState(false)
	state, reason := ns.IsHealthy()
	assert.False(t, state)
	assert.Equal(t, reason, "name service down")

	ns.updateState(true)
	state, reason = ns.IsHealthy()
	assert.True(t, state)
	assert.Equal(t, reason, "")
}

func TestFailedStartService(t *testing.T) {

	var wg sync.WaitGroup
	ns := NewService(true, "-9", "api", &chi.Mux{}, &wg)

	ns.StartService()
	wg.Wait()

	state, reason := ns.IsHealthy()
	assert.False(t, state)
	assert.Equal(t, reason, "api service down")
}
