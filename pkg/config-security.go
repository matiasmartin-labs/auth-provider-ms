package pkg

import "time"

type OAuth2ClientConfig interface {
	GetClientID() string
	GetClientSecret() string
	GetRedirectURI() string
	GetScopes() []string
	GetState() string
	GetUserInfoURI() string
}

type oAuth2ClientConfig struct {
	ClientID     string   `mapstructure:"client-id"`
	ClientSecret string   `mapstructure:"client-secret"`
	RedirectURI  string   `mapstructure:"redirect-uri"`
	State        string   `mapstructure:"state"`
	Scopes       []string `mapstructure:"scopes"`
	UserInfoURI  string   `mapstructure:"user-info-uri"`
}

func (o *oAuth2ClientConfig) GetClientID() string {
	return o.ClientID
}

func (o *oAuth2ClientConfig) GetClientSecret() string {
	return o.ClientSecret
}

func (o *oAuth2ClientConfig) GetRedirectURI() string {
	return o.RedirectURI
}

func (o *oAuth2ClientConfig) GetScopes() []string {
	return o.Scopes
}

func (o *oAuth2ClientConfig) GetState() string {
	return o.State
}

func (o *oAuth2ClientConfig) GetUserInfoURI() string {
	return o.UserInfoURI
}

type RedirectConfig interface {
	GetEnabled() bool
	GetURL() string
}

type redirectConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	URL     string `mapstructure:"url"`
}

func (r *redirectConfig) GetEnabled() bool {
	return r.Enabled
}

func (r *redirectConfig) GetURL() string {
	return r.URL
}

type CookieConfig interface {
	GetSecure() bool
	GetHTTPOnly() bool
	GetSameSite() string
	GetMaxAge() time.Duration
}

type cookieConfig struct {
	Secure   bool          `mapstructure:"secure"`
	HTTPOnly bool          `mapstructure:"http-only"`
	SameSite string        `mapstructure:"same-site"`
	MaxAge   time.Duration `mapstructure:"max-age"`
}

func (c *cookieConfig) GetSecure() bool {
	return c.Secure
}

func (c *cookieConfig) GetMaxAge() time.Duration {
	return c.MaxAge
}

func (c *cookieConfig) GetHTTPOnly() bool {
	return c.HTTPOnly
}

func (c *cookieConfig) GetSameSite() string {
	return c.SameSite
}

type LoginConfig interface {
	GetAllowedEmails() []string
}

type loginConfig struct {
	AllowedEmails []string `mapstructure:"allowed-emails"`
}

func (l *loginConfig) GetAllowedEmails() []string {
	return l.AllowedEmails
}

type JWTConfig interface {
	GetIssuer() string
	GetAudience() string
	GetExpirationTime() time.Duration
}

type jwtConfig struct {
	Issuer         string        `mapstructure:"issuer"`
	Audience       string        `mapstructure:"aud"`
	ExpirationTime time.Duration `mapstructure:"expiration-time"`
}

func (j *jwtConfig) GetIssuer() string {
	return j.Issuer
}

func (j *jwtConfig) GetAudience() string {
	return j.Audience
}

func (j *jwtConfig) GetExpirationTime() time.Duration {
	return j.ExpirationTime
}

type OAuth2Config interface {
	GetGoogleConfig() OAuth2ClientConfig
}

type oAuth2Config struct {
	Client map[string]*oAuth2ClientConfig `mapstructure:"client"`
}

func (o *oAuth2Config) GetGoogleConfig() OAuth2ClientConfig {
	config, exists := o.Client["google"]
	if !exists {
		return nil
	}
	return config
}

type SecurityConfig interface {
	GetOAuth2Config() OAuth2Config
	GetRedirectConfig() RedirectConfig
	GetCookieConfig() CookieConfig
	GetLoginConfig() LoginConfig
	GetJWTConfig() JWTConfig
}

type securityConfig struct {
	OAuth2   *oAuth2Config   `mapstructure:"oauth2"`
	Redirect *redirectConfig `mapstructure:"redirect"`
	Cookie   *cookieConfig   `mapstructure:"cookie"`
	Login    *loginConfig    `mapstructure:"login"`
	JWT      *jwtConfig      `mapstructure:"jwt"`
}

func (s *securityConfig) GetOAuth2Config() OAuth2Config {
	return s.OAuth2
}

func (s *securityConfig) GetRedirectConfig() RedirectConfig {
	return s.Redirect
}

func (s *securityConfig) GetCookieConfig() CookieConfig {
	return s.Cookie
}

func (s *securityConfig) GetLoginConfig() LoginConfig {
	return s.Login
}

func (s *securityConfig) GetJWTConfig() JWTConfig {
	return s.JWT
}
