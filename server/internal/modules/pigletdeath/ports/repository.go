package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	operatordomain "server/internal/modules/operator/domain"
	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	sowdomain "server/internal/modules/sow/domain"
)

// PigletDeathFilter narrows the piglet deaths returned by the list operation.
// A nil field is ignored so an empty filter returns every death; the date range
// is inclusive on both ends and the sow is matched by its code through the farrowing.
type PigletDeathFilter struct {
	SowCode  *string
	FromDate *time.Time
	ToDate   *time.Time
}

// PigletDeathRepository defines the persistence operations of the piglet death
// module. It also exposes the sow, farrowing, partial weaging and operator lookups
// required to register a death and the atomic write that inserts it while reducing
// the current piglets balance of the farrowing, plus the delete that restores it.
type PigletDeathRepository interface {
	GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	GetLastFarrowing(ctx context.Context, sowID uuid.UUID) (*farrowingdomain.Farrowing, error)
	GetLastPartialWeagingDate(ctx context.Context, farrowingID uuid.UUID) (time.Time, error)
	GetOperator(ctx context.Context, id uuid.UUID) (*operatordomain.Operator, error)
	GetByID(ctx context.Context, id uuid.UUID) (*pigletdeathdomain.PigletDeath, error)
	GetFarrowingByID(ctx context.Context, id uuid.UUID) (*farrowingdomain.Farrowing, error)
	HasWeagingAfter(ctx context.Context, farrowingID uuid.UUID, date time.Time) (bool, error)
	HasPartialWeagingAfter(ctx context.Context, farrowingID uuid.UUID, date time.Time) (bool, error)
	Create(ctx context.Context, death *pigletdeathdomain.PigletDeath, farrowing *farrowingdomain.Farrowing) error
	Update(ctx context.Context, death *pigletdeathdomain.PigletDeath) error
	Delete(ctx context.Context, death *pigletdeathdomain.PigletDeath, farrowing *farrowingdomain.Farrowing) error
	// List returns one page of piglet deaths matching the filter together with
	// the total count of matches, so callers can paginate the result.
	List(ctx context.Context, filter PigletDeathFilter, limit int, offset int) ([]*pigletdeathdomain.PigletDeath, int, error)
}
