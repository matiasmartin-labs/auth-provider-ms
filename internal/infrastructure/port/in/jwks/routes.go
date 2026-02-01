package jwks

import (
	"github.com/gin-gonic/gin"
	"github.com/matiasmartin-labs/auth-provider-ms/pkg"
)

func JwksHandler(ctx *gin.Context) {

	publicJwk, err := pkg.KeyPairHolder.PublicJWK()
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": "failed to get public JWK",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"keys": []interface{}{publicJwk},
	})
}
