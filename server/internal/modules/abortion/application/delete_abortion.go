package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"server/internal/modules/abortion/ports"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

// DeleteAbortionCommand identifies the abortion to delete and its actor.
// The states to restore are derived by the use case.
type DeleteAbortionCommand struct {
	ID        uuid.UUID
	DeletedBy uuid.UUID
}

type DeleteAbortionService struct {
	repository ports.AbortionRepository
	clock      func() time.Time
}

// NewDeleteAbortionService builds a delete-abortion use case with its repository
// and clock, ready to execute DeleteAbortionCommand values.
func NewDeleteAbortionService(repository ports.AbortionRepository, clock func() time.Time) *DeleteAbortionService {
	return &DeleteAbortionService{repository: repository, clock: clock}
}

// Execute deletes an abortion and restores the sow to Gestando and its service
// to Confirmado. It blocks the deletion when the sow already has a mount or a sow
// removal dated after the abortion, and persists every change in one transaction.
func (s *DeleteAbortionService) Execute(ctx context.Context, command DeleteAbortionCommand) error {
	abortion, err := s.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	hasFutureEvents, err := s.repository.HasFutureEvents(ctx, abortion.SowID, abortion.AbortionDate)
	if err != nil {
		return err
	}
	if hasFutureEvents {
		return ports.ErrAbortionNotDeletable
	}

	sow, err := s.repository.GetSow(ctx, abortion.SowID)
	if err != nil {
		return err
	}

	service, err := s.repository.GetService(ctx, abortion.ServiceID)
	if err != nil {
		return err
	}

	now := s.clock()
	if err := sow.ChangeState(sowdomain.StatePregnant, command.DeletedBy, now); err != nil {
		return err
	}
	if err := service.ChangeState(servicedomain.StateConfirmed, command.DeletedBy, now); err != nil {
		return err
	}

	return s.repository.Delete(ctx, abortion, sow, service)
}
