package server

import (
	"strings"
	"time"

	fwkapp "github.com/matiasmartin-labs/common-fwk/app"
	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"

	googlein "github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/in/google"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/in/jwks"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/in/me"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/in/signout"
	googleout "github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/out/google"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/out/token"
)

const googleUserInfoURI = "https://www.googleapis.com/oauth2/v2/userinfo"

// Routes wires all HTTP routes onto the application.
// All runtime dependencies are resolved directly from app.Application.
func Routes(app *fwkapp.Application) error {
	cfg := app.GetConfig()
	googleProvider := cfg.Security.Auth.OAuth2.Providers["google"]
	cookieCfg := cfg.Security.Auth.Cookie
	jwtCfg := cfg.Security.Auth.JWT

	oauth2Config := &oauth2.Config{
		ClientID:     googleProvider.ClientID,
		ClientSecret: googleProvider.ClientSecret,
		RedirectURL:  googleProvider.RedirectURL,
		Scopes:       googleProvider.Scopes,
		Endpoint:     googleoauth.Endpoint,
	}

	allowedEmails := resolveAllowedEmails(cfg.Security.Auth.Login.Email)

	// v0.9.0: private key retrieved from Application — no manual keypair needed
	tokenGen := token.NewJwtGenerator(token.JwtGeneratorConfig{
		PrivateKey:     app.GetRSAPrivateKey(),
		Issuer:         jwtCfg.Issuer,
		Audience:       "auth-provider-clients",
		ExpirationTime: time.Duration(jwtCfg.TTLMinutes) * time.Minute,
	})

	googleProviderAdapter := googleout.NewGoogleProviderAdapter(oauth2Config, googleUserInfoURI, allowedEmails)

	googleHandler := googlein.NewGoogleOAuth2Handler(
		googleProviderAdapter,
		tokenGen,
		googlein.GoogleOAuth2Config{
			OAuth2Config:    oauth2Config,
			State:           "",
			CookieName:      cookieCfg.Name,
			CookieMaxAge:    0,
			CookieSecure:    cookieCfg.Secure,
			CookieHTTPOnly:  cookieCfg.HTTPOnly,
			CookieSameSite:  cookieCfg.SameSite,
			RedirectEnabled: false,
		},
	)

	// v0.10.0: public key and key ID retrieved from Application — no jwks.KeyPair needed
	if err := app.RegisterGET("/.well-known/jwks.json", jwks.NewJwksHandlerFromPublicKey(app.GetRSAPublicKey(), app.GetRSAKeyID())); err != nil {
		return err
	}
	if err := app.RegisterGET("/login/oauth2/code/google", googleHandler.GoogleCallbackHandler); err != nil {
		return err
	}
	if err := app.RegisterGET("/oauth2/authorization/google", googleHandler.GoogleLoginHandler); err != nil {
		return err
	}
	if err := app.RegisterPOST("/api/v1/auth/sign-out", signout.NewSignOutHandler(cookieCfg.Secure)); err != nil {
		return err
	}
	if err := app.RegisterProtectedGET("/api/v1/auth/me", me.MeHandler); err != nil {
		return err
	}

	return nil
}

// resolveAllowedEmails splits a comma-separated email list or wraps a single value.
func resolveAllowedEmails(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		email := strings.TrimSpace(p)
		if email != "" {
			out = append(out, email)
		}
	}

	if len(out) == 0 {
		return nil
	}

	return out
}
