package pkg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfiguration_WithValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := `
server:
  port: 8080
  read-timeout: 10s
  write-timeout: 15s
  max-header-bytes: 1048576
security:
  oauth2:
    client:
      google:
        client-id: test-client-id
        client-secret: test-secret
        redirect-uri: http://localhost:8080/callback
        scopes:
          - email
          - profile
        state: random-state
        user-info-uri: https://www.googleapis.com/oauth2/v2/userinfo
  redirect:
    enabled: true
    url: http://localhost:3000
  cookie:
    secure: true
    max-age: 1h
    http-only: true
    same-site: Lax
  login:
    allowed-emails:
      - test@example.com
  jwt:
    issuer: test-issuer
    audience: test-audience
    expiration-time: 1h
  auth:
    enabled: true
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	config := NewConfiguration()

	assert.NotNil(t, config)
	assert.Equal(t, 8080, config.GetServerConfig().GetPort())
	assert.Equal(t, "test-issuer", config.GetSecurityConfig().GetJWTConfig().GetIssuer())
}

func TestNewConfiguration_WithEnvVarExpansion(t *testing.T) {
	tmpDir := t.TempDir()

	os.Setenv("TEST_CLIENT_ID", "env-client-id")
	defer os.Unsetenv("TEST_CLIENT_ID")

	configContent := `
server:
  port: 9000
security:
  oauth2:
    client:
      google:
        client-id: ${TEST_CLIENT_ID}
        client-secret: test-secret
        redirect-uri: http://localhost:8080/callback
        scopes:
          - email
        state: state
        user-info-uri: https://example.com/userinfo
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	config := NewConfiguration()

	assert.NotNil(t, config)
	assert.Equal(t, "env-client-id", config.GetSecurityConfig().GetOAuth2Config().GetGoogleConfig().GetClientID())
}

func TestNewConfiguration_PanicsOnMissingConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	assert.Panics(t, func() {
		NewConfiguration()
	})
}
