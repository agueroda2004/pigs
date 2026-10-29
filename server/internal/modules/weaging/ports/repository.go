package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	sowdomain "server/internal/modules/sow/domain"
	weagingdomain "server/internal/modules/weaging/domain"
)

// WeagingFilter narrows the weagings returned by the list operation.
// A nil field is ignored so an empty filter returns every weaging; the date
// range is inclusive on both ends and SowCode matches the related sow code
// ignoring case.
type WeagingFilter struct {
	SowCode  *string
	FromDate *time.Time
	ToDate   *time.Time
}

// WeagingRepository defines the persistence operations of the weaging module.
// It also exposes the sow and farrowing lookups required to register a weaging,
// the latest related event date used to validate the weaging date, the update of
// its editable fields, and the atomic writes that insert it while zeroing the
// farrowing piglets and weaning the sow, or remove it while restoring both.
type WeagingRepository interface {
	GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	GetByID(ctx context.Context, id uuid.UUID) (*weagingdomain.Weaging, error)
	GetFarrowingByID(ctx context.Context, id uuid.UUID) (*farrowingdomain.Farrowing, error)
	GetLastFarrowing(ctx context.Context, sowID uuid.UUID) (*farrowingdomain.Farrowing, error)
	GetLatestEventDate(ctx context.Context, farrowingID uuid.UUID) (time.Time, error)
	HasFutureEvents(ctx context.Context, sowID uuid.UUID, after time.Time) (bool, error)
	Create(ctx context.Context, weaging *weagingdomain.Weaging, farrowing *farrowingdomain.Farrowing, sow *sowdomain.Sow) error
	Update(ctx context.Context, weaging *weagingdomain.Weaging) error
	Delete(ctx context.Context, weaging *weagingdomain.Weaging, farrowing *farrowingdomain.Farrowing, sow *sowdomain.Sow) error
	List(ctx context.Context, filter WeagingFilter, limit, offset int) ([]*weagingdomain.Weaging, int, error)
}
