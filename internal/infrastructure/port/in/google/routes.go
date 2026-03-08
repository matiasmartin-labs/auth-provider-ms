package google

import (
	"net/http"

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
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid state parameter",
		})
		return
	}

	code := ctx.Query("code")
	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "code parameter is missing",
		})
		return
	}

	userInfo, err := h.providerRepository.GetUserInfo(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get user info",
		})
		return
	}

	if !userInfo.IsEmailAllowed() {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "email is not allowed",
		})
		return
	}

	token, err := h.tokenGenerator.GenerateToken(userInfo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	securityCfg := pkg.App.Config.GetSecurityConfig()
	cookieCfg := securityCfg.GetCookieConfig()
	ctx.SetCookie(
		"token",
		token,
		int(cookieCfg.GetMaxAge().Seconds()),
		"/",
		"",
		cookieCfg.GetSecure(),
		cookieCfg.GetHTTPOnly(),
	)
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
