package pkg

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Configuration interface {
	GetServerConfig() ServerConfig
	GetSecurityConfig() SecurityConfig
}

type configuration struct {
	Server   *serverConfig   `mapstructure:"server"`
	Security *securityConfig `mapstructure:"security"`
}

func (c *configuration) GetServerConfig() ServerConfig {
	return c.Server
}

func (c *configuration) GetSecurityConfig() SecurityConfig {
	return c.Security
}

var cfg configuration

func initViper() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	configFile := "./config.yaml"
	content, err := os.ReadFile(configFile)
	if err != nil {
		panic(fmt.Errorf("fatal error reading config file: %w", err))
	}

	expanded := os.ExpandEnv(string(content))

	viper.ReadConfig(strings.NewReader(expanded))

	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func NewConfiguration() Configuration {
	initViper()
	return &cfg
}
