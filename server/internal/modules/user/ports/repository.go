package ports

import (
	"context"

	"github.com/google/uuid"

	userdomain "server/internal/modules/user/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *userdomain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*userdomain.User, error)
	List(ctx context.Context) ([]*userdomain.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	Update(ctx context.Context, user *userdomain.User) error
}
