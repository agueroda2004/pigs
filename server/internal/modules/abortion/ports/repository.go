package ports

import (
	"context"

	"github.com/google/uuid"

	abortiondomain "server/internal/modules/abortion/domain"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

// AbortionFilter narrows the abortions returned by the list operation.
// A nil field is ignored so an empty filter returns every abortion.
type AbortionFilter struct {
	SowID *uuid.UUID
}

// AbortionRepository defines the persistence operations of the abortion module.
// It also exposes the sow and last-service lookups required to register an
// abortion and the atomic write that inserts it and updates both states.
type AbortionRepository interface {
	GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	GetLastService(ctx context.Context, sowID uuid.UUID) (*servicedomain.Service, error)
	Create(ctx context.Context, abortion *abortiondomain.Abortion, sow *sowdomain.Sow, service *servicedomain.Service) error
	List(ctx context.Context, filter AbortionFilter) ([]*abortiondomain.Abortion, error)
}
