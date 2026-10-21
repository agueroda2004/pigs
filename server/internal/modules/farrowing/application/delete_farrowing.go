package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"server/internal/modules/farrowing/ports"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

// DeleteFarrowingCommand identifies the farrowing to delete and its actor.
// The states to restore and the parity change are derived by the use case.
type DeleteFarrowingCommand struct {
	ID        uuid.UUID
	DeletedBy uuid.UUID
}

type DeleteFarrowingService struct {
	repository ports.FarrowingRepository
	clock      func() time.Time
}

// NewDeleteFarrowingService builds a delete-farrowing use case with its
// repository and clock, ready to execute DeleteFarrowingCommand values.
func NewDeleteFarrowingService(repository ports.FarrowingRepository, clock func() time.Time) *DeleteFarrowingService {
	return &DeleteFarrowingService{repository: repository, clock: clock}
}

// Execute deletes a farrowing, moves the sow back to Gestando (decrementing its
// parity) and its service back to Confirmado.
// It blocks the deletion when the farrowing still has linked records and
// persists every change in one transaction.
func (s *DeleteFarrowingService) Execute(ctx context.Context, command DeleteFarrowingCommand) error {
	farrowing, err := s.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	earliest, err := s.repository.GetEarliestRelatedDate(ctx, command.ID)
	if err != nil {
		return err
	}
	if earliest != nil {
		return ports.ErrFarrowingHasRelatedRecords
	}

	sow, err := s.repository.GetSow(ctx, farrowing.SowID)
	if err != nil {
		return err
	}

	service, err := s.repository.GetService(ctx, farrowing.ServiceID)
	if err != nil {
		return err
	}

	now := s.clock()
	if err := sow.UnregisterFarrowing(command.DeletedBy, now); err != nil {
		return err
	}
	if err := sow.ChangeState(sowdomain.StatePregnant, command.DeletedBy, now); err != nil {
		return err
	}
	if err := service.ChangeState(servicedomain.StateConfirmed, command.DeletedBy, now); err != nil {
		return err
	}

	return s.repository.Delete(ctx, farrowing, sow, service)
}
