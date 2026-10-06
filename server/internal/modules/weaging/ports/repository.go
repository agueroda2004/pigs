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
// range is inclusive on both ends.
type WeagingFilter struct {
	SowID    *uuid.UUID
	FromDate *time.Time
	ToDate   *time.Time
}

// WeagingRepository defines the persistence operations of the weaging module.
// It also exposes the sow and farrowing lookups required to register a weaging,
// the latest related event date used to validate the weaging date, and the atomic
// write that inserts it while zeroing the farrowing piglets and weaning the sow.
type WeagingRepository interface {
	GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	GetLastFarrowing(ctx context.Context, sowID uuid.UUID) (*farrowingdomain.Farrowing, error)
	GetLatestEventDate(ctx context.Context, farrowingID uuid.UUID) (time.Time, error)
	Create(ctx context.Context, weaging *weagingdomain.Weaging, farrowing *farrowingdomain.Farrowing, sow *sowdomain.Sow) error
	List(ctx context.Context, filter WeagingFilter) ([]*weagingdomain.Weaging, error)
}
