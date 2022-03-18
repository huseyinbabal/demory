package node

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	NodeID              string `mapstructure:"NODE_ID"`
	NodeAddress         string `mapstructure:"NODE_ADDRESS"`
	Port                int    `mapstructure:"PORT"`
	MinPort             int    `mapstructure:"MIN_PORT"`
	MaxPort             int    `mapstructure:"MAX_PORT"`
	DiscoveryStrategy   string `mapstructure:"DISCOVERY_STRATEGY"`
	KubernetesService   string `mapstructure:"KUBERNETES_SERVICE"`
	KubernetesNamespace string `mapstructure:"KUBERNETES_NAMESPACE"`
}

func LoadConfig() (config *Config, e error) {
	viper.SetEnvPrefix("DEMORY")
	bindEnv("NODE_ID")
	bindEnv("NODE_ADDRESS")
	bindEnv("PORT")
	bindEnv("MIN_PORT")
	bindEnv("MAX_PORT")
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
