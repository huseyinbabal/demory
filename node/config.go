package node

import (
	"github.com/huseyinbabal/demory/discovery"
	"github.com/spf13/viper"
	"log"
	"os"
)

type Config struct {
	NodeID              string             `mapstructure:"NODE_ID"`
	Bootstrap           bool               `mapstructure:"BOOTSTRAP"`
	NodeAddress         string             `mapstructure:"NODE_ADDRESS"`
	Port                int                `mapstructure:"PORT"`
	DiscoveryStrategy   discovery.Strategy `mapstructure:"DISCOVERY_STRATEGY"`
	KubernetesService   string             `mapstructure:"KUBERNETES_SERVICE"`
	KubernetesNamespace string             `mapstructure:"KUBERNETES_NAMESPACE"`
}

func LoadConfig() (config Config, e error) {
	viper.SetEnvPrefix("DEMORY")
	bindEnv("NODE_ID")
	bindEnv("BOOTSTRAP")
	bindEnv("NODE_ADDRESS")
	bindEnv("PORT")
	bindEnv("DISCOVERY_STRATEGY")
	bindEnv("KUBERNETES_SERVICE")
	bindEnv("KUBERNETES_NAMESPACE")
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	configFile := viper.GetString("config")
	viper.SetConfigFile(configFile)

	if _, err := os.Stat(configFile); err == nil {
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("failed to load config %v", err)
		}
	}
	e = viper.Unmarshal(&config)
	return
}

func bindEnv(name string) {
	err := viper.BindEnv(name)
	if err != nil {
		log.Printf("failed to bind env variable: %s, err: %v.\n", name, err)
	}
}
