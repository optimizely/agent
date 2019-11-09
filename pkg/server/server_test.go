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

package server

import (
	"net/http"
	"sync"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestStartAndShutdown(t *testing.T) {
	viper.SetDefault("valid.enabled", true)
	viper.SetDefault("valid.port", "1000")
	srv, err := NewServer("valid", handler)
	if !assert.NoError(t, err) {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		srv.ListenAndServe()
	}()

	wg.Wait()
	srv.Shutdown()
}

func TestNotEnabled(t *testing.T) {
	_, err := NewServer("not-enabled", handler)
	if assert.Error(t, err) {
		assert.Equal(t, `"not-enabled" not enabled`, err.Error())
	}
}

func TestFailedStartService(t *testing.T) {
	viper.SetDefault("test.enabled", true)
	viper.SetDefault("test.port", "-9")
	ns, err := NewServer("test", handler)
	assert.NoError(t, err)
	ns.ListenAndServe()
}
