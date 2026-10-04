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
// range is inclusive on both ends.
type FarrowingFilter struct {
	SowID     *uuid.UUID
	ServiceID *uuid.UUID
	FromDate  *time.Time
	ToDate    *time.Time
}

// FarrowingRepository defines the persistence operations of the farrowing module.
// It also exposes the sow, service, operator and medication lookups required to
// register a farrowing and the atomic write that inserts the farrowing, its
// join rows and the resulting sow and service state changes.
type FarrowingRepository interface {
	GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	GetLastService(ctx context.Context, sowID uuid.UUID) (*servicedomain.Service, error)
	GetOperator(ctx context.Context, id uuid.UUID) (*operatordomain.Operator, error)
	GetMedication(ctx context.Context, id uuid.UUID) (*medicationdomain.Medication, error)
	Create(ctx context.Context, farrowing *farrowingdomain.Farrowing, sow *sowdomain.Sow, service *servicedomain.Service) error
	List(ctx context.Context, filter FarrowingFilter) ([]*farrowingdomain.Farrowing, error)
}
