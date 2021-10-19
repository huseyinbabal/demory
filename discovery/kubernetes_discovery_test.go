package discovery

import (
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestNewKubernetesDiscovery(t *testing.T) {
	clientset := fake.NewSimpleClientset(&v1.Endpoints{
		ObjectMeta: v12.ObjectMeta{
			Namespace: "default",
			Name:      "nginx",
			Labels:    map[string]string{"app": "nginx"},
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: "127.0.0.1,127.0.0.2",
					},
				},
				Ports: []v1.EndpointPort{
					{
						Port: 8081,
					},
					{
						Port: 8082,
					},
				},
			},
		},
	})
	client := NewKubernetesDiscoveryWithClient(clientset, "default", "nginx")
	err := client.Discover()
	if err != nil {
		t.Errorf("discovery failed %v", err)
	}
}
