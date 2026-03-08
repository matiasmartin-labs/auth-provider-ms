package server

import (
	"github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/in/google"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/in/jwks"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/in/me"
	googleClient "github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/out/google"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/out/token"
	"github.com/matiasmartin-labs/auth-provider-ms/pkg"
)

func Routes(app *pkg.Application) error {
	tokenGenerator := token.NewJwtGenerator()
	googleProvider := googleClient.NewGoogleProviderAdapter()
	googleHandler := google.NewGoogleOAuth2Handler(googleProvider, tokenGenerator)

	app.RegisterGET("/.well-known/jwks.json", jwks.JwksHandler)
	app.RegisterGET("/login/oauth2/code/google", googleHandler.GoogleCallbackHandler)
	app.RegisterGET("/oauth2/authorization/google", googleHandler.GoogleLoginHandler)

	app.RegisterProtectedGET("/api/v1/auth/me", me.MeHandler)

	return nil
}
