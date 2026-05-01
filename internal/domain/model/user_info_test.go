package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
			userInfo := &UserInfo{Email: tc.email, AllowedEmails: allowedEmails}
			assert.Equal(t, tc.expected, userInfo.IsEmailAllowed())
		})
	}
}

func TestUserInfo_IsEmailAllowed_NotAllowed(t *testing.T) {
	allowedEmails := []string{
		"user1@example.com",
		"user2@example.com",
	}

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
			userInfo := &UserInfo{Email: tc.email, AllowedEmails: allowedEmails}
			assert.Equal(t, tc.expected, userInfo.IsEmailAllowed())
		})
	}
}

func TestUserInfo_IsEmailAllowed_EmptyAllowedList(t *testing.T) {
	userInfo := &UserInfo{Email: "any@example.com", AllowedEmails: []string{}}
	assert.False(t, userInfo.IsEmailAllowed())
}

func TestUserInfo_IsEmailAllowed_NilAllowedList(t *testing.T) {
	userInfo := &UserInfo{Email: "any@example.com", AllowedEmails: nil}
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

	userInfo := &UserInfo{Email: "target@example.com", AllowedEmails: allowedEmails}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = userInfo.IsEmailAllowed()
	}
}
