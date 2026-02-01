package pkg

import "time"

type ServerConfig interface {
	GetPort() int
	GetReadTimeout() time.Duration
	GetWriteTimeout() time.Duration
	GetMaxHeaderBytes() int
}

type serverConfig struct {
	Port           int           `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"read-timeout"`
	WriteTimeout   time.Duration `mapstructure:"write-timeout"`
	MaxHeaderBytes int           `mapstructure:"max-header-bytes"`
}

func (s *serverConfig) GetPort() int {
	return s.Port
}

func (s *serverConfig) GetReadTimeout() time.Duration {
	return s.ReadTimeout
}

func (s *serverConfig) GetWriteTimeout() time.Duration {
	return s.WriteTimeout
}

func (s *serverConfig) GetMaxHeaderBytes() int {
	return s.MaxHeaderBytes
}
