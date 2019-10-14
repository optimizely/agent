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

// Service has generic info for service
type Service struct {
	active bool
	port   string
	name   string

	router *chi.Mux
	wg     *sync.WaitGroup
}

// NewService initializes new service
func NewService(active bool, port, name string, router *chi.Mux, wg *sync.WaitGroup) *Service {
	return &Service{
		active: active,
		port:   port,
		name:   name,
		router: router,
		wg:     wg,
	}
}

// IsAlive checks if the service is alive
func (s Service) IsAlive() bool {
	return s.active
}

// StartService starts the service
func (s *Service) StartService() {
	if s.active {
		s.wg.Add(1)
		go func() {
			log.Printf("Optimizely " + s.name + " server starting at port " + s.port)
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
