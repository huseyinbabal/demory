package fsm

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	transport "github.com/Jille/raft-grpc-transport"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"github.com/huseyinbabal/demory/node"
	"google.golang.org/grpc"
)

type Fsm struct {
	Raft    *raft.Raft
	Manager *transport.Manager
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

	manager := transport.New(raft.ServerAddress(nodeConfig.NodeAddress), []grpc.DialOption{grpc.WithInsecure()})
	r, raftErr := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, manager.Transport())

	if raftErr != nil {
		log.Fatalf("raft error %v", raftErr)
	}
	return &Fsm{
		Raft:    r,
		Manager: manager,
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
