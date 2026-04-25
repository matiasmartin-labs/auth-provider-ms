package me

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/matiasmartin-labs/auth-provider-ms/pkg"
)

type MeResponse struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func MeHandler(ctx *gin.Context) {
	claimsValue, exists := ctx.Get("claims")
	if !exists {
		pkg.WriteAuthError(ctx, http.StatusUnauthorized, pkg.AuthCodeClaimsMissing, "no authentication claims found")
		return
	}

	claims, ok := claimsValue.(*pkg.Claims)
	if !ok {
		pkg.WriteAuthError(ctx, http.StatusInternalServerError, pkg.AuthCodeClaimsInvalid, "invalid authentication claims")
		return
	}

	ctx.JSON(http.StatusOK, MeResponse{
		Email:   claims.Email,
		Name:    claims.Name,
		Picture: claims.Picture,
	})
}
