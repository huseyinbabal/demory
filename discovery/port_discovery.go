package discovery

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/hashicorp/raft"
	proto "github.com/huseyinbabal/demory-proto/golang/demory"
	"google.golang.org/grpc"
)

type portDiscovery struct {
	raft               *raft.Raft
	minPort            int
	maxPort            int
	host               string
	memberAddress      string
	serverId           string
	clusterInitialized bool
}

var _ Discovery = &portDiscovery{}

func NewPortDiscovery(minPort, maxPort int, host, memberAddress, serverId string, raft *raft.Raft) *portDiscovery {
	return &portDiscovery{
		raft:               raft,
		minPort:            minPort,
		maxPort:            maxPort,
		host:               host,
		memberAddress:      memberAddress,
		serverId:           serverId,
		clusterInitialized: false,
	}
}

func (p *portDiscovery) Discover() error {
	if !p.clusterInitialized && len(p.raft.GetConfiguration().Configuration().Servers) == 0 {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(p.serverId),
					Address:  raft.ServerAddress(p.memberAddress),
				},
			},
		}

		cluster := p.raft.BootstrapCluster(cfg)
		if err := cluster.Error(); err == nil {
			log.Println("Cluster is initialized successfully.")
		}
		p.clusterInitialized = true
	}
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
		log.Println("failed to connect " + dest)
	} else {
		defer conn.Close()
		client := proto.NewDemoryClient(conn)
		_, err := client.JoinToCluster(context.Background(), &proto.JoinToClusterRequest{
			ServerAddress: p.memberAddress,
			ServerId:      p.serverId,
			PreviousIndex: 0,
		})
		if err != nil {
			log.Println(err)
		} else {
			log.Println("success")
		}
	}
}
