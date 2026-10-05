package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	sowdomain "server/internal/modules/sow/domain"
)

// PigletFosteringFilter narrows the fosterings returned by the list operation.
// A nil field is ignored so an empty filter returns every fostering; the date
// range is inclusive on both ends and each sow matches its side of the movement.
type PigletFosteringFilter struct {
	DonorSowID    *uuid.UUID
	ReceiverSowID *uuid.UUID
	FromDate      *time.Time
	ToDate        *time.Time
}

// PigletFosteringRepository defines the persistence operations of the fostering
// module. It also exposes the sow and farrowing lookups required to register a
// transfer and the atomic write that inserts it while decreasing the donor and
// increasing the receiver current piglets balances.
type PigletFosteringRepository interface {
	GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	GetLastFarrowing(ctx context.Context, sowID uuid.UUID) (*farrowingdomain.Farrowing, error)
	Create(ctx context.Context, fostering *pigletfosteringdomain.PigletFostering, donor *farrowingdomain.Farrowing, receiver *farrowingdomain.Farrowing) error
	List(ctx context.Context, filter PigletFosteringFilter) ([]*pigletfosteringdomain.PigletFostering, error)
}
