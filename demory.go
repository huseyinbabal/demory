package main

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/raft"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"sync"
	"time"
)

type demory struct {
	DB    map[string]string
	mutex sync.RWMutex
}

var _ raft.FSM = &demory{}

func NewDemory() *demory {
	return &demory{
		DB: make(map[string]string),
	}
}

func (d *demory) Apply(log *raft.Log) interface{} {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	var data proto.PutRequest
	json.Unmarshal(log.Data, &data)
	d.DB[data.Key] = data.Value
	return nil
}

func (d *demory) Snapshot() (raft.FSMSnapshot, error) {
	panic("implement me")
}

func (d *demory) Restore(closer io.ReadCloser) error {
	panic("implement me")
}

type rpcInterface struct {
	raft   *raft.Raft
	demory *demory
	proto.UnimplementedDemoryServer
}

func (r rpcInterface) Put(ctx context.Context, req *proto.PutRequest) (*emptypb.Empty, error) {
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

func (r rpcInterface) Get(ctx context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {
	return &proto.GetResponse{Value: r.demory.DB[req.Key]}, nil
}
