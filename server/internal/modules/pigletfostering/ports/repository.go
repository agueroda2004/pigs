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
// module. It also exposes the sow, farrowing and partial weaging lookups required
// to register or update a transfer, plus the atomic writes that move the piglets
// between both farrowings while persisting the record.
type PigletFosteringRepository interface {
	GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	GetLastFarrowing(ctx context.Context, sowID uuid.UUID) (*farrowingdomain.Farrowing, error)
	GetFarrowingByID(ctx context.Context, id uuid.UUID) (*farrowingdomain.Farrowing, error)
	GetLastPartialWeagingDate(ctx context.Context, farrowingID uuid.UUID) (time.Time, error)
	GetByID(ctx context.Context, id uuid.UUID) (*pigletfosteringdomain.PigletFostering, error)
	Create(ctx context.Context, fostering *pigletfosteringdomain.PigletFostering, donor *farrowingdomain.Farrowing, receiver *farrowingdomain.Farrowing) error
	Update(ctx context.Context, fostering *pigletfosteringdomain.PigletFostering) error
	List(ctx context.Context, filter PigletFosteringFilter) ([]*pigletfosteringdomain.PigletFostering, error)
}
