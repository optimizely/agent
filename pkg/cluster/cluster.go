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
		NumNodes:       ml.NumMembers,
		RetransmitMult: 3,
	}
	node := ml.LocalNode()
	log.Info().Msgf("Local member %s:%d", node.Addr, node.Port)
	return nil
}

// addToQueue message to all members of the cluster.
func addToQueue(header, buf []byte) error {
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

// ListNodes returns an array of all the nodes in the cluster
func ListNodes() []*memberlist.Node {
	return ml.Members()
}

// LocalNode returns the Node detail for the local node
func LocalNode() *memberlist.Node {
	return ml.LocalNode()
}
