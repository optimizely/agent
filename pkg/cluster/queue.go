package cluster

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/optimizely/agent/config"

	"github.com/google/uuid"
	"github.com/hashicorp/memberlist"
	"github.com/rs/zerolog/log"
)

// Reference https://github.com/asim/memberlist/blob/master/memberlist.go

var (
	ml    *memberlist.Memberlist
	queue *memberlist.TransmitLimitedQueue
	lock  = sync.RWMutex{}
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
	c.LogOutput = ioutil.Discard

	var err error
	ml, err = memberlist.Create(c)
	if err != nil {
		return err
	}

	// Attempt to connect to other nodes in the cluster
	if _, err := ml.Join(conf.Nodes); err != nil {
		log.Warn().Err(err).Msg("No nodes were joined. This is likely the first node in the cluster.")
	}

	queue = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return ml.NumMembers()
		},
		RetransmitMult: 1,
	}
	node := ml.LocalNode()
	log.Info().Msgf("Local member %s:%d", node.Addr, node.Port)
	return nil
}

// addToQueue message to all members of the cluster.
func addToQueue(header []byte, buf []byte) error {
	if len(header) != headerLen {
		return fmt.Errorf("invalid header, must be of length %d not %d", headerLen, len(header))
	}

	if queue == nil {
		return errors.New("cluster not configured")
	}

	queue.QueueBroadcast(&broadcast{
		msg:    append(header, buf...),
		notify: nil,
	})

	return nil
}
