package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	sowdomain "server/internal/modules/sow/domain"
	weagingdomain "server/internal/modules/weaging/domain"
	"server/internal/modules/weaging/ports"
)

// DeleteWeagingCommand identifies the weaging to delete and its actor.
// The farrowing, sow and restored quantity are derived from the stored record.
type DeleteWeagingCommand struct {
	ID        uuid.UUID
	DeletedBy uuid.UUID
}

type DeleteWeagingService struct {
	repository ports.WeagingRepository
	clock      func() time.Time
}

// NewDeleteWeagingService builds a delete-weaging use case with its repository and clock.
// It returns a service ready to execute DeleteWeagingCommand values.
func NewDeleteWeagingService(repository ports.WeagingRepository, clock func() time.Time) *DeleteWeagingService {
	return &DeleteWeagingService{repository: repository, clock: clock}
}

// Execute deletes a weaging and gives its quantity back to the farrowing.
// It blocks the deletion when the sow has a later mount or removal, then restores
// the farrowing balance and moves the sow back to Lactando in one transaction.
func (s *DeleteWeagingService) Execute(ctx context.Context, command DeleteWeagingCommand) error {
	weaging, err := s.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	farrowing, err := s.repository.GetFarrowingByID(ctx, weaging.FarrowingID)
	if err != nil {
		return err
	}

	hasFutureEvents, err := s.repository.HasFutureEvents(ctx, weaging.SowID, weaging.WeagingDate)
	if err != nil {
		return err
	}
	if hasFutureEvents {
		return weagingdomain.ErrHasFutureEvents
	}

	now := s.clock()

	sow, err := s.repository.GetSow(ctx, weaging.SowID)
	if err != nil {
		return err
	}
	if err := sow.ChangeState(sowdomain.StateLactating, command.DeletedBy, now); err != nil {
		return err
	}

	if err := farrowing.AddCurrentPiglets(weaging.Quantity, command.DeletedBy, now); err != nil {
		return err
	}

	return s.repository.Delete(ctx, weaging, farrowing, sow)
}
