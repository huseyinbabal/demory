package main

import (
	"fmt"
	"github.com/Jille/raft-grpc-leader-rpc/leaderhealth"
	transport "github.com/Jille/raft-grpc-transport"
	"github.com/Jille/raftadmin"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"github.com/huseyinbabal/demory"
	proto "github.com/huseyinbabal/demory-proto/golang/demory"
	"github.com/huseyinbabal/demory/discovery"
	"github.com/huseyinbabal/demory/node"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"path/filepath"
)

func main() {
	nodeConfig, nodeConfigErr := node.LoadConfig()
	if nodeConfigErr != nil {
		log.Fatalf("node config error %v", nodeConfigErr)
	}
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeConfig.NodeID)
	fsm := demory.NewDemory()
	basedir := filepath.Join("/tmp", nodeConfig.NodeID)
	mkdirErr := os.MkdirAll(basedir, os.ModePerm)

	if mkdirErr != nil {
		log.Fatalf("mkdir error %v", mkdirErr)
	}
	logStore, logStoreErr := boltdb.NewBoltStore(filepath.Join(basedir, "logs.dat"))

	if logStoreErr != nil {
		log.Fatalf("logstore error %v", logStoreErr)
	}
	stableStore, stableStoreErr := boltdb.NewBoltStore(filepath.Join(basedir, "stable.dat"))

	if stableStoreErr != nil {
		log.Fatalf("stablestore error %v", stableStoreErr)
	}
	snapshotStore, snapshotStoreErr := raft.NewFileSnapshotStore(basedir, 3, os.Stderr)

	if snapshotStoreErr != nil {
		log.Fatalf("snapshotstore error %v", snapshotStoreErr)
	}
	manager := transport.New(raft.ServerAddress("localhost:8081"), []grpc.DialOption{grpc.WithInsecure()})
	r, raftErr := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, manager.Transport())

	if raftErr != nil {
		log.Fatalf("raft error %v", raftErr)
	}

	if nodeConfig.Bootstrap {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(nodeConfig.NodeID),
					Address:  raft.ServerAddress(nodeConfig.NodeAddress),
				},
			},
		}

		cluster := r.BootstrapCluster(cfg)
		if err := cluster.Error(); err != nil {
			log.Fatalf("bootstrap error %v", err)
		}
	}

	var memberDiscovery discovery.Discovery
	if nodeConfig.DiscoveryStrategy == discovery.StrategyPort {
		memberDiscovery = discovery.NewPortDiscovery(8000, 8100, "localhost", nodeConfig.NodeAddress, nodeConfig.NodeID, r)
	} else if nodeConfig.DiscoveryStrategy == discovery.StrategyKubernetes {
		memberDiscovery = discovery.NewKubernetesDiscovery(nodeConfig.KubernetesNamespace, nodeConfig.KubernetesService, r)
	} else {
		log.Fatalf("invalid discovery %s", nodeConfig.DiscoveryStrategy)
	}
	discoveryErr := memberDiscovery.Discover()
	if discoveryErr != nil {
		log.Fatalf("discovery err %v", discoveryErr)
	}
	socket, socketErr := net.Listen("tcp", fmt.Sprintf(":%d", nodeConfig.Port))
	if socketErr != nil {
		log.Fatalf("socket error %v", socketErr)
	}

	server := grpc.NewServer()
	proto.RegisterDemoryServer(server, demory.NewRPCInterface(r, fsm))
	manager.Register(server)
	leaderhealth.Setup(r, server, []string{"Leader"})
	raftadmin.Register(server, r)
	reflection.Register(server)
	serveErr := server.Serve(socket)
	if serveErr != nil {
		log.Fatalf("serve error %v", serveErr)
	}
}
