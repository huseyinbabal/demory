package fsm

import (
	"encoding/json"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"github.com/huseyinbabal/demory/node"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Fsm struct {
	Raft    *raft.Raft
	Manager *raft.NetworkTransport
	mutex   sync.RWMutex
}

var _ raft.FSM = &Fsm{}

func New(nodeConfig node.Config) *Fsm {
	var fsm *Fsm

	config := raft.DefaultConfig()

	config.LocalID = raft.ServerID(nodeConfig.NodeID)
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

	//manager := transport.New(raft.ServerAddress(nodeConfig.NodeAddress), []grpc.DialOption{grpc.WithInsecure()})
	tcpAddr, tcpAddrErr := net.ResolveTCPAddr("tcp", nodeConfig.NodeAddress)
	if tcpAddrErr != nil {
		log.Fatalf("failed to resolve a TCP address: %v", tcpAddrErr)
	}

	transport, transportErr := raft.NewTCPTransport(nodeConfig.NodeAddress, tcpAddr, 2, 10*time.Second, os.Stderr)
	if transportErr != nil {
		log.Fatalf("transport err %v", transportErr)
	}
	r, raftErr := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)

	if raftErr != nil {
		log.Fatalf("raft error %v", raftErr)
	}
	return &Fsm{
		Raft:    r,
		Manager: transport,
	}
}

func (f *Fsm) Apply(log *raft.Log) interface{} {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	var applyRequest ApplyRequest

	err := json.Unmarshal(log.Data, &applyRequest)

	if err != nil {
		return ApplyResponse{
			Data:  nil,
			Error: err,
		}
	}

	return ApplyResponse{
		Data:  applyRequest.Command(),
		Error: nil,
	}
}

func (f *Fsm) Snapshot() (raft.FSMSnapshot, error) {
	panic("implement me")
}

func (f *Fsm) Restore(closer io.ReadCloser) error {
	panic("implement me")
}
