package cluster

import (
	"fmt"
	"os"

	"github.com/optimizely/agent/config"

	"github.com/google/uuid"
	"github.com/hashicorp/memberlist"
	"github.com/rs/zerolog/log"
)

// Reference https://github.com/asim/memberlist/blob/master/memberlist.go

var (
	ml    *memberlist.Memberlist
	queue *memberlist.TransmitLimitedQueue
)

// Init intializes the cluster via the cluster config
func Init(conf config.ClusterConfig) error {
	hostname, _ := os.Hostname()
	c := memberlist.DefaultLocalConfig()
	c.Events = &eventDelegate{}
	c.Delegate = &delegate{}
	c.BindAddr = conf.Host
	c.BindPort = conf.Port
	c.Name = hostname + "-" + uuid.New().String()
	//c.LogOutput = log.Logger

	var err error
	ml, err = memberlist.Create(c)
	if err != nil {
		return err
	}

	_, err = ml.Join(conf.Nodes)
	if err != nil {
		return err
	}

	queue = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return ml.NumMembers()
		},
		RetransmitMult: 3,
	}
	node := ml.LocalNode()
	log.Info().Msgf("Local member %s:%d", node.Addr, node.Port)
	return nil
}

// Broadcast message to all members of the cluster.
func Broadcast(header []byte, buf []byte) error {
	if len(header) != headerLen {
		return fmt.Errorf("invalid header, must be of length %d not %d", headerLen, len(header))
	}

	queue.QueueBroadcast(&broadcast{
		msg:    append(header, buf...),
		notify: nil,
	})

	return nil
}

// Listen registers listener functions on broadcast messages.
func Listen(header []byte, listener func([]byte)) {

}
