package ports

import (
	"context"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	operatordomain "server/internal/modules/operator/domain"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

// ServiceFilter narrows the services returned by the list operation.
// Nil fields are ignored so an empty filter returns every service.
type ServiceFilter struct {
	SowID *uuid.UUID
	State *servicedomain.State
}

// ServiceRepository defines the persistence operations of the service module.
// It also exposes the sow, boar and operator lookups required to register a
// service and the atomic write that inserts the service, its mounts and the
// resulting sow state change.
type ServiceRepository interface {
	GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	GetBoar(ctx context.Context, id uuid.UUID) (*boardomain.Boar, error)
	GetOperator(ctx context.Context, id uuid.UUID) (*operatordomain.Operator, error)
	Create(ctx context.Context, service *servicedomain.Service, sow *sowdomain.Sow) error
	List(ctx context.Context, filter ServiceFilter) ([]*servicedomain.Service, error)
}
