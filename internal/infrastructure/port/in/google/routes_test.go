package google

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/domain/model"
	"github.com/matiasmartin-labs/auth-provider-ms/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func decodeAuthErrorPayload(t *testing.T, body []byte) map[string]string {
	t.Helper()

	var payload map[string]string
	err := json.Unmarshal(body, &payload)
	require.NoError(t, err)

	return payload
}

type mockProviderRepository struct {
	userInfo *model.UserInfo
	err      error
}

func (m *mockProviderRepository) GetUserInfo(ctx context.Context, code string) (*model.UserInfo, error) {
	return m.userInfo, m.err
}

type mockTokenGenerator struct {
	token string
	err   error
}

func (m *mockTokenGenerator) GenerateToken(userInfo *model.UserInfo) (string, error) {
	return m.token, m.err
}

type mockOAuth2ClientConfig struct {
	clientID     string
	clientSecret string
	redirectURI  string
	state        string
	scopes       []string
	userInfoURI  string
}

func (m *mockOAuth2ClientConfig) GetClientID() string     { return m.clientID }
func (m *mockOAuth2ClientConfig) GetClientSecret() string { return m.clientSecret }
func (m *mockOAuth2ClientConfig) GetRedirectURI() string  { return m.redirectURI }
func (m *mockOAuth2ClientConfig) GetState() string        { return m.state }
func (m *mockOAuth2ClientConfig) GetScopes() []string     { return m.scopes }
func (m *mockOAuth2ClientConfig) GetUserInfoURI() string  { return m.userInfoURI }

type mockOAuth2Config struct {
	googleConfig *mockOAuth2ClientConfig
}

func (m *mockOAuth2Config) GetGoogleConfig() pkg.OAuth2ClientConfig { return m.googleConfig }

type mockRedirectConfig struct {
	enabled bool
	url     string
}

func (m *mockRedirectConfig) GetEnabled() bool { return m.enabled }
func (m *mockRedirectConfig) GetURL() string   { return m.url }

type mockCookieConfig struct {
	secure   bool
	httpOnly bool
	sameSite string
	maxAge   time.Duration
}

func (m *mockCookieConfig) GetSecure() bool          { return m.secure }
func (m *mockCookieConfig) GetHTTPOnly() bool        { return m.httpOnly }
func (m *mockCookieConfig) GetSameSite() string      { return m.sameSite }
func (m *mockCookieConfig) GetMaxAge() time.Duration { return m.maxAge }

type mockLoginConfig struct {
	allowedEmails []string
}

func (m *mockLoginConfig) GetAllowedEmails() []string { return m.allowedEmails }

type mockJWTConfig struct {
	issuer         string
	audience       string
	expirationTime time.Duration
}

func (m *mockJWTConfig) GetIssuer() string                { return m.issuer }
func (m *mockJWTConfig) GetAudience() string              { return m.audience }
func (m *mockJWTConfig) GetExpirationTime() time.Duration { return m.expirationTime }

type mockAuthConfig struct {
	enabled bool
}

func (m *mockAuthConfig) IsEnabled() bool { return m.enabled }

type mockSecurityConfig struct {
	oauth2Config   *mockOAuth2Config
	redirectConfig *mockRedirectConfig
	cookieConfig   *mockCookieConfig
	loginConfig    *mockLoginConfig
	jwtConfig      *mockJWTConfig
	authConfig     *mockAuthConfig
}

func (m *mockSecurityConfig) GetOAuth2Config() pkg.OAuth2Config     { return m.oauth2Config }
func (m *mockSecurityConfig) GetRedirectConfig() pkg.RedirectConfig { return m.redirectConfig }
func (m *mockSecurityConfig) GetCookieConfig() pkg.CookieConfig     { return m.cookieConfig }
func (m *mockSecurityConfig) GetLoginConfig() pkg.LoginConfig       { return m.loginConfig }
func (m *mockSecurityConfig) GetJWTConfig() pkg.JWTConfig           { return m.jwtConfig }
func (m *mockSecurityConfig) GetAuthConfig() pkg.AuthConfig         { return m.authConfig }

