package google

import "github.com/gin-gonic/gin"

func GoogleCallbackHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "Google OAuth callback received",
	})
}
