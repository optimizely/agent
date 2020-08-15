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
	"encoding/json"

	"github.com/hashicorp/memberlist"
)

// Broadcast is something that can be broadcasted via gossip to
// the memberlist cluster.
type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

// Invalidates checks if enqueuing the current broadcast
// invalidates a previous broadcast
func (b *broadcast) Invalidates(other memberlist.Broadcast) bool {
	return false
}

// Returns a byte form of the message
func (b *broadcast) Message() []byte {
	return b.msg
}

// Finished is invoked when the message will no longer
// be broadcast, either due to invalidation or to the
// transmit limit being reached
func (b *broadcast) Finished() {
	if b.notify != nil {
		close(b.notify)
	}
}

// Broadcast is a convenience method for publishing messages.
func Broadcast(header string, request interface{}) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}

	return addToQueue([]byte(header), payload)
}
