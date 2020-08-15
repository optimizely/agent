/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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

// Package cluster provides basic cluster functionality
package cluster

import (
	"github.com/hashicorp/memberlist"
	"github.com/rs/zerolog/log"
)

// EventDelegate is a simpler delegate that is used only to receive
// notifications about members joining and leaving. The methods in this
// delegate may be called by multiple goroutines, but never concurrently.
// This allows you to reason about ordering.
type eventDelegate struct{}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (ed *eventDelegate) NotifyJoin(node *memberlist.Node) {
	log.Info().Msgf("A node has joined: %s", node.String())
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (ed *eventDelegate) NotifyLeave(node *memberlist.Node) {
	log.Info().Msgf("A node has left: %s", node.String())
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (ed *eventDelegate) NotifyUpdate(node *memberlist.Node) {
	log.Info().Msgf("A node was updated: %s", node.String())
}
