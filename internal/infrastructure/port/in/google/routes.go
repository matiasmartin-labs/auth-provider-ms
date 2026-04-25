package google

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/application/ports"
	"github.com/matiasmartin-labs/auth-provider-ms/pkg"
)

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
}

type GoogleOAuth2Handler interface {
	GoogleLoginHandler(ctx *gin.Context)
	GoogleCallbackHandler(ctx *gin.Context)
}

type googleOAuth2HandlerImpl struct {
	providerRepository ports.ProviderRepository
	tokenGenerator     ports.TokenGenerator
}

func parseSameSite(raw string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "strict":
		return http.SameSiteStrictMode
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSite(0)
	}
}

func NewGoogleOAuth2Handler(providerRepository ports.ProviderRepository, tokenGenerator ports.TokenGenerator) GoogleOAuth2Handler {
	return &googleOAuth2HandlerImpl{
		providerRepository: providerRepository,
		tokenGenerator:     tokenGenerator,
	}
}

func (h *googleOAuth2HandlerImpl) GoogleCallbackHandler(ctx *gin.Context) {
	securityConfig := pkg.App.Config.GetSecurityConfig().GetOAuth2Config().GetGoogleConfig()
	state := ctx.Query("state")
	if state != securityConfig.GetState() {
		pkg.WriteAuthError(ctx, http.StatusBadRequest, pkg.AuthCodeCallbackStateInvalid, "invalid state parameter")
		return
	}

	code := ctx.Query("code")
	if code == "" {
		pkg.WriteAuthError(ctx, http.StatusBadRequest, pkg.AuthCodeCallbackCodeMissing, "code parameter is missing")
		return
	}

	userInfo, err := h.providerRepository.GetUserInfo(ctx, code)
	if err != nil {
		pkg.WriteAuthError(ctx, http.StatusInternalServerError, pkg.AuthCodeProviderFailure, "authentication provider unavailable")
		return
	}

	if !userInfo.IsEmailAllowed() {
		pkg.WriteAuthError(ctx, http.StatusUnauthorized, pkg.AuthCodeEmailNotAllowed, "email is not allowed")
		return
	}

	token, err := h.tokenGenerator.GenerateToken(userInfo)
	if err != nil {
		pkg.WriteAuthError(ctx, http.StatusInternalServerError, pkg.AuthCodeTokenGenerationFailed, "failed to generate authentication token")
		return
	}

	securityCfg := pkg.App.Config.GetSecurityConfig()
	cookieCfg := securityCfg.GetCookieConfig()
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(cookieCfg.GetMaxAge().Seconds()),
		Secure:   cookieCfg.GetSecure(),
		HttpOnly: cookieCfg.GetHTTPOnly(),
		SameSite: parseSameSite(cookieCfg.GetSameSite()),
	})
	redirectCfg := securityCfg.GetRedirectConfig()
	if redirectCfg.GetEnabled() {
		ctx.Redirect(http.StatusTemporaryRedirect, redirectCfg.GetURL())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (h *googleOAuth2HandlerImpl) GoogleLoginHandler(ctx *gin.Context) {
	securityConfig := pkg.App.Config.GetSecurityConfig().GetOAuth2Config().GetGoogleConfig()
	url := pkg.GoogleOAuth2Config.AuthCodeURL(securityConfig.GetState())
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}
