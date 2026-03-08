package pkg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOAuth2ClientConfig_GetClientID(t *testing.T) {
	config := &oAuth2ClientConfig{ClientID: "test-client-id"}
	assert.Equal(t, "test-client-id", config.GetClientID())
}

func TestOAuth2ClientConfig_GetClientSecret(t *testing.T) {
	config := &oAuth2ClientConfig{ClientSecret: "test-secret"}
	assert.Equal(t, "test-secret", config.GetClientSecret())
}

func TestOAuth2ClientConfig_GetRedirectURI(t *testing.T) {
	config := &oAuth2ClientConfig{RedirectURI: "http://localhost:8080/callback"}
	assert.Equal(t, "http://localhost:8080/callback", config.GetRedirectURI())
}

func TestOAuth2ClientConfig_GetScopes(t *testing.T) {
	scopes := []string{"email", "profile"}
	config := &oAuth2ClientConfig{Scopes: scopes}
	assert.Equal(t, scopes, config.GetScopes())
}

func TestOAuth2ClientConfig_GetState(t *testing.T) {
	config := &oAuth2ClientConfig{State: "random-state"}
	assert.Equal(t, "random-state", config.GetState())
}

func TestOAuth2ClientConfig_GetUserInfoURI(t *testing.T) {
	config := &oAuth2ClientConfig{UserInfoURI: "https://api.example.com/userinfo"}
	assert.Equal(t, "https://api.example.com/userinfo", config.GetUserInfoURI())
}

func TestRedirectConfig_GetEnabled(t *testing.T) {
	config := &redirectConfig{Enabled: true}
	assert.True(t, config.GetEnabled())

	config.Enabled = false
	assert.False(t, config.GetEnabled())
}

func TestRedirectConfig_GetURL(t *testing.T) {
	config := &redirectConfig{URL: "http://localhost:3000"}
	assert.Equal(t, "http://localhost:3000", config.GetURL())
}

func TestCookieConfig_GetSecure(t *testing.T) {
	config := &cookieConfig{Secure: true}
	assert.True(t, config.GetSecure())
}

func TestCookieConfig_GetHTTPOnly(t *testing.T) {
	config := &cookieConfig{HTTPOnly: true}
	assert.True(t, config.GetHTTPOnly())
}

func TestCookieConfig_GetSameSite(t *testing.T) {
	config := &cookieConfig{SameSite: "Strict"}
	assert.Equal(t, "Strict", config.GetSameSite())
}

func TestCookieConfig_GetMaxAge(t *testing.T) {
	config := &cookieConfig{MaxAge: time.Hour}
	assert.Equal(t, time.Hour, config.GetMaxAge())
}

func TestLoginConfig_GetAllowedEmails(t *testing.T) {
	emails := []string{"user1@example.com", "user2@example.com"}
	config := &loginConfig{AllowedEmails: emails}
	assert.Equal(t, emails, config.GetAllowedEmails())
}

func TestJWTConfig_GetIssuer(t *testing.T) {
	config := &jwtConfig{Issuer: "test-issuer"}
	assert.Equal(t, "test-issuer", config.GetIssuer())
}

func TestJWTConfig_GetAudience(t *testing.T) {
	config := &jwtConfig{Audience: "test-audience"}
	assert.Equal(t, "test-audience", config.GetAudience())
}

func TestJWTConfig_GetExpirationTime(t *testing.T) {
	config := &jwtConfig{ExpirationTime: 2 * time.Hour}
	assert.Equal(t, 2*time.Hour, config.GetExpirationTime())
}

func TestAuthConfig_IsEnabled(t *testing.T) {
	config := &authConfig{Enabled: true}
	assert.True(t, config.IsEnabled())

	config.Enabled = false
	assert.False(t, config.IsEnabled())
}

func TestOAuth2Config_GetGoogleConfig(t *testing.T) {
	googleCfg := &oAuth2ClientConfig{ClientID: "google-client-id"}
	config := &oAuth2Config{
		Client: map[string]*oAuth2ClientConfig{
			"google": googleCfg,
		},
	}

	result := config.GetGoogleConfig()
	assert.NotNil(t, result)
	assert.Equal(t, "google-client-id", result.GetClientID())
}

func TestOAuth2Config_GetGoogleConfig_NotExists(t *testing.T) {
	config := &oAuth2Config{
		Client: map[string]*oAuth2ClientConfig{},
	}

	result := config.GetGoogleConfig()
	assert.Nil(t, result)
}

func TestSecurityConfig_GetOAuth2Config(t *testing.T) {
	oauth2Cfg := &oAuth2Config{}
	config := &securityConfig{OAuth2: oauth2Cfg}

	result := config.GetOAuth2Config()
	assert.Equal(t, oauth2Cfg, result)
}

func TestSecurityConfig_GetRedirectConfig(t *testing.T) {
	redirectCfg := &redirectConfig{Enabled: true}
	config := &securityConfig{Redirect: redirectCfg}

	result := config.GetRedirectConfig()
	assert.Equal(t, redirectCfg, result)
}

func TestSecurityConfig_GetCookieConfig(t *testing.T) {
	cookieCfg := &cookieConfig{Secure: true}
	config := &securityConfig{Cookie: cookieCfg}

	result := config.GetCookieConfig()
	assert.Equal(t, cookieCfg, result)
}

func TestSecurityConfig_GetLoginConfig(t *testing.T) {
	loginCfg := &loginConfig{AllowedEmails: []string{"test@example.com"}}
	config := &securityConfig{Login: loginCfg}

	result := config.GetLoginConfig()
	assert.Equal(t, loginCfg, result)
}

func TestSecurityConfig_GetJWTConfig(t *testing.T) {
	jwtCfg := &jwtConfig{Issuer: "test"}
	config := &securityConfig{JWT: jwtCfg}

	result := config.GetJWTConfig()
	assert.Equal(t, jwtCfg, result)
}

func TestSecurityConfig_GetAuthConfig(t *testing.T) {
	authCfg := &authConfig{Enabled: true}
	config := &securityConfig{Auth: authCfg}

	result := config.GetAuthConfig()
	assert.Equal(t, authCfg, result)
}

func TestConfiguration_GetServerConfig(t *testing.T) {
	serverCfg := &serverConfig{Port: 8080}
	config := &configuration{Server: serverCfg}

	result := config.GetServerConfig()
	assert.Equal(t, serverCfg, result)
}

func TestConfiguration_GetSecurityConfig(t *testing.T) {
	securityCfg := &securityConfig{}
	config := &configuration{Security: securityCfg}

	result := config.GetSecurityConfig()
	assert.Equal(t, securityCfg, result)
}
