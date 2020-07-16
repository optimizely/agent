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

// Package server provides a basic HTTP server wrapper
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/middleware"

	"github.com/go-chi/render"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Server has generic functionality for service: it starts the service and performs basic checks
type Server struct {
	srv    *http.Server
	logger zerolog.Logger
}

// HealthInfo is holding info about health checks
type HealthInfo struct {
	Status string `json:"status,omitempty"`
}

// NewServer initializes new service.
// Configuration is pulled from viper configuration.
func NewServer(name, port string, handler http.Handler, conf config.ServerConfig) (Server, error) {

	if handler == nil {
		return Server{}, fmt.Errorf(`"%s" handler is not initialized`, name)
	}

	usingTLS := conf.KeyFile != "" && conf.CertFile != ""

	withAllowedHostsHandler := middleware.AllowedHosts(conf.GetAllowedHosts(), port, usingTLS)(handler)
	withHealthMWhandler := healthMW(withAllowedHostsHandler, conf.HealthCheckPath)
	logger := log.With().Str("port", port).Str("name", name).Str("host", conf.Host).Logger()
	srv := &http.Server{
		Addr:         conf.Host + ":" + port,
		Handler:      withHealthMWhandler,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	}

	if usingTLS {
		cfg, err := makeTLSConfig(conf)
		if err != nil {
			return Server{}, err
		}
		srv.TLSConfig = cfg
	}

	return Server{srv: srv, logger: logger}, nil
}

// ListenAndServe starts the server
func (s Server) ListenAndServe() (err error) {

	if s.srv.TLSConfig != nil {
		s.logger.Info().Msg("Starting TLS server.")
		err = s.srv.ListenAndServeTLS("", "")
	} else {
		s.logger.Info().Msg("Starting server.")
		err = s.srv.ListenAndServe()
	}

	if !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error().Err(err).Msg("Server failed.")
		return err
	}

	return nil
}

// Shutdown server gracefully
func (s Server) Shutdown() {
	s.logger.Info().Msg("Shutting down server.")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Error().Err(err).Msg("Failed shutdown.")
	}
}

func makeTLSConfig(conf config.ServerConfig) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(conf.CertFile, conf.KeyFile)
	if err != nil {
		return nil, err
	}

	defaultCiphers := []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	}

	ciphers := blacklistCiphers(conf.DisabledCiphers, defaultCiphers)

	return &tls.Config{
		PreferServerCipherSuites: true,
		CipherSuites:             ciphers,
		MinVersion:               tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
			tls.CurveP384,
		},
		Certificates: []tls.Certificate{cert},
	}, nil
}

func makeDefaultCiphersMap() map[string]uint16 {

	return map[string]uint16{
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}
}

func blacklistCiphers(blacklist []string, defaultCiphers []uint16) []uint16 {

	defaultCiphersMap := makeDefaultCiphersMap()
	blacklistCiphersMap := map[uint16]struct{}{}
	modifiedCiphers := []uint16{}

	for _, disabledCipher := range blacklist {
		if v, ok := defaultCiphersMap[disabledCipher]; ok {
			blacklistCiphersMap[v] = struct{}{}
		}
	}

	for _, cipher := range defaultCiphers {
		if _, ok := blacklistCiphersMap[cipher]; !ok {
			modifiedCiphers = append(modifiedCiphers, cipher)
		}
	}

	return modifiedCiphers
}

// healthMW intercepts requests for the given path to return a StatusOK.
func healthMW(next http.Handler, path string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && strings.HasSuffix(strings.ToLower(r.URL.Path), path) {
			render.JSON(w, r, HealthInfo{Status: "ok"})
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
