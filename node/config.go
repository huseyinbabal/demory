package node

import (
	"github.com/spf13/viper"
	"log"
	"os"
)

type Config struct {
	NodeID              string `mapstructure:"NODE_ID"`
	Bootstrap           bool   `mapstructure:"BOOTSTRAP"`
	NodeAddress         string `mapstructure:"NODE_ADDRESS"`
	Port                int    `mapstructure:"PORT"`
	DiscoveryStrategy   string `mapstructure:"DISCOVERY_STRATEGY"`
	KubernetesService   string `mapstructure:"KUBERNETES_SERVICE"`
	KubernetesNamespace string `mapstructure:"KUBERNETES_NAMESPACE"`
}

func LoadConfig() (config Config, e error) {
	viper.SetEnvPrefix("DEMORY")
	viper.BindEnv("NODE_ID")
	viper.BindEnv("BOOTSTRAP")
	viper.BindEnv("NODE_ADDRESS")
	viper.BindEnv("PORT")
	viper.BindEnv("DISCOVERY_STRATEGY")
	viper.BindEnv("KUBERNETES_SERVICE")
	viper.BindEnv("KUBERNETES_NAMESPACE")
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
