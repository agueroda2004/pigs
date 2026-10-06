package ports

import (
	"context"

	"github.com/google/uuid"

	breeddomain "server/internal/modules/breed/domain"
)

// BreedFilter narrows the breeds returned by the list operation.
// A nil field is ignored so an empty filter returns every breed; the name
// filter is matched partially and case-insensitively.
type BreedFilter struct {
	Name   *string
	Active *bool
}

type BreedRepository interface {
	Create(ctx context.Context, breed *breeddomain.Breed) error
	GetByID(ctx context.Context, id uuid.UUID) (*breeddomain.Breed, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistsByNameExcludingID(ctx context.Context, name string, id uuid.UUID) (bool, error)
	List(ctx context.Context, filter BreedFilter) ([]*breeddomain.Breed, error)
	ListDropdown(ctx context.Context, active bool) ([]breeddomain.BreedDropdown, error)
	Update(ctx context.Context, breed *breeddomain.Breed) error
	Delete(ctx context.Context, id uuid.UUID) error
}
