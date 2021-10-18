package discovery

type Strategy string

const (
	StrategyPort       = "port"
	StrategyKubernetes = "kubernetes"
)

type Discovery interface {
	Discover() error
}