type mockConfiguration struct {
	securityConfig *mockSecurityConfig
}

func (m *mockConfiguration) GetServerConfig() pkg.ServerConfig     { return nil }
func (m *mockConfiguration) GetSecurityConfig() pkg.SecurityConfig { return m.securityConfig }

func setupMockAppWithCookie(state string, allowedEmails []string, redirectEnabled bool, redirectURL string, cookieCfg *mockCookieConfig) {
	if cookieCfg == nil {
		cookieCfg = &mockCookieConfig{
			secure:   false,
			httpOnly: true,
			sameSite: "Strict",
			maxAge:   time.Hour,
		}
	}

	pkg.App = &pkg.Application{
		Config: &mockConfiguration{
			securityConfig: &mockSecurityConfig{
				oauth2Config: &mockOAuth2Config{
					googleConfig: &mockOAuth2ClientConfig{
						state:       state,
						clientID:    "test-client-id",
						redirectURI: "http://localhost:8080/callback",
					},
				},
				redirectConfig: &mockRedirectConfig{
					enabled: redirectEnabled,
					url:     redirectURL,
				},
				cookieConfig: cookieCfg,
				loginConfig: &mockLoginConfig{
					allowedEmails: allowedEmails,
				},
			},
		},
	}

	pkg.GoogleOAuth2Config = &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"email", "profile"},
	}
}

func setupMockApp(state string, allowedEmails []string, redirectEnabled bool, redirectURL string) {
	setupMockAppWithCookie(state, allowedEmails, redirectEnabled, redirectURL, nil)
}

func TestParseSameSite(t *testing.T) {
	testCases := []struct {
		name     string
		raw      string
		expected http.SameSite
	}{
		{name: "strict mixed case with spaces", raw: "  sTRicT  ", expected: http.SameSiteStrictMode},
		{name: "lax mixed case", raw: "LaX", expected: http.SameSiteLaxMode},
		{name: "none mixed case", raw: "nOnE", expected: http.SameSiteNoneMode},
		{name: "empty returns omitted mode", raw: "", expected: http.SameSite(0)},
		{name: "invalid returns omitted mode", raw: "unsupported", expected: http.SameSite(0)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, parseSameSite(tc.raw))
		})
	}
}

func TestGoogleCallbackHandler_InvalidState(t *testing.T) {
	setupMockApp("valid-state", []string{}, false, "")

	handler := NewGoogleOAuth2Handler(&mockProviderRepository{}, &mockTokenGenerator{})

	router := gin.New()
	router.GET("/callback", handler.GoogleCallbackHandler)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=invalid-state&code=test-code", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	payload := decodeAuthErrorPayload(t, w.Body.Bytes())
	assert.Equal(t, pkg.AuthCodeCallbackStateInvalid, payload["code"])
	assert.Equal(t, "invalid state parameter", payload["message"])
	assert.Len(t, payload, 2)
	_, hasLegacyError := payload["error"]
	assert.False(t, hasLegacyError)
}

func TestGoogleCallbackHandler_MissingCode(t *testing.T) {
	setupMockApp("valid-state", []string{}, false, "")

	handler := NewGoogleOAuth2Handler(&mockProviderRepository{}, &mockTokenGenerator{})

	router := gin.New()
	router.GET("/callback", handler.GoogleCallbackHandler)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=valid-state", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	payload := decodeAuthErrorPayload(t, w.Body.Bytes())
	assert.Equal(t, pkg.AuthCodeCallbackCodeMissing, payload["code"])
	assert.Equal(t, "code parameter is missing", payload["message"])
	assert.Len(t, payload, 2)
	_, hasLegacyError := payload["error"]
	assert.False(t, hasLegacyError)
}

