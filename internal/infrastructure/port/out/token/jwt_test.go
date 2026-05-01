package token

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig(privateKey *rsa.PrivateKey) JwtGeneratorConfig {
	return JwtGeneratorConfig{
		PrivateKey:     privateKey,
		Issuer:         "test-issuer",
		Audience:       "test-audience",
		ExpirationTime: time.Hour,
	}
}

func newTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return privateKey
}

func TestNewJwtGenerator(t *testing.T) {
	privateKey := newTestKey(t)
	generator := NewJwtGenerator(newTestConfig(privateKey))
	assert.NotNil(t, generator)
}

func TestJwtGenerator_GenerateToken_Success(t *testing.T) {
	privateKey := newTestKey(t)
	generator := NewJwtGenerator(newTestConfig(privateKey))

	userInfo := &model.UserInfo{
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Picture:   "https://example.com/pic.jpg",
	}

	token, err := generator.GenerateToken(userInfo)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedToken, err := jwt.ParseWithClaims(token, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})

	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	parsedClaims, ok := parsedToken.Claims.(*tokenClaims)
	require.True(t, ok)

	assert.Equal(t, "test@example.com", parsedClaims.Email)
	assert.Equal(t, "John Doe", parsedClaims.Name)
	assert.Equal(t, "https://example.com/pic.jpg", parsedClaims.Picture)
	assert.Equal(t, []string{"USER"}, parsedClaims.Roles)
	assert.Equal(t, "test-issuer", parsedClaims.Issuer)
	assert.Equal(t, "test@example.com", parsedClaims.Subject)
}

func TestJwtGenerator_GenerateToken_ContainsCorrectExpiration(t *testing.T) {
	privateKey := newTestKey(t)
	generator := NewJwtGenerator(newTestConfig(privateKey))

	userInfo := &model.UserInfo{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	token, err := generator.GenerateToken(userInfo)
	require.NoError(t, err)

	parsedToken, err := jwt.ParseWithClaims(token, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})

	require.NoError(t, err)

	parsedClaims := parsedToken.Claims.(*tokenClaims)

	expectedExpiration := time.Now().Add(time.Hour)
	actualExpiration := parsedClaims.ExpiresAt.Time

	assert.WithinDuration(t, expectedExpiration, actualExpiration, 5*time.Second)
}

func TestJwtGenerator_GenerateToken_HasUniqueID(t *testing.T) {
	privateKey := newTestKey(t)
	generator := NewJwtGenerator(newTestConfig(privateKey))

	userInfo := &model.UserInfo{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	token1, _ := generator.GenerateToken(userInfo)
	token2, _ := generator.GenerateToken(userInfo)

	parsedToken1, _ := jwt.ParseWithClaims(token1, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})
	parsedToken2, _ := jwt.ParseWithClaims(token2, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})

	claims1 := parsedToken1.Claims.(*tokenClaims)
	claims2 := parsedToken2.Claims.(*tokenClaims)

	assert.NotEmpty(t, claims1.ID)
	assert.NotEmpty(t, claims2.ID)
	assert.NotEqual(t, claims1.ID, claims2.ID)
}

func TestJwtGenerator_GenerateToken_UsesRS256(t *testing.T) {
	privateKey := newTestKey(t)
	generator := NewJwtGenerator(newTestConfig(privateKey))

	userInfo := &model.UserInfo{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	token, err := generator.GenerateToken(userInfo)
	require.NoError(t, err)

	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, &tokenClaims{})
	require.NoError(t, err)

	assert.Equal(t, "RS256", parsedToken.Method.Alg())
}

func TestJwtGenerator_GenerateToken_FullName(t *testing.T) {
	privateKey := newTestKey(t)
	generator := NewJwtGenerator(newTestConfig(privateKey))

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

		parsedToken, _ := jwt.ParseWithClaims(token, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			return &privateKey.PublicKey, nil
		})

		parsedClaims := parsedToken.Claims.(*tokenClaims)
		assert.Equal(t, tc.expectedName, parsedClaims.Name)
	}
}

func TestJwtGenerator_GenerateToken_ContainsAudience(t *testing.T) {
	privateKey := newTestKey(t)
	generator := NewJwtGenerator(newTestConfig(privateKey))

	userInfo := &model.UserInfo{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	token, err := generator.GenerateToken(userInfo)
	require.NoError(t, err)

	parsedToken, err := jwt.ParseWithClaims(token, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})

	require.NoError(t, err)

	parsedClaims := parsedToken.Claims.(*tokenClaims)

	require.NotNil(t, parsedClaims.Audience)
	assert.Len(t, parsedClaims.Audience, 1)
	assert.Equal(t, "test-audience", parsedClaims.Audience[0])
}

func TestJwtGenerator_GenerateToken_ValidatableByMiddleware(t *testing.T) {
	privateKey := newTestKey(t)
	generator := NewJwtGenerator(newTestConfig(privateKey))

	userInfo := &model.UserInfo{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	token, err := generator.GenerateToken(userInfo)
	require.NoError(t, err)

	_, err = jwt.ParseWithClaims(token, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return &privateKey.PublicKey, nil
	},
		jwt.WithIssuer("test-issuer"),
		jwt.WithAudience("test-audience"),
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}),
	)

	assert.NoError(t, err)
}

func TestJwtGenerator_GenerateToken_FailsValidationWithWrongIssuer(t *testing.T) {
	privateKey := newTestKey(t)
	generator := NewJwtGenerator(newTestConfig(privateKey))

	userInfo := &model.UserInfo{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	token, err := generator.GenerateToken(userInfo)
	require.NoError(t, err)

	_, err = jwt.ParseWithClaims(token, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	},
		jwt.WithIssuer("wrong-issuer"),
		jwt.WithAudience("test-audience"),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "issuer")
}

func TestJwtGenerator_GenerateToken_FailsValidationWithWrongAudience(t *testing.T) {
	privateKey := newTestKey(t)
	generator := NewJwtGenerator(newTestConfig(privateKey))

	userInfo := &model.UserInfo{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	token, err := generator.GenerateToken(userInfo)
	require.NoError(t, err)

	_, err = jwt.ParseWithClaims(token, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	},
		jwt.WithIssuer("test-issuer"),
		jwt.WithAudience("wrong-audience"),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audience")
}

func TestJwtGenerator_GenerateToken_FailsValidationWithMissingAudience(t *testing.T) {
	privateKey := newTestKey(t)
	generator := NewJwtGenerator(JwtGeneratorConfig{
		PrivateKey:     privateKey,
		Issuer:         "test-issuer",
		Audience:       "",
		ExpirationTime: time.Hour,
	})

	userInfo := &model.UserInfo{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	token, err := generator.GenerateToken(userInfo)
	require.NoError(t, err)

	_, err = jwt.ParseWithClaims(token, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	},
		jwt.WithIssuer("test-issuer"),
		jwt.WithAudience("expected-audience"),
	)

	assert.Error(t, err)
}
