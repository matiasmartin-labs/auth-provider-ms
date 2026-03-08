package token

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/domain/model"
	"github.com/matiasmartin-labs/auth-provider-ms/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockKeyPair struct {
	privateKey *rsa.PrivateKey
}

func (m *mockKeyPair) PublicJWK() (map[string]interface{}, error) { return nil, nil }
func (m *mockKeyPair) GetPrivateKey() *rsa.PrivateKey             { return m.privateKey }

type mockJWTConfig struct {
	issuer         string
	audience       string
	expirationTime time.Duration
}

func (m *mockJWTConfig) GetIssuer() string                { return m.issuer }
func (m *mockJWTConfig) GetAudience() string              { return m.audience }
func (m *mockJWTConfig) GetExpirationTime() time.Duration { return m.expirationTime }

type mockSecurityConfig struct {
	jwtConfig *mockJWTConfig
}

func (m *mockSecurityConfig) GetOAuth2Config() pkg.OAuth2Config     { return nil }
func (m *mockSecurityConfig) GetRedirectConfig() pkg.RedirectConfig { return nil }
func (m *mockSecurityConfig) GetCookieConfig() pkg.CookieConfig     { return nil }
func (m *mockSecurityConfig) GetLoginConfig() pkg.LoginConfig       { return nil }
func (m *mockSecurityConfig) GetJWTConfig() pkg.JWTConfig           { return m.jwtConfig }
func (m *mockSecurityConfig) GetAuthConfig() pkg.AuthConfig         { return nil }

type mockConfiguration struct {
	securityConfig *mockSecurityConfig
}

func (m *mockConfiguration) GetServerConfig() pkg.ServerConfig     { return nil }
func (m *mockConfiguration) GetSecurityConfig() pkg.SecurityConfig { return m.securityConfig }

func setupMockApp() *rsa.PrivateKey {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	pkg.App = &pkg.Application{
		Config: &mockConfiguration{
			securityConfig: &mockSecurityConfig{
				jwtConfig: &mockJWTConfig{
					issuer:         "test-issuer",
					audience:       "test-audience",
					expirationTime: time.Hour,
				},
			},
		},
		KeyPair: &mockKeyPair{privateKey: privateKey},
	}

	return privateKey
}

func TestNewJwtGenerator(t *testing.T) {
	generator := NewJwtGenerator()
	assert.NotNil(t, generator)
}

func TestJwtGenerator_GenerateToken_Success(t *testing.T) {
	privateKey := setupMockApp()

	generator := NewJwtGenerator()

	userInfo := &model.UserInfo{
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Picture:   "https://example.com/pic.jpg",
	}

	token, err := generator.GenerateToken(userInfo)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedToken, err := jwt.ParseWithClaims(token, &pkg.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})

	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	parsedClaims, ok := parsedToken.Claims.(*pkg.Claims)
	require.True(t, ok)

	assert.Equal(t, "test@example.com", parsedClaims.Email)
	assert.Equal(t, "John Doe", parsedClaims.Name)
	assert.Equal(t, "https://example.com/pic.jpg", parsedClaims.Picture)
	assert.Equal(t, []string{"USER"}, parsedClaims.Roles)
	assert.Equal(t, "test-issuer", parsedClaims.Issuer)
	assert.Equal(t, "test@example.com", parsedClaims.Subject)
}

func TestJwtGenerator_GenerateToken_ContainsCorrectExpiration(t *testing.T) {
	privateKey := setupMockApp()

	generator := NewJwtGenerator()

	userInfo := &model.UserInfo{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	token, err := generator.GenerateToken(userInfo)
	require.NoError(t, err)

	parsedToken, err := jwt.ParseWithClaims(token, &pkg.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})

	require.NoError(t, err)

	parsedClaims := parsedToken.Claims.(*pkg.Claims)

	expectedExpiration := time.Now().Add(time.Hour)
	actualExpiration := parsedClaims.ExpiresAt.Time

	assert.WithinDuration(t, expectedExpiration, actualExpiration, 5*time.Second)
}

func TestJwtGenerator_GenerateToken_HasUniqueID(t *testing.T) {
	privateKey := setupMockApp()

	generator := NewJwtGenerator()

	userInfo := &model.UserInfo{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	token1, _ := generator.GenerateToken(userInfo)
	token2, _ := generator.GenerateToken(userInfo)

	parsedToken1, _ := jwt.ParseWithClaims(token1, &pkg.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})
	parsedToken2, _ := jwt.ParseWithClaims(token2, &pkg.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})

	claims1 := parsedToken1.Claims.(*pkg.Claims)
	claims2 := parsedToken2.Claims.(*pkg.Claims)

	assert.NotEmpty(t, claims1.ID)
	assert.NotEmpty(t, claims2.ID)
	assert.NotEqual(t, claims1.ID, claims2.ID)
}

func TestJwtGenerator_GenerateToken_UsesRS256(t *testing.T) {
	setupMockApp()

	generator := NewJwtGenerator()

	userInfo := &model.UserInfo{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	token, err := generator.GenerateToken(userInfo)
	require.NoError(t, err)

	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, &pkg.Claims{})
	require.NoError(t, err)

	assert.Equal(t, "RS256", parsedToken.Method.Alg())
}

func TestJwtGenerator_GenerateToken_FullName(t *testing.T) {
	privateKey := setupMockApp()

	generator := NewJwtGenerator()

	testCases := []struct {
		firstName    string
		lastName     string
		expectedName string
	}{
		{"John", "Doe", "John Doe"},
		{"María", "García", "María García"},
		{"", "Smith", " Smith"},
		{"Jane", "", "Jane "},
	}

	for _, tc := range testCases {
		userInfo := &model.UserInfo{
			Email:     "test@example.com",
			FirstName: tc.firstName,
			LastName:  tc.lastName,
		}

		token, err := generator.GenerateToken(userInfo)
		require.NoError(t, err)

		parsedToken, _ := jwt.ParseWithClaims(token, &pkg.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return &privateKey.PublicKey, nil
		})

		parsedClaims := parsedToken.Claims.(*pkg.Claims)
		assert.Equal(t, tc.expectedName, parsedClaims.Name)
	}
}
