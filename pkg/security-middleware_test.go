package pkg

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decodeAuthErrorPayload(t *testing.T, body []byte) map[string]string {
	t.Helper()

	var payload map[string]string
	err := json.Unmarshal(body, &payload)
	require.NoError(t, err)

	return payload
}

func init() {
	gin.SetMode(gin.TestMode)
}

type mockAuthConfig struct {
	enabled bool
}

func (m *mockAuthConfig) IsEnabled() bool {
	return m.enabled
}

type mockJWTConfig struct {
	issuer         string
	audience       string
	expirationTime time.Duration
}

func (m *mockJWTConfig) GetIssuer() string                { return m.issuer }
func (m *mockJWTConfig) GetAudience() string              { return m.audience }
func (m *mockJWTConfig) GetExpirationTime() time.Duration { return m.expirationTime }

type mockSecurityConfig struct {
	authConfig *mockAuthConfig
	jwtConfig  *mockJWTConfig
}

func (m *mockSecurityConfig) GetOAuth2Config() OAuth2Config     { return nil }
func (m *mockSecurityConfig) GetRedirectConfig() RedirectConfig { return nil }
func (m *mockSecurityConfig) GetCookieConfig() CookieConfig     { return nil }
func (m *mockSecurityConfig) GetLoginConfig() LoginConfig       { return nil }
func (m *mockSecurityConfig) GetJWTConfig() JWTConfig           { return m.jwtConfig }
func (m *mockSecurityConfig) GetAuthConfig() AuthConfig         { return m.authConfig }

type mockConfiguration struct {
	securityConfig *mockSecurityConfig
}

func (m *mockConfiguration) GetServerConfig() ServerConfig     { return nil }
func (m *mockConfiguration) GetSecurityConfig() SecurityConfig { return m.securityConfig }

type mockKeyPair struct {
	privateKey *rsa.PrivateKey
}

func (m *mockKeyPair) PublicJWK() (map[string]interface{}, error) { return nil, nil }
func (m *mockKeyPair) GetPrivateKey() *rsa.PrivateKey             { return m.privateKey }

func setupTestApp(authEnabled bool, issuer, audience string) (*Application, *rsa.PrivateKey) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	app := &Application{
		Config: &mockConfiguration{
			securityConfig: &mockSecurityConfig{
				authConfig: &mockAuthConfig{enabled: authEnabled},
				jwtConfig: &mockJWTConfig{
					issuer:         issuer,
					audience:       audience,
					expirationTime: time.Hour,
				},
			},
		},
		KeyPair: &mockKeyPair{privateKey: privateKey},
	}

	App = app
	return app, privateKey
}

func generateTestToken(privateKey *rsa.PrivateKey, claims Claims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, _ := token.SignedString(privateKey)
	return signedToken
}

