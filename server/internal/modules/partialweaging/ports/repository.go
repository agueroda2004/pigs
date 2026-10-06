package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	partialweagingdomain "server/internal/modules/partialweaging/domain"
	sowdomain "server/internal/modules/sow/domain"
)

// PartialWeagingFilter narrows the partial weagings returned by the list operation.
// A nil field is ignored so an empty filter returns every partial weaging; the date
// range is inclusive on both ends.
type PartialWeagingFilter struct {
	SowID    *uuid.UUID
	FromDate *time.Time
	ToDate   *time.Time
}

// PartialWeagingRepository defines the persistence operations of the partial weaging
// module. It also exposes the sow and farrowing lookups required to register a partial
// weaging, the latest related event date used to validate the weaging date, and the
// atomic write that inserts it while reducing the farrowing piglets and updating the
// sow state and the nurse flags.
type PartialWeagingRepository interface {
	GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	GetLastFarrowing(ctx context.Context, sowID uuid.UUID) (*farrowingdomain.Farrowing, error)
	GetLatestEventDate(ctx context.Context, farrowingID uuid.UUID) (time.Time, error)
	Create(ctx context.Context, weaging *partialweagingdomain.PartialWeaging, farrowing *farrowingdomain.Farrowing, sow *sowdomain.Sow) error
	List(ctx context.Context, filter PartialWeagingFilter) ([]*partialweagingdomain.PartialWeaging, error)
}
