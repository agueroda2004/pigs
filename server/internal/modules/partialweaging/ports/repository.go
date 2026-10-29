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
// range is inclusive on both ends and the sow is matched by its code.
type PartialWeagingFilter struct {
	SowCode  *string
	FromDate *time.Time
	ToDate   *time.Time
}

// PartialWeagingRepository defines the persistence operations of the partial weaging
// module. It also exposes the sow and farrowing lookups required to register a partial
// weaging, the latest related event and partial weaging dates used to validate the
// weaging date, and the atomic write that inserts it while reducing the farrowing
// piglets and updating the sow state and the nurse flags. Deletion exposes the event
// lookups that block it and the atomic write that restores the farrowing and sow.
// Updates persist the editable weight, note and audit fields only.
type PartialWeagingRepository interface {
	GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	GetLastFarrowing(ctx context.Context, sowID uuid.UUID) (*farrowingdomain.Farrowing, error)
	GetLatestEventDate(ctx context.Context, farrowingID uuid.UUID) (time.Time, error)
	GetLastPartialWeagingDate(ctx context.Context, farrowingID uuid.UUID) (time.Time, error)
	Create(ctx context.Context, weaging *partialweagingdomain.PartialWeaging, farrowing *farrowingdomain.Farrowing, sow *sowdomain.Sow) error
	GetByID(ctx context.Context, id uuid.UUID) (*partialweagingdomain.PartialWeaging, error)
	GetFarrowingByID(ctx context.Context, id uuid.UUID) (*farrowingdomain.Farrowing, error)
	// List returns one page of partial weagings matching the filter together with
	// the total count of matches, so callers can paginate the result.
	List(ctx context.Context, filter PartialWeagingFilter, limit int, offset int) ([]*partialweagingdomain.PartialWeaging, int, error)
	HasFutureEvents(ctx context.Context, sowID uuid.UUID, after time.Time) (bool, error)
	HasLaterNurseEvents(ctx context.Context, farrowingID uuid.UUID, after time.Time) (bool, error)
	GetPreviousNurseDate(ctx context.Context, farrowingID uuid.UUID, before time.Time) (time.Time, error)
	Update(ctx context.Context, weaging *partialweagingdomain.PartialWeaging) error
	Delete(ctx context.Context, weaging *partialweagingdomain.PartialWeaging, farrowing *farrowingdomain.Farrowing, sow *sowdomain.Sow) error
}
