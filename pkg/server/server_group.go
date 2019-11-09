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
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

// Group encapsulates managing multiple Server instances
type Group struct {
	done context.CancelFunc
	eg   *errgroup.Group
	ctx  context.Context
	wg   sync.WaitGroup
}

// NewGroup creares a new server group.
func NewGroup(ctx context.Context) *Group {
	nctx, done := context.WithCancel(ctx)
	eg, gctx := errgroup.WithContext(nctx)

	return &Group{
		done: done,
		eg:   eg,
		ctx:  gctx,
		wg:   sync.WaitGroup{},
	}
}

// GoListenAndServe constructs a NewServer and adds it to the Group.
// Two goroutines are started. One for the http listener and one
// to initiate a graceful shitdown. This method blocks on adding the
// go routines to maintain startup order.
func (g *Group) GoListenAndServe(name string, handler http.Handler) {
	server, err := NewServer(name, handler)
	if err != nil {
		log.Info().Err(err).Msg("Not starting server")
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	g.eg.Go(func() error {
		wg.Done()
		defer g.done()
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

// Shutdown initiates a graceful shutdown of srevices
func (g *Group) Shutdown() {
	g.done()
}
