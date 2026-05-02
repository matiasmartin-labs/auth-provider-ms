package server

import (
	"crypto/rsa"
	"strings"
	"time"

	fwkapp "github.com/matiasmartin-labs/common-fwk/app"
	fwkconfig "github.com/matiasmartin-labs/common-fwk/config"
	fwklogging "github.com/matiasmartin-labs/common-fwk/logging"
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

// Bootstrap holds the runtime dependencies resolved at startup.
type Bootstrap struct {
	PrivateKey *rsa.PrivateKey
	KeyPair    *jwks.KeyPair
	Config     fwkconfig.Config
	Logger     fwklogging.Logger
}

// Routes wires all HTTP routes onto the application using the provided bootstrap.
func Routes(app *fwkapp.Application, b *Bootstrap) error {
	// v0.5.0: use GetConfig() read-only accessor for runtime config snapshot
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

	tokenGen := token.NewJwtGenerator(token.JwtGeneratorConfig{
		PrivateKey:     b.PrivateKey,
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
			State:           resolveState(googleProvider),
			CookieName:      cookieCfg.Name,
			CookieMaxAge:    0,
			CookieSecure:    cookieCfg.Secure,
			CookieHTTPOnly:  cookieCfg.HTTPOnly,
			CookieSameSite:  cookieCfg.SameSite,
			RedirectEnabled: false,
		},
	)

	if err := app.RegisterGET("/.well-known/jwks.json", jwks.NewJwksHandler(b.KeyPair)); err != nil {
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

// resolveState reads the OAuth2 state from the provider config.
// common-fwk does not model state directly; fall back to empty string.
func resolveState(_ fwkconfig.OAuth2ProviderConfig) string {
	return ""
}
