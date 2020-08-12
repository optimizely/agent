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