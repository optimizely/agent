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

func getNodeMeta(conf *config.AgentConfig) NodeMeta {
	meta := NodeMeta{
		Servers: make(map[string]string),
	}

	scheme := "http://"
	if conf.Server.KeyFile != "" && conf.Server.CertFile != "" {
		scheme = "https://"
	}
	host := scheme + conf.Server.Host + ":"

	meta.Servers["api"] = host + conf.API.Port
	meta.Servers["admin"] = host + conf.Admin.Port
	meta.Servers["webhook"] = host + conf.Webhook.Port

	return meta
}

// Init intializes the cluster via the cluster config
func Init(conf *config.AgentConfig) error {
	hostname, _ := os.Hostname()
	c := memberlist.DefaultLocalConfig()
	c.Events = &eventDelegate{}

	c.Delegate = &delegate{getNodeMeta(conf)}
	c.BindAddr = conf.Cluster.Host
	c.BindPort = conf.Cluster.Port
	c.Name = hostname + "-" + uuid.New().String()
	c.LogOutput = ioutil.Discard

	var err error
	ml, err = memberlist.Create(c)
	if err != nil {
		return err
	}

	// Attempt to connect to other nodes in the cluster
	if _, err := ml.Join(conf.Cluster.Nodes); err != nil {
		log.Warn().Err(err).Msg("No nodes were joined. This is likely the first node in the cluster.")
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

func ListNodes() []*memberlist.Node {
	return ml.Members()
}

func LocalNode() *memberlist.Node {
	return ml.LocalNode()
}
