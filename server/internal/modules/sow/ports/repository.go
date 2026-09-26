package ports

import (
	"context"

	"github.com/google/uuid"

	sowdomain "server/internal/modules/sow/domain"
)

// SowFilter holds the optional criteria used to filter the sow list.
// A nil field means the criterion is not applied and every filter combines with AND.
type SowFilter struct {
	Code    *string
	BreedID *uuid.UUID
	Origin  *sowdomain.Origin
	Active  *bool
	State   *sowdomain.State
}

type SowRepository interface {
	Create(ctx context.Context, sow *sowdomain.Sow) error
	GetByID(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	ExistsByCode(ctx context.Context, code string) (bool, error)
	List(ctx context.Context, filter SowFilter) ([]*sowdomain.Sow, error)
	// ListOptions returns lightweight sow read models; a true active restricts
	// to serviceable states while nil or false return every state.
	ListOptions(ctx context.Context, active *bool) ([]sowdomain.SowOption, error)
	Update(ctx context.Context, sow *sowdomain.Sow) error
	UpdateState(ctx context.Context, sow *sowdomain.Sow) error
}
