package google

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	fwkerrors "github.com/matiasmartin-labs/common-fwk/errors"
	httpgin "github.com/matiasmartin-labs/common-fwk/http/gin"

	"github.com/matiasmartin-labs/auth-provider-ms/internal/application/ports"
)

// GoogleOAuth2Config holds the configuration needed by the Google OAuth2 handler.
type GoogleOAuth2Config struct {
	OAuth2Config  *oauth2.Config
	State         string
	CookieName    string
	CookieMaxAge  int
	CookieSecure  bool
	CookieHTTPOnly bool
	CookieSameSite string
	RedirectEnabled bool
	RedirectURL     string
}

type GoogleOAuth2Handler interface {
	GoogleLoginHandler(ctx *gin.Context)
	GoogleCallbackHandler(ctx *gin.Context)
}

type googleOAuth2HandlerImpl struct {
	providerRepository ports.ProviderRepository
	tokenGenerator     ports.TokenGenerator
	cfg                GoogleOAuth2Config
}

func NewGoogleOAuth2Handler(
	providerRepository ports.ProviderRepository,
	tokenGenerator ports.TokenGenerator,
	cfg GoogleOAuth2Config,
) GoogleOAuth2Handler {
	return &googleOAuth2HandlerImpl{
		providerRepository: providerRepository,
		tokenGenerator:     tokenGenerator,
		cfg:                cfg,
	}
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

func (h *googleOAuth2HandlerImpl) GoogleCallbackHandler(ctx *gin.Context) {
	state := ctx.Query("state")
	if state != h.cfg.State {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httpgin.ErrorResponse{
			Code:    fwkerrors.CodeCallbackStateInvalid,
			Message: "invalid state parameter",
		})
		return
	}

	code := ctx.Query("code")
	if code == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httpgin.ErrorResponse{
			Code:    fwkerrors.CodeCallbackCodeMissing,
			Message: "code parameter is missing",
		})
		return
	}

	userInfo, err := h.providerRepository.GetUserInfo(ctx, code)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, httpgin.ErrorResponse{
			Code:    fwkerrors.CodeProviderFailure,
			Message: "authentication provider unavailable",
		})
		return
	}

	if !userInfo.IsEmailAllowed() {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, httpgin.ErrorResponse{
			Code:    fwkerrors.CodeEmailNotAllowed,
			Message: "email is not allowed",
		})
		return
	}

	token, err := h.tokenGenerator.GenerateToken(userInfo)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, httpgin.ErrorResponse{
			Code:    fwkerrors.CodeTokenGenerationFailed,
			Message: "failed to generate authentication token",
		})
		return
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   h.cfg.CookieMaxAge,
		Secure:   h.cfg.CookieSecure,
		HttpOnly: h.cfg.CookieHTTPOnly,
		SameSite: parseSameSite(h.cfg.CookieSameSite),
	})

	if h.cfg.RedirectEnabled {
		ctx.Redirect(http.StatusTemporaryRedirect, h.cfg.RedirectURL)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *googleOAuth2HandlerImpl) GoogleLoginHandler(ctx *gin.Context) {
	url := h.cfg.OAuth2Config.AuthCodeURL(h.cfg.State)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}
