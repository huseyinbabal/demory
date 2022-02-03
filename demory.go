package demory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	proto "github.com/huseyinbabal/demory-proto/golang/demory"
	"github.com/huseyinbabal/demory/ds/cache"

	"github.com/Jille/raft-grpc-leader-rpc/leaderhealth"
	"github.com/Jille/raftadmin"
	"github.com/hashicorp/raft"
	"github.com/huseyinbabal/demory/discovery"
	"github.com/huseyinbabal/demory/ds/hashmap"
	"github.com/huseyinbabal/demory/fsm"
	"github.com/huseyinbabal/demory/node"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Demory is for representing data structure storage
// It also has basic api interface for data operations.
type Demory struct {
	hashMap *hashmap.HashMap
	cache   *cache.Cache
	fsm     *fsm.Fsm
	config  *node.Config
	proto.UnimplementedDemoryServer
}

// New for creating new instance of in-memory database.
func New() *Demory {
	nodeConfig, nodeConfigErr := node.LoadConfig()
	if nodeConfigErr != nil {
		log.Fatalf("node config error %v", nodeConfigErr)
	}

	return &Demory{
		hashMap: hashmap.New(),
		cache:   cache.New(),
		fsm:     fsm.New(*nodeConfig),
		config:  nodeConfig,
	}
}

// MapPut saves data into store.
func (d *Demory) MapPut(ctx context.Context, req *proto.MapPutRequest) (*emptypb.Empty, error) {
	bytes, bytesErr := json.Marshal(fsm.ApplyRequest{
		Command: func() interface{} {
			return d.hashMap.Put(req.GetName(), req.GetKey(), req.GetValue())
		},
	})
	if bytesErr != nil {
		return new(emptypb.Empty), bytesErr
	}

	apply := d.fsm.Raft.Apply(bytes, time.Second)

	if err := apply.Error(); err != nil {
		return new(emptypb.Empty), err
	}

	return new(emptypb.Empty), nil
}

// MapGet retrieves data from store.
func (d *Demory) MapGet(ctx context.Context, req *proto.MapGetRequest) (*proto.MapGetResponse, error) {
	return &proto.MapGetResponse{Value: d.hashMap.Get(req.GetName(), req.GetKey())}, nil
}

// MapPutIfAbsent inserts value at specified key if there is no value
func (d *Demory) MapPutIfAbsent(ctx context.Context, req *proto.MapPutIfAbsentRequest) (*emptypb.Empty, error) {
	bytes, bytesErr := json.Marshal(fsm.ApplyRequest{
		Command: func() interface{} {
			return d.hashMap.PutIfAbsent(req.GetName(), req.GetKey(), req.GetValue())
		},
	})
	if bytesErr != nil {
		return new(emptypb.Empty), bytesErr
	}

	apply := d.fsm.Raft.Apply(bytes, time.Second)

	if err := apply.Error(); err != nil {
		return new(emptypb.Empty), err
	}

	return new(emptypb.Empty), nil
}

// MapRemove removes the value at specified key
func (d *Demory) MapRemove(ctx context.Context, req *proto.MapRemoveRequest) (*emptypb.Empty, error) {
	bytes, bytesErr := json.Marshal(fsm.ApplyRequest{
		Command: func() interface{} {
			return d.hashMap.Remove(req.GetName(), req.GetKey())
		},
	})
	if bytesErr != nil {
		return new(emptypb.Empty), bytesErr
	}

	apply := d.fsm.Raft.Apply(bytes, time.Second)

	if err := apply.Error(); err != nil {
		return new(emptypb.Empty), err
	}

	return new(emptypb.Empty), nil
}

// MapClear clears all the entries in map specified with name
func (d *Demory) MapClear(ctx context.Context, req *proto.MapClearRequest) (*emptypb.Empty, error) {
	bytes, bytesErr := json.Marshal(fsm.ApplyRequest{
		Command: func() interface{} {
			return d.hashMap.Clear(req.GetName())
		},
	})
	if bytesErr != nil {
		return new(emptypb.Empty), bytesErr
	}

	apply := d.fsm.Raft.Apply(bytes, time.Second)

	if err := apply.Error(); err != nil {
		return new(emptypb.Empty), err
	}

	return new(emptypb.Empty), nil
}

