package cluster

import "github.com/rs/zerolog/log"

// Delegate is the interface that clients must implement if they want to hook
// into the gossip layer of Memberlist. All the methods must be thread-safe,
// as they can and generally will be called concurrently.
type delegate struct{}

const headerLen = 1

type listener = func([]byte)

var listeners = make(map[string]listener)

// Listen registers listener functions on broadcast messages.
func Listen(header string, listener func([]byte)) {
	if len(header) != headerLen {
		log.Warn().Msgf("Invalid message header %s. Should be length %d", header, headerLen)
		return
	}

	lock.Lock()
	defer lock.Unlock()
	log.Info().Msgf("Adding broadcast listener with header: %s", header)
	listeners[header] = listener
}

// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
func (d delegate) NodeMeta(limit int) []byte {
	return []byte{}
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed
func (d *delegate) NotifyMsg(b []byte) {
	if len(b) == 0 {
		return
	}

	if fun, ok := listeners[string(b[:headerLen])]; ok {
		fun(b[headerLen:])
	}
}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit. Care should be taken that this method does not block,
// since doing so would block the entire UDP packet receive loop.
func (d delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return queue.GetBroadcasts(overhead, limit)
}

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (d *delegate) LocalState(join bool) []byte {
	return []byte{}
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (d *delegate) MergeRemoteState(buf []byte, join bool) {
	// TODO implement me
}
