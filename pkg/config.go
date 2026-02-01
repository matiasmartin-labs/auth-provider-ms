package pkg

import (
	"fmt"

	"github.com/spf13/viper"
)

type Configuration interface {
	GetServerConfig() ServerConfig
}

type configuration struct {
	Server *serverConfig `mapstructure:"server"`
}

func (c *configuration) GetServerConfig() ServerConfig {
	return c.Server
}

var cfg configuration

func initViper() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func NewConfiguration() Configuration {
	initViper()
	return &cfg
}
