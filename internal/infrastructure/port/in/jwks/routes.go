package jwks

import "github.com/gin-gonic/gin"

func JwksHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"keys": []string{},
	})
}
