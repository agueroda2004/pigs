package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
	sowdomain "server/internal/modules/sow/domain"
)

// DeleteServiceCommand identifies the service to delete and its actor.
// The service state and the previous sow state are derived by the use case.
type DeleteServiceCommand struct {
	ID        uuid.UUID
	DeletedBy uuid.UUID
}

type DeleteServiceService struct {
	repository ports.ServiceRepository
	clock      func() time.Time
}

// NewDeleteServiceService builds a delete-service use case with its repository
// and clock, ready to execute DeleteServiceCommand values.
func NewDeleteServiceService(repository ports.ServiceRepository, clock func() time.Time) *DeleteServiceService {
	return &DeleteServiceService{repository: repository, clock: clock}
}

// Execute deletes a confirmed service and restores the sow to its previous state.
// It blocks any service that is not in StateConfirmed because those states already
// have related events, reopens the previous failed service and persists the
// service, its mounts, the previous service and the sow state in one transaction.
func (s *DeleteServiceService) Execute(ctx context.Context, command DeleteServiceCommand) error {
	service, err := s.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	if err := service.EnsureDeletable(); err != nil {
		return err
	}

	lastState, err := sowdomain.ParseState(service.LastState)
	if err != nil {
		return err
	}

	sow, err := s.repository.GetSow(ctx, service.SowID)
	if err != nil {
		return err
	}

	previous, err := s.repository.GetPreviousService(ctx, service.SowID, service.ID)
	if err != nil && !errors.Is(err, ports.ErrServiceNotFound) {
		return err
	}

	now := s.clock()
	if err := sow.ChangeState(lastState, command.DeletedBy, now); err != nil {
		return err
	}

	var previousService *servicedomain.Service
	if err == nil && previous != nil && previous.State == servicedomain.StateFailed {
		if err := previous.ChangeState(servicedomain.StateConfirmed, command.DeletedBy, now); err != nil {
			return err
		}
		previousService = previous
	}

	return s.repository.Delete(ctx, service, sow, previousService)
}
