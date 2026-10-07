package ports

import (
	"context"
	"time"

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
	// List returns one page of sows matching the filter plus the total count
	// of matches, so callers can paginate the result.
	List(ctx context.Context, filter SowFilter, limit, offset int) ([]*sowdomain.Sow, int, error)
	// ListDropdown returns lightweight sow read models filtered by the active
	// flag and an optional list of states; empty states apply no state filter.
	ListDropdown(ctx context.Context, active *bool, states []sowdomain.State) ([]sowdomain.SowDropdown, error)
	// LastServiceDate returns the most recent mount date registered for a sow
	// or nil when the sow has no service yet.
	LastServiceDate(ctx context.Context, sowID uuid.UUID) (*time.Time, error)
	Update(ctx context.Context, sow *sowdomain.Sow) error
	UpdateState(ctx context.Context, sow *sowdomain.Sow) error
}
