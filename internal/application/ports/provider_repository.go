package ports

import (
	"context"

	"github.com/matiasmartin-labs/auth-provider-ms/internal/domain/model"
)

type ProviderRepository interface {
	GetUserInfo(ctx context.Context, code string) (*model.UserInfo, error)
}
