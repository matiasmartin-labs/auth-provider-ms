package server

import (
	"crypto/rand"
	"crypto/rsa"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	fwkapp "github.com/matiasmartin-labs/common-fwk/app"
	fwkconfig "github.com/matiasmartin-labs/common-fwk/config"
	fwkjwt "github.com/matiasmartin-labs/common-fwk/security/jwt"
	fwkkeys "github.com/matiasmartin-labs/common-fwk/security/keys"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/in/jwks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestBootstrap(t *testing.T) (*fwkapp.Application, *Bootstrap) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyPair := jwks.NewKeyPair(privateKey)

	cfg := fwkconfig.Config{
		Server: fwkconfig.ServerConfig{Host: "127.0.0.1", Port: 8080},
		Security: fwkconfig.SecurityConfig{
			Auth: fwkconfig.AuthConfig{
				JWT: fwkconfig.JWTConfig{
					Issuer:     "test-issuer",
					TTLMinutes: 60,
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

	resolver := fwkkeys.NewRSAResolver(privateKey, keyPair.KeyID)
	validator, err := fwkjwt.NewValidator(fwkjwt.Options{
		Methods:  []string{"RS256"},
		Issuer:   "test-issuer",
		Resolver: resolver,
	})
	require.NoError(t, err)

	b := &Bootstrap{
		PrivateKey: privateKey,
		KeyPair:    keyPair,
		Config:     cfg,
	}

	app := fwkapp.NewApplication().
		UseConfig(cfg).
		UseServer().
		UseServerSecurity(validator)

	return app, b
}

func TestRoutes_RegistersSignOutEndpoint(t *testing.T) {
	app, b := newTestBootstrap(t)

	err := Routes(app, b)
	require.NoError(t, err)

	// Use a test listener so we can exercise the actual handler via RunListener.
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
	app, b := newTestBootstrap(t)

	err := Routes(app, b)
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

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	b := &Bootstrap{
		PrivateKey: privateKey,
		KeyPair:    jwks.NewKeyPair(privateKey),
		Config:     fwkconfig.Config{},
	}

	// Application without UseServer() — server not ready.
	app := fwkapp.NewApplication()

	err = Routes(app, b)
	assert.Error(t, err)
}

// httpRecorder creates a single-request test server without the net/http/httptest.Server overhead.
func httpRecorder(t *testing.T, app *fwkapp.Application, b *Bootstrap, method, path string) *httptest.ResponseRecorder {
	t.Helper()

	err := Routes(app, b)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	// Launch a test listener to exercise the actual routes.
	ln, lnErr := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, lnErr)
	go func() { _ = app.RunListener(ln) }()
	t.Cleanup(func() { _ = ln.Close() })

	resp, err := http.NewRequest(method, "http://"+ln.Addr().String()+path, nil)
	require.NoError(t, err)

	client := &http.Client{}
	httpResp, err := client.Do(resp)
	require.NoError(t, err)
	defer httpResp.Body.Close()

	w.Code = httpResp.StatusCode

	return w
}

func TestRoutes_RegistersGoogleLoginEndpoint(t *testing.T) {
	app, b := newTestBootstrap(t)
	w := httpRecorder(t, app, b, http.MethodGet, "/oauth2/authorization/google")
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}
