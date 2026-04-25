package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/matiasmartin-labs/auth-provider-ms/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createServerTestConfig(t *testing.T) (cleanup func()) {
	t.Helper()

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
    enabled: false
    url: ""
  cookie:
    secure: false
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

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))

	return func() {
		require.NoError(t, os.Chdir(originalDir))
	}
}

func TestRoutes_RegistersSignOutEndpoint(t *testing.T) {
	cleanup := createServerTestConfig(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	app := pkg.NewApplication().
		UseConfig().
		UseServer().
		UseServerSecurity()

	err := Routes(app)
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-out", nil)
	response := httptest.NewRecorder()

	app.Server.Handler.ServeHTTP(response, request)

	assert.NotEqual(t, http.StatusNotFound, response.Code)
	assert.Equal(t, http.StatusNoContent, response.Code)
}
