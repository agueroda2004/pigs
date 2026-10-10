package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	operatordomain "server/internal/modules/operator/domain"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

// ServiceFilter narrows the services returned by the list operation.
// Nil fields are ignored so an empty filter returns every service. SowCode
// matches the code of the related sow ignoring case.
type ServiceFilter struct {
	SowCode *string
	State   *servicedomain.State
}

// ServiceRepository defines the persistence operations of the service module.
// It also exposes the sow, boar and operator lookups required to register a
// service and the atomic write that inserts the service, its mounts and the
// resulting sow state change.
type ServiceRepository interface {
	GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	GetBoar(ctx context.Context, id uuid.UUID) (*boardomain.Boar, error)
	GetOperator(ctx context.Context, id uuid.UUID) (*operatordomain.Operator, error)
	// GetByID fetches a service with its mounts by its identifier.
	GetByID(ctx context.Context, id uuid.UUID) (*servicedomain.Service, error)
	// GetLastService fetches the most recent service of a sow with its mounts.
	// It returns ErrServiceNotFound when the sow has no service.
	GetLastService(ctx context.Context, sowID uuid.UUID) (*servicedomain.Service, error)
	// GetPreviousService fetches the most recent service of a sow other than the
	// excluded one, with its mounts. It returns ErrServiceNotFound when none exists.
	GetPreviousService(ctx context.Context, sowID uuid.UUID, excludeID uuid.UUID) (*servicedomain.Service, error)
	// GetLastAbortionDate returns the most recent abortion date of a sow.
	// It returns nil when the sow has no abortion.
	GetLastAbortionDate(ctx context.Context, sowID uuid.UUID) (*time.Time, error)
	Create(ctx context.Context, service *servicedomain.Service, sow *sowdomain.Sow) error
	// Update persists the service fields and replaces its mounts in one transaction.
	Update(ctx context.Context, service *servicedomain.Service) error
	// List returns one page of services matching the filter together with the
	// total count of matches, so callers can paginate the result.
	List(ctx context.Context, filter ServiceFilter, limit, offset int) ([]*servicedomain.Service, int, error)
	// Delete removes a service and its mounts, restoring the sow state in one
	// atomic transaction.
	Delete(ctx context.Context, service *servicedomain.Service, sow *sowdomain.Sow) error
}
