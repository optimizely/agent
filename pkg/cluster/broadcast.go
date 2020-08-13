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

// Broadcast... should be replaced with NamedBroadcast
func Broadcast(header string, request interface{}) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}

	return addToQueue([]byte(header), payload)
}