func TestAuthMiddleware_AuthDisabled(t *testing.T) {
	app, _ := setupTestApp(false, "test-issuer", "test-audience")

	router := gin.New()
	router.GET("/test", app.AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	app, _ := setupTestApp(true, "test-issuer", "test-audience")

	router := gin.New()
	router.GET("/test", app.AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	payload := decodeAuthErrorPayload(t, w.Body.Bytes())
	assert.Equal(t, AuthCodeTokenMissing, payload["code"])
	assert.Equal(t, "missing authentication token", payload["message"])
	assert.Len(t, payload, 2)
	_, hasLegacyError := payload["error"]
	assert.False(t, hasLegacyError)
}

func TestAuthMiddleware_ValidTokenInHeader(t *testing.T) {
	app, privateKey := setupTestApp(true, "test-issuer", "test-audience")

	claims := Claims{
		Name:    "Test User",
		Email:   "test@example.com",
		Picture: "https://example.com/pic.jpg",
		Roles:   []string{"USER"},
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    "test-issuer",
			Subject:   "test@example.com",
			Audience:  jwt.ClaimStrings{"test-audience"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := generateTestToken(privateKey, claims)

	router := gin.New()
	router.GET("/test", app.AuthMiddleware(), func(c *gin.Context) {
		claimsFromCtx, exists := c.Get("claims")
		require.True(t, exists)

		parsedClaims := claimsFromCtx.(*Claims)
		c.JSON(http.StatusOK, gin.H{
			"email": parsedClaims.Email,
			"name":  parsedClaims.Name,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test@example.com")
	assert.Contains(t, w.Body.String(), "Test User")
}

func TestAuthMiddleware_ValidTokenInCookie(t *testing.T) {
	app, privateKey := setupTestApp(true, "test-issuer", "test-audience")

	claims := Claims{
		Name:    "Cookie User",
		Email:   "cookie@example.com",
		Picture: "https://example.com/cookie-pic.jpg",
		Roles:   []string{"USER"},
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    "test-issuer",
			Subject:   "cookie@example.com",
			Audience:  jwt.ClaimStrings{"test-audience"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := generateTestToken(privateKey, claims)

	router := gin.New()
	router.GET("/test", app.AuthMiddleware(), func(c *gin.Context) {
		claimsFromCtx, _ := c.Get("claims")
		parsedClaims := claimsFromCtx.(*Claims)
		c.JSON(http.StatusOK, gin.H{"email": parsedClaims.Email})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "cookie@example.com")
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	app, privateKey := setupTestApp(true, "test-issuer", "test-audience")

	claims := Claims{
		Name:  "Expired User",
		Email: "expired@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    "test-issuer",
			Subject:   "expired@example.com",
			Audience:  jwt.ClaimStrings{"test-audience"},
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expirado hace 1 hora
		},
	}

	token := generateTestToken(privateKey, claims)

	router := gin.New()
	router.GET("/test", app.AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	payload := decodeAuthErrorPayload(t, w.Body.Bytes())
	assert.Equal(t, AuthCodeTokenInvalid, payload["code"])
	assert.Equal(t, "invalid or expired token", payload["message"])
	assert.Len(t, payload, 2)
	_, hasLegacyError := payload["error"]
	assert.False(t, hasLegacyError)
}

func TestAuthMiddleware_InvalidIssuer(t *testing.T) {
	app, privateKey := setupTestApp(true, "test-issuer", "test-audience")

	claims := Claims{
		Name:  "Wrong Issuer User",
		Email: "wrong@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    "wrong-issuer", // Issuer incorrecto
			Subject:   "wrong@example.com",
			Audience:  jwt.ClaimStrings{"test-audience"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := generateTestToken(privateKey, claims)

	router := gin.New()
	router.GET("/test", app.AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidAudience(t *testing.T) {
	app, privateKey := setupTestApp(true, "test-issuer", "test-audience")

	claims := Claims{
		Name:  "Wrong Audience User",
		Email: "wrong@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    "test-issuer",
			Subject:   "wrong@example.com",
			Audience:  jwt.ClaimStrings{"wrong-audience"}, // Audience incorrecto
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := generateTestToken(privateKey, claims)

	router := gin.New()
	router.GET("/test", app.AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_MalformedToken(t *testing.T) {
	app, _ := setupTestApp(true, "test-issuer", "test-audience")

	router := gin.New()
	router.GET("/test", app.AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-format")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	payload := decodeAuthErrorPayload(t, w.Body.Bytes())
	assert.Equal(t, AuthCodeTokenInvalid, payload["code"])
	assert.Equal(t, "invalid or expired token", payload["message"])
	assert.Len(t, payload, 2)
	_, hasLegacyError := payload["error"]
	assert.False(t, hasLegacyError)
}

func TestAuthMiddleware_InvalidAuthorizationFormat(t *testing.T) {
	app, _ := setupTestApp(true, "test-issuer", "test-audience")

	router := gin.New()
	router.GET("/test", app.AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "just-a-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	payload := decodeAuthErrorPayload(t, w.Body.Bytes())
	assert.Equal(t, AuthCodeTokenMissing, payload["code"])
	assert.Equal(t, "missing authentication token", payload["message"])
	assert.Len(t, payload, 2)
	_, hasLegacyError := payload["error"]
	assert.False(t, hasLegacyError)
}

func TestAuthMiddleware_HeaderTakesPrecedenceOverCookie(t *testing.T) {
	app, privateKey := setupTestApp(true, "test-issuer", "test-audience")

	headerClaims := Claims{
		Name:  "Header User",
		Email: "header@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    "test-issuer",
			Subject:   "header@example.com",
			Audience:  jwt.ClaimStrings{"test-audience"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	headerToken := generateTestToken(privateKey, headerClaims)

	cookieClaims := Claims{
		Name:  "Cookie User",
		Email: "cookie@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    "test-issuer",
			Subject:   "cookie@example.com",
			Audience:  jwt.ClaimStrings{"test-audience"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	cookieToken := generateTestToken(privateKey, cookieClaims)

	router := gin.New()
	router.GET("/test", app.AuthMiddleware(), func(c *gin.Context) {
		claimsFromCtx, _ := c.Get("claims")
		parsedClaims := claimsFromCtx.(*Claims)
		c.JSON(http.StatusOK, gin.H{"email": parsedClaims.Email})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+headerToken)
	req.AddCookie(&http.Cookie{Name: "token", Value: cookieToken})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "header@example.com")
	assert.NotContains(t, w.Body.String(), "cookie@example.com")
}

func TestExtractToken_FromHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer my-test-token")

	token := extractToken(c)

	assert.Equal(t, "my-test-token", token)
}

func TestExtractToken_FromCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.AddCookie(&http.Cookie{Name: "token", Value: "cookie-token"})

	token := extractToken(c)

	assert.Equal(t, "cookie-token", token)
}

func TestExtractToken_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	token := extractToken(c)

	assert.Empty(t, token)
}

func TestExtractToken_BearerCaseInsensitive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "BEARER my-token")

	token := extractToken(c)

	assert.Equal(t, "my-token", token)
}
