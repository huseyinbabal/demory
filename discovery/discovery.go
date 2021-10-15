package discovery

type DiscoveryStrategy string

const (
	Port       = "port"
	Kubernetes = "kubernetes"
)

type Discovery interface {
	Discover() error
}
