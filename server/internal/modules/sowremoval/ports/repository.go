package ports

import (
	"context"

	"github.com/google/uuid"

	abortiondomain "server/internal/modules/abortion/domain"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
	sowremovaldomain "server/internal/modules/sowremoval/domain"
)

// SowRemovalFilter narrows the removals returned by the list operation.
// A nil field is ignored so an empty filter returns every removal.
type SowRemovalFilter struct {
	SowID *uuid.UUID
}

// SowRemovalRepository defines the persistence operations of the removal module.
// It also exposes the sow, last-service and last-abortion lookups required to
// register a removal and the atomic writes that insert, update or delete the
// removal along with the resulting sow and service states.
type SowRemovalRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*sowremovaldomain.SowRemoval, error)
	GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	GetLastService(ctx context.Context, sowID uuid.UUID) (*servicedomain.Service, error)
	GetLastAbortion(ctx context.Context, sowID uuid.UUID) (*abortiondomain.Abortion, error)
	Create(ctx context.Context, removal *sowremovaldomain.SowRemoval, sow *sowdomain.Sow, service *servicedomain.Service) error
	Update(ctx context.Context, removal *sowremovaldomain.SowRemoval, sow *sowdomain.Sow) error
	Delete(ctx context.Context, removal *sowremovaldomain.SowRemoval, sow *sowdomain.Sow, service *servicedomain.Service) error
	List(ctx context.Context, filter SowRemovalFilter) ([]*sowremovaldomain.SowRemoval, error)
}
