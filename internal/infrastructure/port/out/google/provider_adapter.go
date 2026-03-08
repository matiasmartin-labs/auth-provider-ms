package google

import (
	"context"
	"encoding/json"

	"github.com/matiasmartin-labs/auth-provider-ms/internal/application/ports"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/domain/model"
	"github.com/matiasmartin-labs/auth-provider-ms/pkg"
)

type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
}

type GoogleProviderAdapter struct{}

func NewGoogleProviderAdapter() ports.ProviderRepository {
	return &GoogleProviderAdapter{}
}

func (g *GoogleProviderAdapter) GetUserInfo(ctx context.Context, code string) (*model.UserInfo, error) {
	token, err := pkg.GoogleOAuth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	googleConfig := pkg.App.Config.GetSecurityConfig().GetOAuth2Config().GetGoogleConfig()

	client := pkg.GoogleOAuth2Config.Client(ctx, token)
	resp, err := client.Get(googleConfig.GetUserInfoURI())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gUserInfo googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&gUserInfo); err != nil {
		return nil, err
	}

	userInfo := &model.UserInfo{
		Email:     gUserInfo.Email,
		FirstName: gUserInfo.FirstName,
		LastName:  gUserInfo.LastName,
		Picture:   gUserInfo.Picture,
	}

	return userInfo, nil
}
