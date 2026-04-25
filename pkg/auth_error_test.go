package pkg

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteAuthError(t *testing.T) {
	testCases := []struct {
		name           string
		status         int
		code           string
		message        string
		expectedStatus int
	}{
		{
			name:           "writes unauthorized token missing payload",
			status:         http.StatusUnauthorized,
			code:           AuthCodeTokenMissing,
			message:        "missing authentication token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "writes internal provider failure payload",
			status:         http.StatusInternalServerError,
			code:           AuthCodeProviderFailure,
			message:        "authentication provider unavailable",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			WriteAuthError(ctx, tc.status, tc.code, tc.message)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var payload map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &payload)
			require.NoError(t, err)

			assert.Equal(t, tc.code, payload["code"])
			assert.Equal(t, tc.message, payload["message"])
			assert.Len(t, payload, 2)
			_, hasLegacyError := payload["error"]
			assert.False(t, hasLegacyError)
		})
	}
}

func TestAuthErrorCodes_AreStableAndSnakeCase(t *testing.T) {
	testCases := []struct {
		name     string
		actual   string
		expected string
	}{
		{name: "token missing", actual: AuthCodeTokenMissing, expected: "auth_token_missing"},
		{name: "token invalid", actual: AuthCodeTokenInvalid, expected: "auth_token_invalid"},
		{name: "callback state invalid", actual: AuthCodeCallbackStateInvalid, expected: "auth_callback_state_invalid"},
		{name: "callback code missing", actual: AuthCodeCallbackCodeMissing, expected: "auth_callback_code_missing"},
		{name: "email not allowed", actual: AuthCodeEmailNotAllowed, expected: "auth_email_not_allowed"},
		{name: "provider failure", actual: AuthCodeProviderFailure, expected: "auth_provider_failure"},
		{name: "token generation failed", actual: AuthCodeTokenGenerationFailed, expected: "auth_token_generation_failed"},
		{name: "claims missing", actual: AuthCodeClaimsMissing, expected: "auth_claims_missing"},
		{name: "claims invalid", actual: AuthCodeClaimsInvalid, expected: "auth_claims_invalid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.actual)
		})
	}
}
