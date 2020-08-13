package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"

	"github.com/optimizely/agent/pkg/cluster"
)

type ClusterInfo struct {
	Nodes []NodeInfo `json:"nodes"`
}

type NodeInfo struct {
	Name    string            `json:"name"`
	Host    string            `json:"host"`
	Servers map[string]string `json:"servers"`
}

func GetClusterInfo(w http.ResponseWriter, r *http.Request) {
	nodes := cluster.ListNodes()
	nodeInfo := make([]NodeInfo, len(nodes))
	for i, node := range cluster.ListNodes() {
		nodeInfo[i] = NodeInfo{
			Name: node.Name,
			Host: node.Addr.String(),
		}

		meta := &cluster.NodeMeta{}
		if err := json.Unmarshal(node.Meta, meta); err == nil {
			nodeInfo[i].Servers = meta.Servers
		} else {
			log.Warn().Err(err).Msg("cannot parse node meta")
		}

	}

	render.JSON(w, r, &ClusterInfo{Nodes: nodeInfo})
}
