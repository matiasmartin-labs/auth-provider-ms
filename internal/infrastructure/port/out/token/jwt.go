package token

import (
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	fwkclaims "github.com/matiasmartin-labs/common-fwk/security/claims"

	"github.com/matiasmartin-labs/auth-provider-ms/internal/application/ports"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/domain/model"
)

// JwtGeneratorConfig holds the signing configuration for the JWT generator.
type JwtGeneratorConfig struct {
	PrivateKey     *rsa.PrivateKey
	Issuer         string
	Audience       string
	ExpirationTime time.Duration
}

type jwtGenerator struct {
	cfg JwtGeneratorConfig
}

// NewJwtGenerator returns a TokenGenerator using RS256 and the provided config.
func NewJwtGenerator(cfg JwtGeneratorConfig) ports.TokenGenerator {
	return &jwtGenerator{cfg: cfg}
}

// tokenClaims holds the RS256 JWT payload including OIDC profile fields.
type tokenClaims struct {
	Roles   []string `json:"roles"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Picture string   `json:"picture"`
	jwt.RegisteredClaims
}

// GenerateToken signs a JWT for the given user using RS256.
func (jg *jwtGenerator) GenerateToken(userInfo *model.UserInfo) (string, error) {
	claims := tokenClaims{
		Roles:   []string{"USER"},
		Name:    userInfo.GetFullName(),
		Email:   userInfo.Email,
		Picture: userInfo.Picture,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    jg.cfg.Issuer,
			Subject:   userInfo.Email,
			Audience:  jwt.ClaimStrings{jg.cfg.Audience},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jg.cfg.ExpirationTime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(jg.cfg.PrivateKey)
}

// ToClaims converts tokenClaims into a common-fwk claims.Claims value.
// This is used by handlers that need to read OIDC profile fields after validation.
func ToClaims(src *tokenClaims) fwkclaims.Claims {
	return fwkclaims.Claims{
		Subject: src.Subject,
		Email:   src.Email,
		Name:    src.Name,
		Picture: src.Picture,
		Roles:   src.Roles,
	}
}
