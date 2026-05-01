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
	fwkerrors "github.com/matiasmartin-labs/common-fwk/errors"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
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

// cookieOptions groups cookie settings for test helpers.
type cookieOptions struct {
	secure   bool
	httpOnly bool
	sameSite string
	maxAge   int
}

var defaultCookieOptions = cookieOptions{
	secure:   false,
	httpOnly: true,
	sameSite: "Strict",
	maxAge:   0,
}

// buildConfig builds a GoogleOAuth2Config for tests.
func buildConfig(state string, redirectEnabled bool, redirectURL string, cookie cookieOptions) GoogleOAuth2Config {
	return GoogleOAuth2Config{
		OAuth2Config: &oauth2.Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-secret",
			RedirectURL:  "http://localhost:8080/callback",
			Scopes:       []string{"email", "profile"},
			Endpoint:     googleoauth.Endpoint,
		},
		State:           state,
		CookieName:      "token",
		CookieMaxAge:    cookie.maxAge,
		CookieSecure:    cookie.secure,
		CookieHTTPOnly:  cookie.httpOnly,
		CookieSameSite:  cookie.sameSite,
		RedirectEnabled: redirectEnabled,
		RedirectURL:     redirectURL,
	}
}

func buildDefaultConfig(state string, redirectEnabled bool, redirectURL string) GoogleOAuth2Config {
	return buildConfig(state, redirectEnabled, redirectURL, defaultCookieOptions)
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
	cfg := buildDefaultConfig("valid-state", false, "")
	handler := NewGoogleOAuth2Handler(&mockProviderRepository{}, &mockTokenGenerator{}, cfg)

	router := gin.New()
	router.GET("/callback", handler.GoogleCallbackHandler)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=invalid-state&code=test-code", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	payload := decodeAuthErrorPayload(t, w.Body.Bytes())
	assert.Equal(t, fwkerrors.CodeCallbackStateInvalid, payload["code"])
	assert.Equal(t, "invalid state parameter", payload["message"])
	assert.Len(t, payload, 2)
	_, hasLegacyError := payload["error"]
	assert.False(t, hasLegacyError)
}

func TestGoogleCallbackHandler_MissingCode(t *testing.T) {
	cfg := buildDefaultConfig("valid-state", false, "")
	handler := NewGoogleOAuth2Handler(&mockProviderRepository{}, &mockTokenGenerator{}, cfg)

	router := gin.New()
	router.GET("/callback", handler.GoogleCallbackHandler)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=valid-state", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	payload := decodeAuthErrorPayload(t, w.Body.Bytes())
	assert.Equal(t, fwkerrors.CodeCallbackCodeMissing, payload["code"])
	assert.Equal(t, "code parameter is missing", payload["message"])
	assert.Len(t, payload, 2)
	_, hasLegacyError := payload["error"]
	assert.False(t, hasLegacyError)
}

func TestGoogleCallbackHandler_ProviderError(t *testing.T) {
	cfg := buildDefaultConfig("valid-state", false, "")
	mockProvider := &mockProviderRepository{
		userInfo: nil,
		err:      errors.New("provider error"),
	}
	handler := NewGoogleOAuth2Handler(mockProvider, &mockTokenGenerator{}, cfg)

	router := gin.New()
	router.GET("/callback", handler.GoogleCallbackHandler)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=valid-state&code=test-code", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	payload := decodeAuthErrorPayload(t, w.Body.Bytes())
	assert.Equal(t, fwkerrors.CodeProviderFailure, payload["code"])
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
			expectedCode:    fwkerrors.CodeEmailNotAllowed,
			expectedMessage: "email is not allowed",
		},
		{
			name:            "with redirect configured",
			redirectEnabled: true,
			redirectURL:     "http://localhost:3000/dashboard",
			expectedCode:    fwkerrors.CodeEmailNotAllowed,
			expectedMessage: "email is not allowed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := buildDefaultConfig("valid-state", tc.redirectEnabled, tc.redirectURL)
			mockProvider := &mockProviderRepository{
				userInfo: &model.UserInfo{
					Email:         "notallowed@example.com",
					FirstName:     "Test",
					LastName:      "User",
					AllowedEmails: []string{"allowed@example.com"},
				},
				err: nil,
			}
			handler := NewGoogleOAuth2Handler(mockProvider, &mockTokenGenerator{}, cfg)

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
	cfg := buildDefaultConfig("valid-state", false, "")
	mockProvider := &mockProviderRepository{
		userInfo: &model.UserInfo{
			Email:         "test@example.com",
			FirstName:     "Test",
			LastName:      "User",
			AllowedEmails: []string{"test@example.com"},
		},
		err: nil,
	}
	mockToken := &mockTokenGenerator{
		token: "",
		err:   errors.New("token generation error"),
	}
	handler := NewGoogleOAuth2Handler(mockProvider, mockToken, cfg)

	router := gin.New()
	router.GET("/callback", handler.GoogleCallbackHandler)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=valid-state&code=test-code", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	payload := decodeAuthErrorPayload(t, w.Body.Bytes())
	assert.Equal(t, fwkerrors.CodeTokenGenerationFailed, payload["code"])
	assert.Equal(t, "failed to generate authentication token", payload["message"])
	assert.Len(t, payload, 2)
	_, hasLegacyError := payload["error"]
	assert.False(t, hasLegacyError)
}

func TestGoogleCallbackHandler_Success_NoRedirect(t *testing.T) {
	cfg := buildDefaultConfig("valid-state", false, "")
	mockProvider := &mockProviderRepository{
		userInfo: &model.UserInfo{
			Email:         "test@example.com",
			FirstName:     "Test",
			LastName:      "User",
			AllowedEmails: []string{"test@example.com"},
		},
		err: nil,
	}
	mockToken := &mockTokenGenerator{
		token: "generated-jwt-token",
		err:   nil,
	}
	handler := NewGoogleOAuth2Handler(mockProvider, mockToken, cfg)

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
	cfg := buildDefaultConfig("valid-state", true, "http://localhost:3000/dashboard")
	mockProvider := &mockProviderRepository{
		userInfo: &model.UserInfo{
			Email:         "test@example.com",
			FirstName:     "Test",
			LastName:      "User",
			AllowedEmails: []string{"test@example.com"},
		},
		err: nil,
	}
	mockToken := &mockTokenGenerator{
		token: "generated-jwt-token",
		err:   nil,
	}
	handler := NewGoogleOAuth2Handler(mockProvider, mockToken, cfg)

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
			cfg := buildConfig("valid-state", false, "", cookieOptions{
				secure:   true,
				httpOnly: true,
				sameSite: tc.sameSite,
				maxAge:   int(time.Hour.Seconds()),
			})
			mockProvider := &mockProviderRepository{
				userInfo: &model.UserInfo{
					Email:         "test@example.com",
					FirstName:     "Test",
					LastName:      "User",
					AllowedEmails: []string{"test@example.com"},
				},
				err: nil,
			}
			mockToken := &mockTokenGenerator{token: "generated-jwt-token", err: nil}
			handler := NewGoogleOAuth2Handler(mockProvider, mockToken, cfg)

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
	cfg := buildDefaultConfig("test-state", false, "")
	handler := NewGoogleOAuth2Handler(&mockProviderRepository{}, &mockTokenGenerator{}, cfg)

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
	cfg := buildDefaultConfig("", false, "")

	handler := NewGoogleOAuth2Handler(mockProvider, mockToken, cfg)

	assert.NotNil(t, handler)
	_, ok := handler.(GoogleOAuth2Handler)
	assert.True(t, ok)
}