// CachePut saves data into store.
func (d *Demory) CachePut(ctx context.Context, req *proto.CachePutRequest) (*emptypb.Empty, error) {
	bytes, bytesErr := json.Marshal(fsm.ApplyRequest{
		Command: func() interface{} {
			d.cache.Put(req.GetName(), req.GetKey(), req.GetValue())
			return nil
		},
	})
	if bytesErr != nil {
		return new(emptypb.Empty), bytesErr
	}

	apply := d.fsm.Raft.Apply(bytes, time.Second)

	if err := apply.Error(); err != nil {
		return new(emptypb.Empty), err
	}

	return new(emptypb.Empty), nil
}

// CacheGet retrieves data from store.
func (d *Demory) CacheGet(ctx context.Context, req *proto.CacheGetRequest) (*proto.CacheGetResponse, error) {
	return &proto.CacheGetResponse{Value: d.cache.Get(req.GetName(), req.GetKey())}, nil
}

// Remove removes the value at specified key
func (d *Demory) CacheRemove(ctx context.Context, req *proto.CacheRemoveRequest) (*emptypb.Empty, error) {
	bytes, bytesErr := json.Marshal(fsm.ApplyRequest{
		Command: func() interface{} {
			d.cache.Remove(req.GetName(), req.GetKey())
			return nil
		},
	})
	if bytesErr != nil {
		return new(emptypb.Empty), bytesErr
	}

	apply := d.fsm.Raft.Apply(bytes, time.Second)

	if err := apply.Error(); err != nil {
		return new(emptypb.Empty), err
	}

	return new(emptypb.Empty), nil
}

// CacheClear clears all the entries in cache specified with name
func (d *Demory) CacheClear(ctx context.Context, req *proto.CacheClearRequest) (*emptypb.Empty, error) {
	bytes, bytesErr := json.Marshal(fsm.ApplyRequest{
		Command: func() interface{} {
			d.cache.Clear(req.GetName())
			return nil
		},
	})
	if bytesErr != nil {
		return new(emptypb.Empty), bytesErr
	}

	apply := d.fsm.Raft.Apply(bytes, time.Second)

	if err := apply.Error(); err != nil {
		return new(emptypb.Empty), err
	}

	return new(emptypb.Empty), nil
}

// JoinToCluster is  used by port discovery to allow joining cluster
// by using leader node.
func (d *Demory) JoinToCluster(ctx context.Context, request *proto.JoinToClusterRequest) (*proto.JoinToClusterResponse,
	error) {
	if d.fsm.Raft.State() == raft.Leader {
		d.fsm.Raft.AddVoter(raft.ServerID(request.ServerId), raft.ServerAddress(request.ServerAddress),
			request.PreviousIndex, time.Second)
		return &proto.JoinToClusterResponse{Result: proto.JoinToClusterResponse_SUCCESS}, nil

	}
	return &proto.JoinToClusterResponse{
		Result: proto.JoinToClusterResponse_NOT_A_LEADER,
	}, errors.New("not a leader")
}

func (d *Demory) Run() {
	nodeConfig := d.config

	discoveryErr := discovery.Get(nodeConfig, d.fsm.Raft).Discover()

	if discoveryErr != nil {
		log.Fatalf("discovery err %v", discoveryErr)
	}

	socket, socketErr := net.Listen("tcp", fmt.Sprintf(":%d", nodeConfig.Port))
	if socketErr != nil {
		log.Fatalf("socket error %v", socketErr)
	}

	server := grpc.NewServer()
	proto.RegisterDemoryServer(server, d)
	leaderhealth.Setup(d.fsm.Raft, server, []string{"Leader"})
	raftadmin.Register(server, d.fsm.Raft)
	reflection.Register(server)
	serveErr := server.Serve(socket)

	if serveErr != nil {
		log.Fatalf("serve error %v", serveErr)
	}
}
