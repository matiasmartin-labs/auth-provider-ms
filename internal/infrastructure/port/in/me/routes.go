package me

import (
	"net/http"

	"github.com/gin-gonic/gin"
	fwkerrors "github.com/matiasmartin-labs/common-fwk/errors"
	httpgin "github.com/matiasmartin-labs/common-fwk/http/gin"
	fwkclaims "github.com/matiasmartin-labs/common-fwk/security/claims"
)

type MeResponse struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func MeHandler(ctx *gin.Context) {
	cl, ok := httpgin.GetClaims(ctx, "claims")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, httpgin.ErrorResponse{
			Code:    fwkerrors.CodeClaimsMissing,
			Message: "no authentication claims found",
		})
		return
	}

	ctx.JSON(http.StatusOK, toMeResponse(cl))
}

func toMeResponse(cl fwkclaims.Claims) MeResponse {
	return MeResponse{
		Email:   cl.Email,
		Name:    cl.Name,
		Picture: cl.Picture,
	}
}
