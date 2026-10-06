package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
)

// BoarFilter holds the optional criteria used to filter the boar list.
// A nil field means the criterion is not applied and every filter combines with AND.
type BoarFilter struct {
	Code    *string
	BreedID *uuid.UUID
	Origin  *boardomain.Origin
	Active  *bool
}

type BoarRepository interface {
	Create(ctx context.Context, boar *boardomain.Boar) error
	GetByID(ctx context.Context, id uuid.UUID) (*boardomain.Boar, error)
	GetEarliestEventDate(ctx context.Context, boarID uuid.UUID) (time.Time, error)
	ExistsByCode(ctx context.Context, code string) (bool, error)
	List(ctx context.Context, filter BoarFilter) ([]*boardomain.Boar, error)
	ListDropdown(ctx context.Context, active *bool, state *boardomain.State) ([]boardomain.BoarDropdown, error)
	Update(ctx context.Context, boar *boardomain.Boar) error
	UpdateState(ctx context.Context, boar *boardomain.Boar) error
	Delete(ctx context.Context, id uuid.UUID) error
}
