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
	"crypto/tls"
	"github.com/optimizely/agent/config"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

var conf = config.ServerConfig{}

func TestStartAndShutdown(t *testing.T) {
	srv, err := NewServer("valid", "1000", handler, conf)
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
	_, err := NewServer("not-enabled", "0", handler, conf)
	if assert.Error(t, err) {
		assert.Equal(t, `"not-enabled" not enabled`, err.Error())
	}
}

func TestFailedStartService(t *testing.T) {
	ns, err := NewServer("test", "-9", handler, conf)
	assert.NoError(t, err)
	ns.ListenAndServe()
}

func TestFailedTSLStartService(t *testing.T) {
	cfg := config.ServerConfig{
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 8 * time.Second,
		CertFile:     "testdata/example-cert.pem",
		KeyFile:      "testdata/example-key.pem1",
	}
	ns, err := NewServer("test", "9", handler, cfg)
	assert.Error(t, err)
	assert.Equal(t, ns, Server{})
}

func TestServerConfigs(t *testing.T) {
	cfg := config.ServerConfig{
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 8 * time.Second,
	}
	ns, err := NewServer("test", "1000", handler, cfg)
	assert.NoError(t, err)

	assert.Equal(t, cfg.ReadTimeout, ns.srv.ReadTimeout)
	assert.Equal(t, cfg.WriteTimeout, ns.srv.WriteTimeout)
}

func TestTSLServerConfigs(t *testing.T) {
	cfg := config.ServerConfig{
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 8 * time.Second,
		CertFile:     "testdata/example-cert.pem",
		KeyFile:      "testdata/example-key.pem",
	}
	ns, err := NewServer("test", "1000", handler, cfg)
	assert.NoError(t, err)

	assert.Equal(t, cfg.ReadTimeout, ns.srv.ReadTimeout)
	assert.Equal(t, cfg.WriteTimeout, ns.srv.WriteTimeout)
	assert.NotNil(t, ns.srv.TLSConfig)
}

func TestBlacklistCiphers(t *testing.T) {

	blacklist := []string{
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
	}

	ciphers := blacklistCiphers(blacklist)

	expectedCiphers := []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	}
	assert.Equal(t, expectedCiphers, ciphers)

}
