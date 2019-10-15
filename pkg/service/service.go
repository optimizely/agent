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

// Package service //
package service

import (
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
)

// Service has generic functionality for service: it starts the service and performs basic checks
type Service struct {
	active  bool
	enabled bool
	port    string
	name    string

	router *chi.Mux
	wg     *sync.WaitGroup
}

// NewService initializes new service
func NewService(enabled bool, port, name string, router *chi.Mux, wg *sync.WaitGroup) *Service {
	return &Service{
		enabled: enabled,
		port:    port,
		name:    name,
		router:  router,
		wg:      wg,
	}
}

// IsHealthy checks if the service is healthy
func (s Service) IsHealthy() (_ bool, _ string) {
	if s.enabled && !s.active {
		return s.active, s.name + " service down"
	}
	return true, ""
}

// StartService starts the service
func (s *Service) StartService() {
	if s.enabled {
		s.wg.Add(1)
		go func() {
			log.Printf("Optimizely " + s.name + " server starting at port " + s.port)
			s.updateState(true)
			if err := http.ListenAndServe(":"+s.port, s.router); err != nil {
				s.updateState(false)
				s.wg.Done()
				log.Error().Msg("Failed to start Optimizely " + s.name + " server.")
			}
		}()
	} else {
		log.Printf("Optimizely " + s.name + " server has not started")
	}
}

func (s *Service) updateState(state bool) {
	s.active = state
}
