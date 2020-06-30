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
	"fmt"
	"sync"

	"github.com/optimizely/agent/config"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

// Group encapsulates managing multiple Server instances
type Group struct {
	stop context.CancelFunc
	eg   *errgroup.Group
	ctx  context.Context
	conf config.ServerConfig
}

// NewGroup creares a new server group.
func NewGroup(ctx context.Context, conf config.ServerConfig) *Group {
	nctx, stop := context.WithCancel(ctx)
	eg, gctx := errgroup.WithContext(nctx)

	return &Group{
		stop: stop,
		eg:   eg,
		ctx:  gctx,
		conf: conf,
	}
}

// GoListenAndServe constructs a NewServer and adds it to the Group.
// Two goroutines are started. One for the http listener and one
// to initiate a graceful shutdown. This method blocks on adding the
// go routines to maintain startup order of each listener.
func (g *Group) GoListenAndServe(name, port string, handler chi.Router) {

	if port == "0" {
		log.Info().Msg(fmt.Sprintf(`"%s" not enabled`, name))
		return
	}

	server, err := NewServer(name, port, handler, g.conf)

	if err != nil {
		log.Error().Err(err).Msg("Failed starting server")
		g.stop()
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	g.eg.Go(func() error {
		wg.Done()
		defer g.stop()
		return server.ListenAndServe()
	})

	// Shutdown on signal
	wg.Add(1)
	g.eg.Go(func() error {
		wg.Done()
		<-g.ctx.Done()
		server.Shutdown()
		return g.ctx.Err()
	})

	wg.Wait()
}

// Wait waits for all servers to complete before returning
func (g *Group) Wait() error {
	return g.eg.Wait()
}
