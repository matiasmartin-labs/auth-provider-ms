package server

import (
	"github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/in/google"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/in/jwks"
	"github.com/matiasmartin-labs/auth-provider-ms/pkg"
)

func Routes(app *pkg.Application) error {
	app.Handler.GET("/.well-known/jwks.json", jwks.JwksHandler)
	app.Handler.GET("/auth/google/callback", google.GoogleCallbackHandler)
	return nil
}
