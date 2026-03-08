package pkg

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var GoogleOAuth2Config *oauth2.Config

func (app *Application) UseGoogleProvider() *Application {
	securityConfig := app.Config.GetSecurityConfig().GetOAuth2Config().GetGoogleConfig()
	GoogleOAuth2Config = &oauth2.Config{
		ClientID:     securityConfig.GetClientID(),
		ClientSecret: securityConfig.GetClientSecret(),
		RedirectURL:  securityConfig.GetRedirectURI(),
		Scopes:       securityConfig.GetScopes(),
		Endpoint:     google.Endpoint,
	}
	return app
}
