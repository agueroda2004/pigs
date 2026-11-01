package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	medicationdomain "server/internal/modules/medication/domain"
	operatordomain "server/internal/modules/operator/domain"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

// FarrowingFilter narrows the farrowings returned by the list operation.
// A nil field is ignored so an empty filter returns every farrowing; the date
// range is inclusive on both ends and SowCode matches the related sow code
// ignoring case.
type FarrowingFilter struct {
	SowCode  *string
	FromDate *time.Time
	ToDate   *time.Time
}

// FarrowingRepository defines the persistence operations of the farrowing module.
// It also exposes the sow, service, operator and medication lookups required to
// register a farrowing and the atomic write that inserts the farrowing, its
// join rows and the resulting sow and service state changes.
type FarrowingRepository interface {
	GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	GetLastService(ctx context.Context, sowID uuid.UUID) (*servicedomain.Service, error)
	// GetByID fetches a farrowing by its identifier with its sow code and links.
	// It returns ErrFarrowingNotFound when no row matches.
	GetByID(ctx context.Context, id uuid.UUID) (*farrowingdomain.Farrowing, error)
	// GetService fetches a service by its identifier with its mounts.
	// It returns ErrServiceNotFound when no row matches.
	GetService(ctx context.Context, id uuid.UUID) (*servicedomain.Service, error)
	GetOperator(ctx context.Context, id uuid.UUID) (*operatordomain.Operator, error)
	GetMedication(ctx context.Context, id uuid.UUID) (*medicationdomain.Medication, error)
	// GetEarliestRelatedDate returns the earliest date among the deaths, fosterings,
	// partial weagings and weagings linked to a farrowing, or nil when it has none.
	GetEarliestRelatedDate(ctx context.Context, farrowingID uuid.UUID) (*time.Time, error)
	Create(ctx context.Context, farrowing *farrowingdomain.Farrowing, sow *sowdomain.Sow, service *servicedomain.Service) error
	// Update persists the editable fields of a farrowing and optionally replaces
	// its operator and medication links.
	Update(ctx context.Context, farrowing *farrowingdomain.Farrowing, replaceOperators, replaceMedications bool) error
	// Delete removes a farrowing and, in the same transaction, persists the
	// restored sow and service states. It returns ErrFarrowingHasRelatedRecords
	// when the farrowing still has linked deaths, fosterings or weagings.
	Delete(ctx context.Context, farrowing *farrowingdomain.Farrowing, sow *sowdomain.Sow, service *servicedomain.Service) error
	// List returns one page of farrowings matching the filter together with the
	// total count of matches, so callers can paginate the result.
	List(ctx context.Context, filter FarrowingFilter, limit, offset int) ([]*farrowingdomain.Farrowing, int, error)
	// HistoryBySow returns the farrowing history read models of a sow code with
	// their operator and medication names, or an empty slice when it has none.
	HistoryBySow(ctx context.Context, sowCode string) ([]*farrowingdomain.FarrowingHistoryBySow, error)
}