func TestGoogleCallbackHandler_ProviderError(t *testing.T) {
	setupMockApp("valid-state", []string{}, false, "")

	mockProvider := &mockProviderRepository{
		userInfo: nil,
		err:      errors.New("provider error"),
	}
	handler := NewGoogleOAuth2Handler(mockProvider, &mockTokenGenerator{})

	router := gin.New()
	router.GET("/callback", handler.GoogleCallbackHandler)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=valid-state&code=test-code", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	payload := decodeAuthErrorPayload(t, w.Body.Bytes())
	assert.Equal(t, pkg.AuthCodeProviderFailure, payload["code"])
	assert.Equal(t, "authentication provider unavailable", payload["message"])
	assert.Len(t, payload, 2)
	_, hasLegacyError := payload["error"]
	assert.False(t, hasLegacyError)
}

func TestGoogleCallbackHandler_EmailNotAllowed(t *testing.T) {
	testCases := []struct {
		name            string
		redirectEnabled bool
		redirectURL     string
		expectedCode    string
		expectedMessage string
	}{
		{
			name:            "without redirect",
			redirectEnabled: false,
			redirectURL:     "",
			expectedCode:    pkg.AuthCodeEmailNotAllowed,
			expectedMessage: "email is not allowed",
		},
		{
			name:            "with redirect configured",
			redirectEnabled: true,
			redirectURL:     "http://localhost:3000/dashboard",
			expectedCode:    pkg.AuthCodeEmailNotAllowed,
			expectedMessage: "email is not allowed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupMockApp("valid-state", []string{"allowed@example.com"}, tc.redirectEnabled, tc.redirectURL)

			mockProvider := &mockProviderRepository{
				userInfo: &model.UserInfo{
					Email:     "notallowed@example.com",
					FirstName: "Test",
					LastName:  "User",
				},
				err: nil,
			}
			handler := NewGoogleOAuth2Handler(mockProvider, &mockTokenGenerator{})

			router := gin.New()
			router.GET("/callback", handler.GoogleCallbackHandler)

			req := httptest.NewRequest(http.MethodGet, "/callback?state=valid-state&code=test-code", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			payload := decodeAuthErrorPayload(t, w.Body.Bytes())
			assert.Equal(t, tc.expectedCode, payload["code"])
			assert.Equal(t, tc.expectedMessage, payload["message"])
			assert.Len(t, payload, 2)
			_, hasLegacyError := payload["error"]
			assert.False(t, hasLegacyError)
			assert.Empty(t, w.Result().Cookies())
			assert.Empty(t, w.Header().Get("Location"))
		})
	}
}

func TestGoogleCallbackHandler_TokenGenerationError(t *testing.T) {
	setupMockApp("valid-state", []string{"test@example.com"}, false, "")

	mockProvider := &mockProviderRepository{
		userInfo: &model.UserInfo{
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
		},
		err: nil,
	}
	mockToken := &mockTokenGenerator{
		token: "",
		err:   errors.New("token generation error"),
	}
	handler := NewGoogleOAuth2Handler(mockProvider, mockToken)

	router := gin.New()
	router.GET("/callback", handler.GoogleCallbackHandler)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=valid-state&code=test-code", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	payload := decodeAuthErrorPayload(t, w.Body.Bytes())
	assert.Equal(t, pkg.AuthCodeTokenGenerationFailed, payload["code"])
	assert.Equal(t, "failed to generate authentication token", payload["message"])
	assert.Len(t, payload, 2)
	_, hasLegacyError := payload["error"]
	assert.False(t, hasLegacyError)
}

func TestGoogleCallbackHandler_Success_NoRedirect(t *testing.T) {
	setupMockApp("valid-state", []string{"test@example.com"}, false, "")

	mockProvider := &mockProviderRepository{
		userInfo: &model.UserInfo{
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
		},
		err: nil,
	}
	mockToken := &mockTokenGenerator{
		token: "generated-jwt-token",
		err:   nil,
	}
	handler := NewGoogleOAuth2Handler(mockProvider, mockToken)

	router := gin.New()
	router.GET("/callback", handler.GoogleCallbackHandler)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=valid-state&code=test-code", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "generated-jwt-token")

	cookies := w.Result().Cookies()
	var tokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "token" {
			tokenCookie = c
			break
		}
	}
	assert.NotNil(t, tokenCookie)
	assert.Equal(t, "generated-jwt-token", tokenCookie.Value)
}

