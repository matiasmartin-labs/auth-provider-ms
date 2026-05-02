package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// newFakeTokenServer returns an httptest.Server that responds to POST /token
// with a minimal valid OAuth2 token response.
func newFakeTokenServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"fake-token","token_type":"Bearer","expires_in":3600}`))
	}))
}

// newFakeUserInfoServer returns an httptest.Server whose handler is provided by the caller.
func newFakeUserInfoServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

// buildAdapter wires up an adapter pointing at the given token server and userinfo URL.
func buildAdapter(tokenServerURL, userInfoURL string, allowedEmails []string) *GoogleProviderAdapter {
	cfg := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  tokenServerURL + "/auth",
			TokenURL: tokenServerURL + "/token",
		},
	}
	return &GoogleProviderAdapter{
		oauth2Config:  cfg,
		userInfoURI:   userInfoURL,
		allowedEmails: allowedEmails,
	}
}

// ---- ProviderError.Error() tests ----

func TestProviderError_Error(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedSubstr string
	}{
		{
			name:           "forbidden returns status text Forbidden",
			statusCode:     http.StatusForbidden,
			expectedSubstr: "Forbidden",
		},
		{
			name:           "unauthorized returns status text Unauthorized",
			statusCode:     http.StatusUnauthorized,
			expectedSubstr: "Unauthorized",
		},
		{
			name:           "internal server error returns status text",
			statusCode:     http.StatusInternalServerError,
			expectedSubstr: "Internal Server Error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := &ProviderError{StatusCode: tc.statusCode}
			msg := err.Error()
			assert.NotEmpty(t, msg, "Error() must return a non-empty string")
			assert.True(t, strings.Contains(msg, tc.expectedSubstr),
				"Error() %q should contain %q", msg, tc.expectedSubstr)
		})
	}
}

// ---- GetUserInfo tests ----

func TestGetUserInfo_CodeExchangeFails(t *testing.T) {
	// Token server that always returns 400 so Exchange() will fail.
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer tokenSrv.Close()

	adapter := buildAdapter(tokenSrv.URL, "http://unused", nil)
	_, err := adapter.GetUserInfo(context.Background(), "bad-code")
	assert.Error(t, err, "Exchange failure must propagate an error")
}

func TestGetUserInfo_UserInfoNon200(t *testing.T) {
	tokenSrv := newFakeTokenServer(t)
	defer tokenSrv.Close()

	tests := []struct {
		name       string
		statusCode int
	}{
		{name: "forbidden (403)", statusCode: http.StatusForbidden},
		{name: "internal server error (500)", statusCode: http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userInfoSrv := newFakeUserInfoServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			})
			defer userInfoSrv.Close()

			adapter := buildAdapter(tokenSrv.URL, userInfoSrv.URL, nil)
			_, err := adapter.GetUserInfo(context.Background(), "any-code")
			require.Error(t, err)

			var provErr *ProviderError
			require.ErrorAs(t, err, &provErr, "error must be a *ProviderError")
			assert.Equal(t, tc.statusCode, provErr.StatusCode)
		})
	}
}

func TestGetUserInfo_MalformedJSON(t *testing.T) {
	tokenSrv := newFakeTokenServer(t)
	defer tokenSrv.Close()

	userInfoSrv := newFakeUserInfoServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{ this is not valid json `))
	})
	defer userInfoSrv.Close()

	adapter := buildAdapter(tokenSrv.URL, userInfoSrv.URL, nil)
	_, err := adapter.GetUserInfo(context.Background(), "any-code")
	assert.Error(t, err, "malformed JSON must return an error")
}

func TestGetUserInfo_HappyPath(t *testing.T) {
	allowedEmails := []string{"alice@example.com", "bob@example.com"}

	tokenSrv := newFakeTokenServer(t)
	defer tokenSrv.Close()

	googleResp := googleUserInfo{
		ID:            "123",
		Email:         "alice@example.com",
		VerifiedEmail: true,
		Picture:       "https://example.com/pic.jpg",
		FirstName:     "Alice",
		LastName:      "Smith",
	}

	userInfoSrv := newFakeUserInfoServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(googleResp); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	})
	defer userInfoSrv.Close()

	adapter := buildAdapter(tokenSrv.URL, userInfoSrv.URL, allowedEmails)
	info, err := adapter.GetUserInfo(context.Background(), "valid-code")
	require.NoError(t, err)
	require.NotNil(t, info)

	assert.Equal(t, googleResp.Email, info.Email)
	assert.Equal(t, googleResp.FirstName, info.FirstName)
	assert.Equal(t, googleResp.LastName, info.LastName)
	assert.Equal(t, googleResp.Picture, info.Picture)
	assert.Equal(t, allowedEmails, info.AllowedEmails, "AllowedEmails must match what was passed to NewGoogleProviderAdapter")
}
