package discovery

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/raft"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/utils/strings/slices"
)

type kubernetesDiscovery struct {
	Clientset          kubernetes.Interface
	Namespace          string
	Service            string
	NodeAddress        string
	NodeID             string
	Raft               *raft.Raft
	ClusterInitialized bool
}

var _ Discovery = &kubernetesDiscovery{}

func NewKubernetesDiscovery(namespace string, service string, nodeAddress string, nodeID string, r *raft.Raft) *kubernetesDiscovery {
	config, configErr := rest.InClusterConfig()
	if configErr != nil {
		log.Fatalf("Failed to get k8s in cluster config %v", configErr)
	}
	clientset, clientsetErr := kubernetes.NewForConfig(config)
	if clientsetErr != nil {
		log.Fatalf("Failed to get clientset %v", clientsetErr)
	}
	return &kubernetesDiscovery{
		Namespace:          namespace,
		Service:            service,
		NodeAddress:        nodeAddress,
		NodeID:             nodeID,
		Raft:               r,
		Clientset:          clientset,
		ClusterInitialized: false,
	}
}

func NewKubernetesDiscoveryWithClient(clientset *fake.Clientset, namespace string, service string) *kubernetesDiscovery {
	return &kubernetesDiscovery{
		Clientset: clientset,
		Namespace: namespace,
		Service:   service,
	}
}

func (k kubernetesDiscovery) Discover() error {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			<-ticker.C
			if err := k.discover(); err != nil {
				log.Printf("failed to discover cluster nodes %v.\n", err)
			}
		}
	}()
	return nil
}

func (k kubernetesDiscovery) discover() error {
	if !k.ClusterInitialized && len(k.Raft.GetConfiguration().Configuration().Servers) == 0 {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(k.NodeID),
					Address:  raft.ServerAddress(k.NodeAddress),
				},
			},
		}

		cluster := k.Raft.BootstrapCluster(cfg)
		if err := cluster.Error(); err == nil {
			log.Println("Cluster is initialized successfully.")
		}
		k.ClusterInitialized = true
	}
	if k.Raft.Leader() != raft.ServerAddress(k.NodeAddress) {
		return nil
	}
	list, err := k.Clientset.CoreV1().Endpoints(k.Namespace).List(context.Background(), v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", "app.kubernetes.io/name", k.Service),
	})
	if err != nil {
		return err
	}
	endpoints := make(map[string]string)
	for _, item := range list.Items {
		if len(item.Subsets) == 0 {
			continue
		}
		for i := 0; i < len(item.Subsets[0].Addresses); i++ {
			endpoints[item.Subsets[0].Addresses[i].Hostname] = fmt.Sprintf("%s:%d", item.Subsets[0].Addresses[i].IP, item.Subsets[0].Ports[0].Port)
		}
	}
	k.refreshClusterMembersIfNeeded(endpoints)
	return nil
}

func (k kubernetesDiscovery) refreshClusterMembersIfNeeded(foundMembers map[string]string) {
	var existingClusterMembers []string
	var foundClusterMembers []string
	membersToAdd := make(map[string]string)
	var membersToRemove []string

	for _, member := range foundMembers {
		foundClusterMembers = append(foundClusterMembers, member)
	}

	for _, server := range k.Raft.GetConfiguration().Configuration().Servers {
		if !slices.Contains(foundClusterMembers, string(server.Address)) {
			membersToRemove = append(membersToRemove, string(server.ID))
			k.Raft.RemoveServer(server.ID, 0, time.Second)
		} else {
			existingClusterMembers = append(existingClusterMembers, string(server.Address))
		}
	}

	for nodeID, nodeAddress := range foundMembers {
		if !slices.Contains(existingClusterMembers, nodeAddress) {
			membersToAdd[nodeID] = nodeAddress
			k.Raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(nodeAddress), 0, time.Second)
		}
	}

	if len(membersToAdd) > 0 {
		log.Printf("Found new nodes: %v\n", membersToAdd)
	}

	if len(membersToRemove) > 0 {
		log.Printf("Removed nodes: %v\n", membersToRemove)
	}
}