func TestGoogleCallbackHandler_Success_WithRedirect(t *testing.T) {
	setupMockApp("valid-state", []string{"test@example.com"}, true, "http://localhost:3000/dashboard")

	mockProvider := &mockProviderRepository{
		userInfo: &model.UserInfo{
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
		},
		err: nil,
	}
	mockToken := &mockTokenGenerator{
		token: "generated-jwt-token",
		err:   nil,
	}
	handler := NewGoogleOAuth2Handler(mockProvider, mockToken)

	router := gin.New()
	router.GET("/callback", handler.GoogleCallbackHandler)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=valid-state&code=test-code", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, "http://localhost:3000/dashboard", w.Header().Get("Location"))
}

func TestGoogleCallbackHandler_Success_SameSiteMapping(t *testing.T) {
	testCases := []struct {
		name               string
		sameSite           string
		expectedSameSite   string
		expectSameSiteAttr bool
	}{
		{name: "strict mixed case", sameSite: "sTRicT", expectedSameSite: "SameSite=Strict", expectSameSiteAttr: true},
		{name: "lax mixed case", sameSite: "laX", expectedSameSite: "SameSite=Lax", expectSameSiteAttr: true},
		{name: "none mixed case", sameSite: "NoNe", expectedSameSite: "SameSite=None", expectSameSiteAttr: true},
		{name: "empty fallback omits samesite", sameSite: "", expectSameSiteAttr: false},
		{name: "invalid fallback omits samesite", sameSite: "unsupported", expectSameSiteAttr: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupMockAppWithCookie("valid-state", []string{"test@example.com"}, false, "", &mockCookieConfig{
				secure:   true,
				httpOnly: true,
				sameSite: tc.sameSite,
				maxAge:   time.Hour,
			})

			mockProvider := &mockProviderRepository{
				userInfo: &model.UserInfo{
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
				},
				err: nil,
			}
			mockToken := &mockTokenGenerator{token: "generated-jwt-token", err: nil}
			handler := NewGoogleOAuth2Handler(mockProvider, mockToken)

			router := gin.New()
			router.GET("/callback", handler.GoogleCallbackHandler)

			req := httptest.NewRequest(http.MethodGet, "/callback?state=valid-state&code=test-code", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			setCookieHeader := w.Header().Get("Set-Cookie")
			assert.Contains(t, setCookieHeader, "token=generated-jwt-token")
			assert.Contains(t, setCookieHeader, "Path=/")
			assert.Contains(t, setCookieHeader, "Max-Age=3600")
			assert.Contains(t, setCookieHeader, "HttpOnly")
			assert.Contains(t, setCookieHeader, "Secure")

			if tc.expectSameSiteAttr {
				assert.Contains(t, setCookieHeader, tc.expectedSameSite)
			} else {
				assert.NotContains(t, setCookieHeader, "SameSite")
			}
		})
	}
}

func TestGoogleLoginHandler_Redirect(t *testing.T) {
	setupMockApp("test-state", []string{}, false, "")

	handler := NewGoogleOAuth2Handler(&mockProviderRepository{}, &mockTokenGenerator{})

	router := gin.New()
	router.GET("/login", handler.GoogleLoginHandler)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)

	location := w.Header().Get("Location")
	assert.Contains(t, location, "state=test-state")
	assert.Contains(t, location, "client_id=test-client-id")
	assert.Contains(t, location, "response_type=code")
}

func TestNewGoogleOAuth2Handler(t *testing.T) {
	mockProvider := &mockProviderRepository{}
	mockToken := &mockTokenGenerator{}

	handler := NewGoogleOAuth2Handler(mockProvider, mockToken)

	assert.NotNil(t, handler)
	_, ok := handler.(GoogleOAuth2Handler)
	assert.True(t, ok)
}
