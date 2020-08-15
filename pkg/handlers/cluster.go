// ****************************************************************************
// * Copyright 2020, Optimizely, Inc. and contributors                   *
// *                                                                          *
// * Licensed under the Apache License, Version 2.0 (the "License");          *
// * you may not use this file except in compliance with the License.         *
// * You may obtain a copy of the License at                                  *
// *                                                                          *
// *    http://www.apache.org/licenses/LICENSE-2.0                            *
// *                                                                          *
// * Unless required by applicable law or agreed to in writing, software      *
// * distributed under the License is distributed on an "AS IS" BASIS,        *
// * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
// * See the License for the specific language governing permissions and      *
// * limitations under the License.                                           *
// ***************************************************************************/

// Package handlers //
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"

	"github.com/optimizely/agent/pkg/cluster"
	"github.com/optimizely/agent/pkg/optimizely"
)

// ClusterInfo contains the current state along with a list of all the cluster members
type ClusterInfo struct {
	Nodes []NodeInfo       `json:"nodes"`
	State optimizely.State `json:"state"`
}

// NodeInfo includes the details for a Agent node
type NodeInfo struct {
	Name      string            `json:"name"`
	Host      string            `json:"host"`
	Servers   map[string]string `json:"servers"`
	LocalNode bool              `json:"localNode"`
}

// GetClusterInfo returns the cluster detail for this node
func GetClusterInfo(w http.ResponseWriter, r *http.Request) {
	nodes := cluster.ListNodes()
	nodeInfo := make([]NodeInfo, len(nodes))
	localNode := cluster.LocalNode()
	for i, node := range cluster.ListNodes() {
		nodeInfo[i] = NodeInfo{
			Name:      node.Name,
			Host:      node.Addr.String(),
			LocalNode: node == localNode,
		}

		meta := &cluster.NodeMeta{}
		if err := json.Unmarshal(node.Meta, meta); err == nil {
			nodeInfo[i].Servers = meta.Servers
		} else {
			log.Warn().Err(err).Msg("cannot parse node meta")
		}

	}

	render.JSON(w, r, &ClusterInfo{Nodes: nodeInfo, State: optimizely.LocalState()})
}
