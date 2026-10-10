package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	abortiondomain "server/internal/modules/abortion/domain"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

// AbortionFilter narrows the abortions returned by the list operation.
// A nil field is ignored so an empty filter returns every abortion. SowCode
// matches the code of the related sow ignoring case.
type AbortionFilter struct {
	SowCode *string
}

// AbortionRepository defines the persistence operations of the abortion module.
// It also exposes the sow and last-service lookups required to register an
// abortion and the atomic write that inserts it and updates both states.
type AbortionRepository interface {
	GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error)
	// GetByID fetches an abortion by its identifier.
	// It returns ErrAbortionNotFound when no row matches.
	GetByID(ctx context.Context, id uuid.UUID) (*abortiondomain.Abortion, error)
	// GetService fetches the service that produced the abortion, with its mounts.
	// It returns ErrServiceNotFound when no row matches.
	GetService(ctx context.Context, id uuid.UUID) (*servicedomain.Service, error)
	GetLastService(ctx context.Context, sowID uuid.UUID) (*servicedomain.Service, error)
	Create(ctx context.Context, abortion *abortiondomain.Abortion, sow *sowdomain.Sow, service *servicedomain.Service) error
	// Update persists the editable fields of an abortion.
	Update(ctx context.Context, abortion *abortiondomain.Abortion) error
	// HasFutureEvents reports whether the sow has a mount or a sow removal dated
	// after the given reference date, which blocks deleting its abortion.
	HasFutureEvents(ctx context.Context, sowID uuid.UUID, after time.Time) (bool, error)
	// Delete removes an abortion and restores the sow and its service states in
	// one atomic transaction.
	Delete(ctx context.Context, abortion *abortiondomain.Abortion, sow *sowdomain.Sow, service *servicedomain.Service) error
	// List returns one page of abortions matching the filter together with the
	// total count of matches, so callers can paginate the result.
	List(ctx context.Context, filter AbortionFilter, limit, offset int) ([]*abortiondomain.Abortion, int, error)
}
