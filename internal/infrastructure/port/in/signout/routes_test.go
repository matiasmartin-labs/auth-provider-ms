package signout

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignOutHandler_ClearsTokenCookieAndReturnsNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testCases := []struct {
		name           string
		secureEnabled  bool
		expectedSecure bool
	}{
		{
			name:           "secure cookie disabled",
			secureEnabled:  false,
			expectedSecure: false,
		},
		{
			name:           "secure cookie enabled",
			secureEnabled:  true,
			expectedSecure: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/api/v1/auth/sign-out", NewSignOutHandler(tc.secureEnabled))

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-out", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNoContent, w.Code)
			assert.Empty(t, w.Body.String())

			tokenCookie := findTokenCookie(t, w.Result().Cookies())
			assert.Equal(t, "", tokenCookie.Value)
			assert.Equal(t, "/", tokenCookie.Path)
			assert.Equal(t, 0, tokenCookie.MaxAge)
			assert.True(t, tokenCookie.HttpOnly)
			assert.Equal(t, tc.expectedSecure, tokenCookie.Secure)
		})
	}
}

func TestSignOutHandler_IsIdempotentAcrossRepeatedCalls(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/api/v1/auth/sign-out", NewSignOutHandler(true))

	firstRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-out", nil)
	firstResponse := httptest.NewRecorder()
	router.ServeHTTP(firstResponse, firstRequest)

	secondRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-out", nil)
	secondResponse := httptest.NewRecorder()
	router.ServeHTTP(secondResponse, secondRequest)

	assert.Equal(t, http.StatusNoContent, firstResponse.Code)
	assert.Empty(t, firstResponse.Body.String())
	firstTokenCookie := findTokenCookie(t, firstResponse.Result().Cookies())
	assert.Equal(t, "", firstTokenCookie.Value)
	assert.Equal(t, "/", firstTokenCookie.Path)
	assert.Equal(t, 0, firstTokenCookie.MaxAge)
	assert.True(t, firstTokenCookie.HttpOnly)
	assert.True(t, firstTokenCookie.Secure)

	assert.Equal(t, http.StatusNoContent, secondResponse.Code)
	assert.Empty(t, secondResponse.Body.String())
	secondTokenCookie := findTokenCookie(t, secondResponse.Result().Cookies())
	assert.Equal(t, "", secondTokenCookie.Value)
	assert.Equal(t, "/", secondTokenCookie.Path)
	assert.Equal(t, 0, secondTokenCookie.MaxAge)
	assert.True(t, secondTokenCookie.HttpOnly)
	assert.True(t, secondTokenCookie.Secure)
}

func findTokenCookie(t *testing.T, cookies []*http.Cookie) *http.Cookie {
	t.Helper()
	require.NotEmpty(t, cookies)

	for _, cookie := range cookies {
		if cookie.Name == "token" {
			return cookie
		}
	}

	require.Fail(t, "token cookie was not set")
	return nil
}
