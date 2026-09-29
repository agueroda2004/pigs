package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	boarremovaldomain "server/internal/modules/boarremoval/domain"
)

// BoarRemovalFilter narrows the removals returned by the list operation.
// A nil field is ignored so an empty filter returns every removal.
type BoarRemovalFilter struct {
	BoarID *uuid.UUID
}

// BoarRemovalRepository defines the persistence operations of the removal module.
// It also exposes the boar and last-mount lookups required to register a removal
// and the atomic writes that insert, update or delete it along with the resulting
// boar state.
type BoarRemovalRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*boarremovaldomain.BoarRemoval, error)
	GetBoar(ctx context.Context, id uuid.UUID) (*boardomain.Boar, error)
	GetLastMountDate(ctx context.Context, boarID uuid.UUID) (time.Time, error)
	Create(ctx context.Context, removal *boarremovaldomain.BoarRemoval, boar *boardomain.Boar) error
	Update(ctx context.Context, removal *boarremovaldomain.BoarRemoval, boar *boardomain.Boar) error
	Delete(ctx context.Context, removal *boarremovaldomain.BoarRemoval, boar *boardomain.Boar) error
	List(ctx context.Context, filter BoarRemovalFilter) ([]*boarremovaldomain.BoarRemoval, error)
}
