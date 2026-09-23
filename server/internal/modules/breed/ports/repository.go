package ports

import (
	"context"

	"github.com/google/uuid"

	breeddomain "server/internal/modules/breed/domain"
)

type BreedRepository interface {
	Create(ctx context.Context, breed *breeddomain.Breed) error
	GetByID(ctx context.Context, id uuid.UUID) (*breeddomain.Breed, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	List(ctx context.Context) ([]*breeddomain.Breed, error)
	Update(ctx context.Context, breed *breeddomain.Breed) error
}
