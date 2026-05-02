package server

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	fwkapp "github.com/matiasmartin-labs/common-fwk/app"
	fwkconfig "github.com/matiasmartin-labs/common-fwk/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestApp builds a fully wired Application for route tests using RS256 with
// generated keys, mirroring the production config path.
func newTestApp(t *testing.T) *fwkapp.Application {
	t.Helper()

	gin.SetMode(gin.TestMode)

	cfg := fwkconfig.Config{
		Server: fwkconfig.ServerConfig{
			Host:           "127.0.0.1",
			Port:           8080,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		Security: fwkconfig.SecurityConfig{
			Auth: fwkconfig.AuthConfig{
				JWT: fwkconfig.JWTConfig{
					Algorithm:  "RS256",
					Issuer:     "test-issuer",
					TTLMinutes: 60,
					RS256: fwkconfig.RS256Config{
						KeySource: fwkconfig.RS256KeySourceGenerated,
						KeyID:     "test-key",
					},
				},
				Cookie: fwkconfig.CookieConfig{
					Name:     "token",
					Secure:   false,
					HTTPOnly: true,
					SameSite: "Lax",
				},
				OAuth2: fwkconfig.OAuth2Config{
					Providers: map[string]fwkconfig.OAuth2ProviderConfig{
						"google": {
							ClientID:     "test-client-id",
							ClientSecret: "test-secret",
							RedirectURL:  "http://localhost:8080/callback",
							AuthURL:      "https://accounts.google.com/o/oauth2/auth",
							TokenURL:     "https://oauth2.googleapis.com/token",
							Scopes:       []string{"email", "profile"},
						},
					},
				},
				Login: fwkconfig.LoginConfig{
					Email: "test@example.com",
				},
			},
		},
	}

	app, err := fwkapp.NewApplication().
		UseConfig(cfg).
		UseServer().
		UseServerSecurityFromConfig()
	require.NoError(t, err)

	return app
}

func TestRoutes_RegistersSignOutEndpoint(t *testing.T) {
	app := newTestApp(t)

	err := Routes(app)
	require.NoError(t, err)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	serverDone := make(chan error, 1)
	go func() { serverDone <- app.RunListener(ln) }()
	t.Cleanup(func() { _ = ln.Close() })

	addr := ln.Addr().String()
	resp, err := http.Post("http://"+addr+"/api/v1/auth/sign-out", "", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestRoutes_RegistersJWKSEndpoint(t *testing.T) {
	app := newTestApp(t)

	err := Routes(app)
	require.NoError(t, err)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go func() { _ = app.RunListener(ln) }()
	t.Cleanup(func() { _ = ln.Close() })

	addr := ln.Addr().String()
	resp, err := http.Get("http://" + addr + "/.well-known/jwks.json")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRoutes_ReturnsError_WhenServerNotReady(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Application without UseServer() — server not ready.
	app := fwkapp.NewApplication()

	err := Routes(app)
	assert.Error(t, err)
}

// httpRecorder creates a single-request test server and returns a recorder with the response code.
func httpRecorder(t *testing.T, app *fwkapp.Application, method, path string) *httptest.ResponseRecorder {
	t.Helper()

	err := Routes(app)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	ln, lnErr := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, lnErr)
	go func() { _ = app.RunListener(ln) }()
	t.Cleanup(func() { _ = ln.Close() })

	req, err := http.NewRequest(method, "http://"+ln.Addr().String()+path, nil)
	require.NoError(t, err)

	httpResp, err := (&http.Client{}).Do(req)
	require.NoError(t, err)
	defer httpResp.Body.Close()

	w.Code = httpResp.StatusCode

	return w
}

func TestRoutes_RegistersGoogleLoginEndpoint(t *testing.T) {
	app := newTestApp(t)
	w := httpRecorder(t, app, http.MethodGet, "/oauth2/authorization/google")
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}
