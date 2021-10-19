package demory

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/hashicorp/raft"
	proto "github.com/huseyinbabal/demory-proto/golang/demory"
	"google.golang.org/protobuf/types/known/emptypb"

	"io"
	"sync"
	"time"
)

// Demory is for representing data structure storage
// It also has basic api interface for data operations.
type Demory struct {
	mapStore map[string]string
	mutex    sync.RWMutex
}

var _ raft.FSM = &Demory{}

// NewDemory for creating new instance of in-memory database
func NewDemory() *Demory {
	return &Demory{
		mapStore: make(map[string]string),
	}
}

// Apply is for storing log message to stable store within RAFT protocol
func (d *Demory) Apply(log *raft.Log) interface{} {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	var data proto.MapPutRequest
	err := json.Unmarshal(log.Data, &data)
	if err != nil {
		return err
	}
	d.mapStore[data.Key] = data.Value
	return nil
}

// Snapshot is for taking the snapshot of existing logs
func (d *Demory) Snapshot() (raft.FSMSnapshot, error) {
	panic("implement me")
}

// Restore is used for getting last recent data into stable store
func (d *Demory) Restore(closer io.ReadCloser) error {
	panic("implement me")
}

// RPCInterface is public interface for consumers.
type RPCInterface struct {
	raft   *raft.Raft
	demory *Demory
	proto.UnimplementedDemoryServer
}

// NewRPCInterface creates an instance of RPCInterface to serve data
// related endpoints to consumers
func NewRPCInterface(raft *raft.Raft, demory *Demory) *RPCInterface {
	return &RPCInterface{
		raft:   raft,
		demory: demory,
	}
}

// Put saves data into store
func (r RPCInterface) Put(ctx context.Context, req *proto.MapPutRequest) (*emptypb.Empty, error) {
	bytes, bytesErr := json.Marshal(req)
	if bytesErr != nil {
		return new(emptypb.Empty), bytesErr
	}
	apply := r.raft.Apply(bytes, time.Second)
	if err := apply.Error(); err != nil {
		return new(emptypb.Empty), err
	}
	return new(emptypb.Empty), nil
}

// Get retrieves data from store
func (r RPCInterface) Get(ctx context.Context, req *proto.MapGetRequest) (*proto.MapGetResponse, error) {
	return &proto.MapGetResponse{Value: r.demory.mapStore[req.Key]}, nil
}

// JoinToCluster is  used by port discovery to allow joining cluster
// By using leader node
func (r RPCInterface) JoinToCluster(ctx context.Context, request *proto.JoinToClusterRequest) (*proto.JoinToClusterResponse, error) {
	if r.raft.State() == raft.Leader {
		r.raft.AddVoter(raft.ServerID(request.ServerId), raft.ServerAddress(request.ServerAddress), request.PreviousIndex, time.Second)
		return &proto.JoinToClusterResponse{Result: proto.JoinToClusterResponse_SUCCESS}, nil
	}
	return &proto.JoinToClusterResponse{Result: proto.JoinToClusterResponse_NOT_A_LEADER}, errors.New("not a leader")
}
