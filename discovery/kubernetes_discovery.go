package discovery

import (
	"context"
	"fmt"
	"github.com/hashicorp/raft"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
	"time"
)

type kubernetesDiscovery struct {
	Clientset kubernetes.Interface
	Namespace string
	Service   string
	Raft      *raft.Raft
}

var _ Discovery = &kubernetesDiscovery{}

func NewKubernetesDiscovery(namespace string, service string, r *raft.Raft) *kubernetesDiscovery {
	config, configErr := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if configErr != nil {
		log.Fatalf("Failed to get k8s in cluster config %v", configErr)
	}
	clientset, clientsetErr := kubernetes.NewForConfig(config)
	if clientsetErr != nil {
		log.Fatalf("Failed to get clientset %v", clientsetErr)
	}
	return &kubernetesDiscovery{
		Namespace: namespace,
		Service:   service,
		Raft:      r,
		Clientset: clientset,
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

	log.Println("discovering...")
	list, err := k.Clientset.CoreV1().Endpoints(k.Namespace).List(context.Background(), v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", "app", k.Service),
	})
	if err != nil {
		return err
	}
	var endpoints []string
	for _, item := range list.Items {
		for i := 0; i < len(item.Subsets[0].Addresses); i++ {
			endpoints = append(endpoints, fmt.Sprintf("%s:%d", item.Subsets[0].Addresses[i].IP, item.Subsets[0].Ports[0].Port))

		}
	}
	log.Println(endpoints)
	return nil
}
