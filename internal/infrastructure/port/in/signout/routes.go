package signout

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewSignOutHandler returns a handler that clears the auth cookie.
func NewSignOutHandler(cookieSecure bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.SetCookie("token", "", 0, "/", "", cookieSecure, true)
		ctx.Status(http.StatusNoContent)
	}
}
