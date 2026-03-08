package me

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/matiasmartin-labs/auth-provider-ms/pkg"
)

// MeResponse representa la respuesta del endpoint /auth/me
type MeResponse struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// MeHandler retorna la información del usuario autenticado
func MeHandler(ctx *gin.Context) {
	claimsValue, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "no authentication claims found",
		})
		return
	}

	claims, ok := claimsValue.(*pkg.Claims)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid claims format",
		})
		return
	}

	ctx.JSON(http.StatusOK, MeResponse{
		Email:   claims.Email,
		Name:    claims.Name,
		Picture: claims.Picture,
	})
}
