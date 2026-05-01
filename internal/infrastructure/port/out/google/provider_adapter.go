package google

import (
	"context"
	"encoding/json"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/matiasmartin-labs/auth-provider-ms/internal/application/ports"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/domain/model"
)

type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
}

// GoogleProviderAdapter fetches user info from the Google OAuth2 userinfo endpoint.
type GoogleProviderAdapter struct {
	oauth2Config  *oauth2.Config
	userInfoURI   string
	allowedEmails []string
}

// NewGoogleProviderAdapter returns a ProviderRepository wired to the given OAuth2 config.
func NewGoogleProviderAdapter(oauth2Config *oauth2.Config, userInfoURI string, allowedEmails []string) ports.ProviderRepository {
	return &GoogleProviderAdapter{
		oauth2Config:  oauth2Config,
		userInfoURI:   userInfoURI,
		allowedEmails: allowedEmails,
	}
}

func (g *GoogleProviderAdapter) GetUserInfo(ctx context.Context, code string) (*model.UserInfo, error) {
	token, err := g.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	client := g.oauth2Config.Client(ctx, token)
	resp, err := client.Get(g.userInfoURI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &ProviderError{StatusCode: resp.StatusCode}
	}

	var gUserInfo googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&gUserInfo); err != nil {
		return nil, err
	}

	return &model.UserInfo{
		Email:         gUserInfo.Email,
		FirstName:     gUserInfo.FirstName,
		LastName:      gUserInfo.LastName,
		Picture:       gUserInfo.Picture,
		AllowedEmails: g.allowedEmails,
	}, nil
}

// ProviderError is returned when the upstream provider returns a non-200 status.
type ProviderError struct {
	StatusCode int
}

func (e *ProviderError) Error() string {
	return "provider returned status " + http.StatusText(e.StatusCode)
}
