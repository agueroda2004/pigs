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
// is inclusive on both ends and the sow is matched through the farrowing.
type PigletDeathFilter struct {
	SowID    *uuid.UUID
	FromDate *time.Time
	ToDate   *time.Time
}

// PigletDeathRepository defines the persistence operations of the piglet death
// module. It also exposes the sow, farrowing and operator lookups required to
// register a death and the atomic write that inserts it while reducing the
// current piglets balance of the farrowing.
type PigletDeathRepository interface {
	GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	GetLastFarrowing(ctx context.Context, sowID uuid.UUID) (*farrowingdomain.Farrowing, error)
	GetOperator(ctx context.Context, id uuid.UUID) (*operatordomain.Operator, error)
	Create(ctx context.Context, death *pigletdeathdomain.PigletDeath, farrowing *farrowingdomain.Farrowing) error
	List(ctx context.Context, filter PigletDeathFilter) ([]*pigletdeathdomain.PigletDeath, error)
}
