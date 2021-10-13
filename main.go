package main

import (
	"flag"
	"fmt"
	"github.com/Jille/raft-grpc-leader-rpc/leaderhealth"
	transport "github.com/Jille/raft-grpc-transport"
	"github.com/Jille/raftadmin"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	proto "github.com/huseyinbabal/demory-proto/golang/demory"
	"github.com/huseyinbabal/demory/discovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"path/filepath"
)

var (
	bootstrap   = flag.Bool("bootstrap", false, "Bootstrap RAFT Cluster")
	port        = flag.Int("port", 0, "Port")
	nodeId      = flag.String("node_id", "", "Node ID")
	nodeAddress = flag.String("node_address", "", "Node Address")
)

func main() {
	flag.Parse()
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(*nodeId)
	fsm := NewDemory()
	basedir := filepath.Join("/tmp", *nodeId)
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

	if *bootstrap {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(*nodeId),
					Address:  raft.ServerAddress(*nodeAddress),
				},
			},
		}

		cluster := r.BootstrapCluster(cfg)
		if err := cluster.Error(); err != nil {
			log.Fatalf("bootstrap error %v", err)
		}
	} else {
		discovery.NewPortDiscovery(8000, 8100, "localhost", *nodeAddress, *nodeId).Discover()
	}

	socket, socketErr := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if socketErr != nil {
		log.Fatalf("socket error %v", socketErr)
	}

	server := grpc.NewServer()
	proto.RegisterDemoryServer(server, &rpcInterface{
		raft:   r,
		demory: fsm,
	})
	manager.Register(server)
	leaderhealth.Setup(r, server, []string{"Leader"})
	raftadmin.Register(server, r)
	reflection.Register(server)
	serveErr := server.Serve(socket)
	if serveErr != nil {
		log.Fatalf("serve error %v", serveErr)
	}
}
