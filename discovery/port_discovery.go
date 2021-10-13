package discovery

import (
	"context"
	"fmt"
	"github.com/hashicorp/raft"
	proto "github.com/huseyinbabal/demory-proto/golang/demory"
	"google.golang.org/grpc"
	"sync"
)

type portDiscovery struct {
	raft          *raft.Raft
	minPort       int
	maxPort       int
	host          string
	memberAddress string
	serverId      string
}

var _ Discovery = &portDiscovery{}

func NewPortDiscovery(minPort, maxPort int, host, memberAddress, serverId string, raft *raft.Raft) *portDiscovery {
	return &portDiscovery{
		raft:          raft,
		minPort:       minPort,
		maxPort:       maxPort,
		host:          host,
		memberAddress: memberAddress,
		serverId:      serverId,
	}
}

func (p *portDiscovery) Discover() error {
	if p.raft.State() != raft.Follower {
		return nil
	}
	var wg sync.WaitGroup
	for i := p.minPort; i < p.maxPort; i++ {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			p.dial(p.host, port)
		}(i)
	}
	wg.Wait()
	return nil
}

func (p *portDiscovery) dial(host string, port int) {
	dest := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.Dial(dest, grpc.WithInsecure())
	if err != nil {
		fmt.Println("failed to connect " + dest)
	} else {
		defer conn.Close()
		client := proto.NewDemoryClient(conn)
		_, err := client.JoinToCluster(context.Background(), &proto.JoinToClusterRequest{
			ServerAddress: p.memberAddress,
			ServerId:      p.serverId,
			PreviousIndex: 0,
		})
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("success")
		}
	}
}
