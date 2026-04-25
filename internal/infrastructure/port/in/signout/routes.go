package signout

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/matiasmartin-labs/auth-provider-ms/pkg"
)

func SignOutHandler(ctx *gin.Context) {
	ctx.SetCookie("token", "", 0, "/", "", resolveCookieSecure(), true)
	ctx.Status(http.StatusNoContent)
}

func resolveCookieSecure() bool {
	if pkg.App == nil || pkg.App.Config == nil {
		return false
	}

	securityConfig := pkg.App.Config.GetSecurityConfig()
	if securityConfig == nil {
		return false
	}

	cookieConfig := securityConfig.GetCookieConfig()
	if cookieConfig == nil {
		return false
	}

	return cookieConfig.GetSecure()
}
