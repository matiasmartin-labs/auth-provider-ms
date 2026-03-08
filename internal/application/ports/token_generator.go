package ports

import "github.com/matiasmartin-labs/auth-provider-ms/internal/domain/model"

type TokenGenerator interface {
	GenerateToken(*model.UserInfo) (string, error)
}
