package application

import (
	"context"
	"time"

	"github.com/google/uuid"

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
// have related events, and persists the service, its mounts and the sow state in
// one transaction.
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

	now := s.clock()
	if err := sow.ChangeState(lastState, command.DeletedBy, now); err != nil {
		return err
	}

	return s.repository.Delete(ctx, service, sow)
}
