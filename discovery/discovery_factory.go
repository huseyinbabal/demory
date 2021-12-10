package discovery

import (
	"log"

	"github.com/hashicorp/raft"
	"github.com/huseyinbabal/demory/node"
)

func Get(nodeConfig *node.Config, raft *raft.Raft) Discovery {
	switch nodeConfig.DiscoveryStrategy {
	case StrategyPort:
		return NewPortDiscovery(nodeConfig.MinPort, nodeConfig.MaxPort, "localhost", nodeConfig.NodeAddress, nodeConfig.NodeID, raft)
	case StrategyKubernetes:
		return NewKubernetesDiscovery(nodeConfig.KubernetesNamespace, nodeConfig.KubernetesService, nodeConfig.NodeAddress,
			nodeConfig.NodeID, raft)
	default:
		log.Fatalf("invalid discovery %s", nodeConfig.DiscoveryStrategy)
		return nil
	}
}
