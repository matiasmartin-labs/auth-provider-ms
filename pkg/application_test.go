package pkg

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestConfig(t *testing.T) (cleanup func()) {
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
	os.Chdir(tmpDir)

	return func() {
		os.Chdir(originalDir)
	}
}

func TestNewApplication(t *testing.T) {
	app := NewApplication()

	assert.NotNil(t, app)
	assert.NotNil(t, App) // Global variable should be set
	assert.Same(t, app, App)
}

func TestNewApplication_SetsGlobalApp(t *testing.T) {
	App = nil

	app := NewApplication()

	assert.Same(t, app, App)
}

func TestApplication_UseConfig(t *testing.T) {
	cleanup := createTestConfig(t)
	defer cleanup()

	app := &Application{}
	result := app.UseConfig()

	assert.Same(t, app, result) // Returns same instance for chaining
	assert.NotNil(t, app.Config)
	assert.Equal(t, 8080, app.Config.GetServerConfig().GetPort())
}

func TestApplication_UseServer(t *testing.T) {
	cleanup := createTestConfig(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	app := &Application{}
	app.UseConfig()
	result := app.UseServer()

	assert.Same(t, app, result) // Returns same instance for chaining
	assert.NotNil(t, app.Server)
	assert.Equal(t, ":8080", app.Server.Addr)
	assert.Equal(t, 10*time.Second, app.Server.ReadTimeout)
	assert.Equal(t, 15*time.Second, app.Server.WriteTimeout)
	assert.Equal(t, 1048576, app.Server.MaxHeaderBytes)
}

func TestApplication_RegisterGET(t *testing.T) {
	cleanup := createTestConfig(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	app := &Application{}
	app.UseConfig()
	app.UseServer()

	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	}

	app.RegisterGET("/test", handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	app.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test")
}

func TestApplication_RegisterProtectedGET(t *testing.T) {
	cleanup := createTestConfig(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	app := &Application{}
	app.UseConfig()
	app.UseServer()
	app.UseServerSecurity()

	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "protected"})
	}

	app.RegisterProtectedGET("/protected", handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	app.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestApplication_Run_WithSuccessfulEntryPoint(t *testing.T) {
	cleanup := createTestConfig(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	app := &Application{}
	app.UseConfig()
	app.UseServer()

	entryPointCalled := false
	entryPoint := func(a *Application) error {
		entryPointCalled = true
		assert.Same(t, app, a)
		return nil
	}

	go func() {
		app.Run(entryPoint)
	}()

	time.Sleep(50 * time.Millisecond)

	assert.True(t, entryPointCalled)
}

func TestApplication_Run_WithFailingEntryPoint(t *testing.T) {
	cleanup := createTestConfig(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	app := &Application{}
	app.UseConfig()
	app.UseServer()

	expectedErr := assert.AnError
	entryPoint := func(a *Application) error {
		return expectedErr
	}

	err := app.Run(entryPoint)

	assert.Equal(t, expectedErr, err)
}

func TestApplication_ChainedMethodCalls(t *testing.T) {
	cleanup := createTestConfig(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	app := NewApplication().
		UseConfig().
		UseServer().
		UseServerSecurity()

	assert.NotNil(t, app.Config)
	assert.NotNil(t, app.Server.Handler)
	assert.NotNil(t, app.KeyPair)
}
