package ports

import (
	"context"

	"github.com/google/uuid"

	userdomain "server/internal/modules/user/domain"
)

type UserReader interface {
	FindByUsername(ctx context.Context, username string) (*userdomain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*userdomain.User, error)
}
