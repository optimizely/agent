/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
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
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/optimizely/agent/config"
	plugins "github.com/optimizely/agent/plugins/middleware"

	"github.com/stretchr/testify/assert"
)

var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

var conf = config.ServerConfig{}

func TestStartAndShutdown(t *testing.T) {
	srv, err := NewServer("valid", "6000", handler, conf)
	if !assert.NoError(t, err) {
		return
	}

	start := make(chan struct{})
	finish := make(chan error)
	go func() {
		close(start)
		finish <- srv.ListenAndServe()
	}()

	<-start
	srv.Shutdown()
	assert.NoError(t, <-finish)
}

func TestNoHandler(t *testing.T) {
	ns, err := NewServer("test", "0", nil, conf)
	assert.Error(t, err)
	assert.Equal(t, ns, Server{})
}

func TestNotEnabledServer(t *testing.T) {
	_, err := NewServer("not-enabled", "0", handler, conf)
	assert.NoError(t, err) // this is checked in server group
}

func TestFailedStartService(t *testing.T) {
	ns, err := NewServer("test", "-9", handler, conf)
	assert.NoError(t, err)
	ns.ListenAndServe()
}

func TestFailedTLSStartService(t *testing.T) {
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

func TestTLSServerConfigs(t *testing.T) {
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

	defaultCiphers := []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	}

	blacklist := []string{
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
	}

	ciphers := blacklistCiphers(blacklist, defaultCiphers)

	expectedCiphers := []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	}
	assert.Equal(t, expectedCiphers, ciphers)

}

func TestHealthMW(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "health status api failed")
	})
	mw := healthMW(nextHandler, "/health")
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, "Status code differs")
	expected := string(`{"status":"ok"}`)
	assert.JSONEq(t, expected, rec.Body.String(), "Response body differs")
}

func TestNewServerHandlerRejectsInvalidHost(t *testing.T) {
	confWithAllowedHosts := config.ServerConfig{
		AllowedHosts:    []string{"example.com"},
		HealthCheckPath: "/health",
		Host:            "127.0.0.1",
	}
	srv, err := NewServer("valid_hosts", "1000", handler, confWithAllowedHosts)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "http://evil.com:1000/v1/config", nil)
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	req = httptest.NewRequest("GET", "http://evil.com/v1/config", nil)
	rec = httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestNewServerHandlerAllowsValidHost(t *testing.T) {
	confWithAllowedHosts := config.ServerConfig{
		AllowedHosts:    []string{"example.com"},
		HealthCheckPath: "/health",
		Host:            "127.0.0.1",
	}
	srv, err := NewServer("valid_hosts", "1000", handler, confWithAllowedHosts)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "http://example.com:1000/v1/config", nil)
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest("GET", "http://example.com/v1/config", nil)
	rec = httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

type mockHandler struct {
	wg *sync.WaitGroup
}

func (m *mockHandler) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			m.wg.Done()
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func TestWrapHandler(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	handler := func(w http.ResponseWriter, r *http.Request) {
		wg.Done()
	}

	conf := config.PluginConfigs{}
	creator := func() plugins.Plugin {
		return &mockHandler{wg: wg}
	}

	// Add valid plugins
	for i := 0; i < 5; i++ {
		wg.Add(1)
		name := fmt.Sprintf("mock%d", i)
		plugins.Add(name, creator)
		conf[name] = map[string]interface{}{}
	}

	// Test missing plugin
	conf["DNE"] = map[string]interface{}{}

	// Test failed unmarshalling
	plugins.Add("badConf", creator)
	conf["badConf"] = false

	// Test failed marshalling
	plugins.Add("notJSON", creator)
	conf["notJSON"] = make(chan struct{})

	next := wrapHandler(http.HandlerFunc(handler), conf)

	next.ServeHTTP(nil, nil)

	// Ensure all VALID plugins were executed.
	wg.Wait()
}
