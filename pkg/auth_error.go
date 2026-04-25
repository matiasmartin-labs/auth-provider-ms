package pkg

import "github.com/gin-gonic/gin"

const (
	AuthCodeTokenMissing          = "auth_token_missing"
	AuthCodeTokenInvalid          = "auth_token_invalid"
	AuthCodeCallbackStateInvalid  = "auth_callback_state_invalid"
	AuthCodeCallbackCodeMissing   = "auth_callback_code_missing"
	AuthCodeEmailNotAllowed       = "auth_email_not_allowed"
	AuthCodeProviderFailure       = "auth_provider_failure"
	AuthCodeTokenGenerationFailed = "auth_token_generation_failed"
	AuthCodeClaimsMissing         = "auth_claims_missing"
	AuthCodeClaimsInvalid         = "auth_claims_invalid"
)

type AuthErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteAuthError(ctx *gin.Context, status int, code, message string) {
	ctx.JSON(status, AuthErrorResponse{
		Code:    code,
		Message: message,
	})
}
