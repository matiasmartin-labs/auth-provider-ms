package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/application/ports"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/domain/model"
	"github.com/matiasmartin-labs/auth-provider-ms/pkg"
)

type JwtGenerator struct{}

func NewJwtGenerator() ports.TokenGenerator {
	return &JwtGenerator{}
}

func (jg *JwtGenerator) GenerateToken(userInfo *model.UserInfo) (string, error) {
	privateKey := pkg.App.KeyPair.GetPrivateKey()
	jwtConfig := pkg.App.Config.GetSecurityConfig().GetJWTConfig()
	claims := pkg.Claims{
		Roles:   []string{"USER"},
		Name:    userInfo.GetFullName(),
		Email:   userInfo.Email,
		Picture: userInfo.Picture,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    jwtConfig.GetIssuer(),
			Subject:   userInfo.Email,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtConfig.GetExpirationTime())),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
