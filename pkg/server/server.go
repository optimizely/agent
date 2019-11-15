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

// Package server provides a basic HTTP server wrapper
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Server has generic functionality for service: it starts the service and performs basic checks
type Server struct {
	srv    *http.Server
	logger zerolog.Logger
}

// NewServer initializes new service.
// Configuration is pulled from viper configuration.
func NewServer(name string, handler http.Handler) (Server, error) {
	if !viper.GetBool(name + ".enabled") {
		return Server{}, fmt.Errorf(`"%s" not enabled`, name)
	}

	port := viper.GetString(name + ".port")
	rto := viper.GetDuration("server.readtimeout")
	wto := viper.GetDuration("server.writetimeout")

	logger := log.With().Str("port", port).Str("name", name).Logger()
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  rto,
		WriteTimeout: wto,
	}

	return Server{srv: srv, logger: logger}, nil
}

// ListenAndServe starts the server
func (s Server) ListenAndServe() error {
	s.logger.Info().Msg("Starting server.")
	err := s.srv.ListenAndServe()

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
