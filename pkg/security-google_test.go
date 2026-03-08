package pkg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/google"
)

func TestApplication_UseGoogleProvider(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := `
server:
  port: 8080
security:
  oauth2:
    client:
      google:
        client-id: google-client-id
        client-secret: google-client-secret
        redirect-uri: http://localhost:8080/auth/google/callback
        scopes:
          - email
          - profile
          - openid
        state: random-state
        user-info-uri: https://www.googleapis.com/oauth2/v2/userinfo
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	app := &Application{}
	app.UseConfig()

	result := app.UseGoogleProvider()

	assert.Same(t, app, result) // Returns same instance for chaining
	assert.NotNil(t, GoogleOAuth2Config)
}

func TestApplication_UseGoogleProvider_SetsClientID(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := `
server:
  port: 8080
security:
  oauth2:
    client:
      google:
        client-id: my-google-client-id
        client-secret: my-google-secret
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

	app := &Application{}
	app.UseConfig()
	app.UseGoogleProvider()

	assert.Equal(t, "my-google-client-id", GoogleOAuth2Config.ClientID)
}

func TestApplication_UseGoogleProvider_SetsClientSecret(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := `
server:
  port: 8080
security:
  oauth2:
    client:
      google:
        client-id: client-id
        client-secret: super-secret-key
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

	app := &Application{}
	app.UseConfig()
	app.UseGoogleProvider()

	assert.Equal(t, "super-secret-key", GoogleOAuth2Config.ClientSecret)
}

func TestApplication_UseGoogleProvider_SetsRedirectURL(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := `
server:
  port: 8080
security:
  oauth2:
    client:
      google:
        client-id: client-id
        client-secret: secret
        redirect-uri: https://myapp.com/auth/callback
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

	app := &Application{}
	app.UseConfig()
	app.UseGoogleProvider()

	assert.Equal(t, "https://myapp.com/auth/callback", GoogleOAuth2Config.RedirectURL)
}

func TestApplication_UseGoogleProvider_SetsScopes(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := `
server:
  port: 8080
security:
  oauth2:
    client:
      google:
        client-id: client-id
        client-secret: secret
        redirect-uri: http://localhost/callback
        scopes:
          - email
          - profile
          - openid
        state: state
        user-info-uri: https://example.com/userinfo
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	app := &Application{}
	app.UseConfig()
	app.UseGoogleProvider()

	assert.Equal(t, []string{"email", "profile", "openid"}, GoogleOAuth2Config.Scopes)
}

func TestApplication_UseGoogleProvider_UsesGoogleEndpoint(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := `
server:
  port: 8080
security:
  oauth2:
    client:
      google:
        client-id: client-id
        client-secret: secret
        redirect-uri: http://localhost/callback
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

	app := &Application{}
	app.UseConfig()
	app.UseGoogleProvider()

	assert.Equal(t, google.Endpoint, GoogleOAuth2Config.Endpoint)
}

func TestApplication_UseGoogleProvider_ChainedCalls(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := `
server:
  port: 8080
security:
  oauth2:
    client:
      google:
        client-id: chain-client-id
        client-secret: chain-secret
        redirect-uri: http://localhost/callback
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

	app := &Application{}
	result := app.UseConfig().UseGoogleProvider()

	assert.Same(t, app, result)
	assert.Equal(t, "chain-client-id", GoogleOAuth2Config.ClientID)
}
