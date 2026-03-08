package model

import (
	"testing"
	"time"

	"github.com/matiasmartin-labs/auth-provider-ms/pkg"
	"github.com/stretchr/testify/assert"
)

type mockLoginConfig struct {
	allowedEmails []string
}

func (m *mockLoginConfig) GetAllowedEmails() []string { return m.allowedEmails }

type mockSecurityConfig struct {
	loginConfig *mockLoginConfig
}

func (m *mockSecurityConfig) GetOAuth2Config() pkg.OAuth2Config     { return nil }
func (m *mockSecurityConfig) GetRedirectConfig() pkg.RedirectConfig { return nil }
func (m *mockSecurityConfig) GetCookieConfig() pkg.CookieConfig     { return nil }
func (m *mockSecurityConfig) GetLoginConfig() pkg.LoginConfig       { return m.loginConfig }
func (m *mockSecurityConfig) GetJWTConfig() pkg.JWTConfig           { return nil }
func (m *mockSecurityConfig) GetAuthConfig() pkg.AuthConfig         { return nil }

type mockConfiguration struct {
	securityConfig *mockSecurityConfig
}

func (m *mockConfiguration) GetServerConfig() pkg.ServerConfig     { return nil }
func (m *mockConfiguration) GetSecurityConfig() pkg.SecurityConfig { return m.securityConfig }

type mockKeyPair struct{}

func (m *mockKeyPair) PublicJWK() (map[string]interface{}, error) { return nil, nil }
func (m *mockKeyPair) GetPrivateKey() interface{}                 { return nil }

func setupMockApp(allowedEmails []string) {
	pkg.App = &pkg.Application{
		Config: &mockConfiguration{
			securityConfig: &mockSecurityConfig{
				loginConfig: &mockLoginConfig{
					allowedEmails: allowedEmails,
				},
			},
		},
	}
}

func TestUserInfo_GetFullName(t *testing.T) {
	testCases := []struct {
		name         string
		firstName    string
		lastName     string
		expectedName string
	}{
		{
			name:         "Normal names",
			firstName:    "John",
			lastName:     "Doe",
			expectedName: "John Doe",
		},
		{
			name:         "Spanish names",
			firstName:    "María",
			lastName:     "García López",
			expectedName: "María García López",
		},
		{
			name:         "Empty first name",
			firstName:    "",
			lastName:     "Smith",
			expectedName: " Smith",
		},
		{
			name:         "Empty last name",
			firstName:    "Jane",
			lastName:     "",
			expectedName: "Jane ",
		},
		{
			name:         "Both empty",
			firstName:    "",
			lastName:     "",
			expectedName: " ",
		},
		{
			name:         "Single word name",
			firstName:    "Madonna",
			lastName:     "",
			expectedName: "Madonna ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userInfo := &UserInfo{
				FirstName: tc.firstName,
				LastName:  tc.lastName,
			}

			result := userInfo.GetFullName()
			assert.Equal(t, tc.expectedName, result)
		})
	}
}

func TestUserInfo_IsEmailAllowed_Allowed(t *testing.T) {
	allowedEmails := []string{
		"user1@example.com",
		"user2@example.com",
		"admin@company.org",
	}
	setupMockApp(allowedEmails)

	testCases := []struct {
		email    string
		expected bool
	}{
		{"user1@example.com", true},
		{"user2@example.com", true},
		{"admin@company.org", true},
	}

	for _, tc := range testCases {
		t.Run(tc.email, func(t *testing.T) {
			userInfo := &UserInfo{Email: tc.email}
			assert.Equal(t, tc.expected, userInfo.IsEmailAllowed())
		})
	}
}

func TestUserInfo_IsEmailAllowed_NotAllowed(t *testing.T) {
	allowedEmails := []string{
		"user1@example.com",
		"user2@example.com",
	}
	setupMockApp(allowedEmails)

	testCases := []struct {
		email    string
		expected bool
	}{
		{"notallowed@example.com", false},
		{"hacker@evil.com", false},
		{"", false},
		{"USER1@EXAMPLE.COM", false}, // Case sensitive
	}

	for _, tc := range testCases {
		t.Run(tc.email, func(t *testing.T) {
			userInfo := &UserInfo{Email: tc.email}
			assert.Equal(t, tc.expected, userInfo.IsEmailAllowed())
		})
	}
}

func TestUserInfo_IsEmailAllowed_EmptyAllowedList(t *testing.T) {
	setupMockApp([]string{})

	userInfo := &UserInfo{Email: "any@example.com"}
	assert.False(t, userInfo.IsEmailAllowed())
}

func TestUserInfo_IsEmailAllowed_NilAllowedList(t *testing.T) {
	setupMockApp(nil)

	userInfo := &UserInfo{Email: "any@example.com"}
	assert.False(t, userInfo.IsEmailAllowed())
}

func TestUserInfo_Struct(t *testing.T) {
	userInfo := &UserInfo{
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Picture:   "https://example.com/avatar.jpg",
	}

	assert.Equal(t, "test@example.com", userInfo.Email)
	assert.Equal(t, "John", userInfo.FirstName)
	assert.Equal(t, "Doe", userInfo.LastName)
	assert.Equal(t, "https://example.com/avatar.jpg", userInfo.Picture)
}

func BenchmarkUserInfo_GetFullName(b *testing.B) {
	userInfo := &UserInfo{
		FirstName: "John",
		LastName:  "Doe",
	}

	for i := 0; i < b.N; i++ {
		_ = userInfo.GetFullName()
	}
}

func BenchmarkUserInfo_IsEmailAllowed(b *testing.B) {
	allowedEmails := make([]string, 100)
	for i := 0; i < 100; i++ {
		allowedEmails[i] = "user" + time.Now().String() + "@example.com"
	}
	allowedEmails[50] = "target@example.com"
	setupMockApp(allowedEmails)

	userInfo := &UserInfo{Email: "target@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = userInfo.IsEmailAllowed()
	}
}
