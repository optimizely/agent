package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"

	"github.com/optimizely/agent/pkg/cluster"
	"github.com/optimizely/agent/pkg/optimizely"
)

type ClusterInfo struct {
	Nodes []NodeInfo       `json:"nodes"`
	State optimizely.State `json:"state"`
}

type NodeInfo struct {
	Name      string            `json:"name"`
	Host      string            `json:"host"`
	Servers   map[string]string `json:"servers"`
	LocalNode bool              `json:"localNode"`
}

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
