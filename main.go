package main

import (
	transport "github.com/Jille/raft-grpc-transport"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"path/filepath"
)

func main() {
	config:=raft.DefaultConfig()
	config.LocalID=raft.ServerID("test")
	fsm:=NewDemory()
	basedir:=filepath.Join("/tmp","livecodingturkey")
	mkdirErr := os.MkdirAll(basedir, os.ModePerm)
	if mkdirErr != nil {
		log.Fatalf("mkdir error %v", mkdirErr)
	}
	logStore, logStoreErr := boltdb.NewBoltStore(filepath.Join(basedir,"logs.dat"))
	if logStoreErr != nil {
		log.Fatalf("logstore error %v", logStoreErr)
	}
	stableStore, stableStoreErr := boltdb.NewBoltStore(filepath.Join(basedir,"stable.dat"))
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

	cfg:=raft.Configuration{
		Servers: []raft.Server{
			{
				Suffrage: raft.Voter,
				ID:"test",
				Address: raft.ServerAddress("localhost:8081"),
			},
		},
	}

	cluster := r.BootstrapCluster(cfg)
	if err:=cluster.Error();err!=nil{
		log.Fatalf("bootstrap error %v",err)
	}

	socket, socketErr := net.Listen("tcp", ":8081")
	if socketErr != nil {
		log.Fatalf("socket error %v",socketErr)
	}

	server:= grpc.NewServer()
	proto.RegisterDemoryServer(server,&rpcInterface{
		raft:r,
		demory: fsm,
	})
	manager.Register(server)
	serveErr := server.Serve(socket)
	if serveErr != nil {
		log.Fatalf("serve error %v",serveErr)
	}
}