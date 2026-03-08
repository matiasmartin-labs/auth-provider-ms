package pkg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerConfig_GetPort(t *testing.T) {
	config := &serverConfig{Port: 8080}
	assert.Equal(t, 8080, config.GetPort())
}

func TestServerConfig_GetPort_DefaultValue(t *testing.T) {
	config := &serverConfig{}
	assert.Equal(t, 0, config.GetPort())
}

func TestServerConfig_GetReadTimeout(t *testing.T) {
	config := &serverConfig{ReadTimeout: 10 * time.Second}
	assert.Equal(t, 10*time.Second, config.GetReadTimeout())
}

func TestServerConfig_GetReadTimeout_DefaultValue(t *testing.T) {
	config := &serverConfig{}
	assert.Equal(t, time.Duration(0), config.GetReadTimeout())
}

func TestServerConfig_GetWriteTimeout(t *testing.T) {
	config := &serverConfig{WriteTimeout: 15 * time.Second}
	assert.Equal(t, 15*time.Second, config.GetWriteTimeout())
}

func TestServerConfig_GetWriteTimeout_DefaultValue(t *testing.T) {
	config := &serverConfig{}
	assert.Equal(t, time.Duration(0), config.GetWriteTimeout())
}

func TestServerConfig_GetMaxHeaderBytes(t *testing.T) {
	config := &serverConfig{MaxHeaderBytes: 1 << 20}
	assert.Equal(t, 1<<20, config.GetMaxHeaderBytes())
}

func TestServerConfig_GetMaxHeaderBytes_DefaultValue(t *testing.T) {
	config := &serverConfig{}
	assert.Equal(t, 0, config.GetMaxHeaderBytes())
}

func TestServerConfig_AllFields(t *testing.T) {
	config := &serverConfig{
		Port:           3000,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 2 << 20,
	}

	assert.Equal(t, 3000, config.GetPort())
	assert.Equal(t, 30*time.Second, config.GetReadTimeout())
	assert.Equal(t, 60*time.Second, config.GetWriteTimeout())
	assert.Equal(t, 2<<20, config.GetMaxHeaderBytes())
}
