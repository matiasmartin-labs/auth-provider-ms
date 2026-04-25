package pkg

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Roles   []string `json:"roles"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Picture string   `json:"picture"`
	jwt.RegisteredClaims
}

func (app *Application) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authConfig := app.Config.GetSecurityConfig().GetAuthConfig()
		if authConfig == nil || !authConfig.IsEnabled() {
			c.Next()
			return
		}

		token := extractToken(c)
		if token == "" {
			WriteAuthError(c, http.StatusUnauthorized, AuthCodeTokenMissing, "missing authentication token")
			c.Abort()
			return
		}

		claims, err := app.validateToken(token)
		if err != nil {
			WriteAuthError(c, http.StatusUnauthorized, AuthCodeTokenInvalid, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}

	if cookie, err := c.Cookie("token"); err == nil && cookie != "" {
		return cookie
	}

	return ""
}

func (app *Application) validateToken(tokenString string) (*Claims, error) {
	jwtConfig := app.Config.GetSecurityConfig().GetJWTConfig()
	publicKey := app.KeyPair.GetPrivateKey().Public()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return publicKey, nil
	},
		jwt.WithIssuer(jwtConfig.GetIssuer()),
		jwt.WithAudience(jwtConfig.GetAudience()),
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}),
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
