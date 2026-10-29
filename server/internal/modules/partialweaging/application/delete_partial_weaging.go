package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	partialweagingdomain "server/internal/modules/partialweaging/domain"
	"server/internal/modules/partialweaging/ports"
	sowdomain "server/internal/modules/sow/domain"
)

// DeletePartialWeagingCommand identifies the partial weaging to delete and its actor.
// The farrowing, sow and restored quantity are derived from the stored record.
type DeletePartialWeagingCommand struct {
	ID        uuid.UUID
	DeletedBy uuid.UUID
}

type DeletePartialWeagingService struct {
	repository ports.PartialWeagingRepository
	clock      func() time.Time
}

// NewDeletePartialWeagingService builds a delete-partial-weaging use case with its
// repository and clock, ready to execute DeletePartialWeagingCommand values.
func NewDeletePartialWeagingService(repository ports.PartialWeagingRepository, clock func() time.Time) *DeletePartialWeagingService {
	return &DeletePartialWeagingService{repository: repository, clock: clock}
}

// Execute deletes a partial weaging and gives its quantity back to the farrowing.
// A Nodriza delete is blocked by later deaths, fosterings, partial weagings or weagings
// and rolls the nurse flags back, while the other types are blocked by later mounts or
// sow removals and move the sow back to Lactando. Everything is persisted in one transaction.
func (s *DeletePartialWeagingService) Execute(ctx context.Context, command DeletePartialWeagingCommand) error {
	weaging, err := s.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	farrowing, err := s.repository.GetFarrowingByID(ctx, weaging.FarrowingID)
	if err != nil {
		return err
	}

	now := s.clock()
	var sow *sowdomain.Sow
	if weaging.Type == partialweagingdomain.TypeNodriza {
		if err := s.restoreNurseState(ctx, weaging, farrowing, command.DeletedBy, now); err != nil {
			return err
		}
	} else {
		sow, err = s.restoreSowState(ctx, weaging, command.DeletedBy, now)
		if err != nil {
			return err
		}
	}

	if err := farrowing.AddCurrentPiglets(weaging.Quantity, command.DeletedBy, now); err != nil {
		return err
	}

	return s.repository.Delete(ctx, weaging, farrowing, sow)
}

// restoreNurseState blocks a Nodriza delete when the farrowing has later events and
// otherwise restores the nurse flags to the previous Nodriza date, or clears them.
func (s *DeletePartialWeagingService) restoreNurseState(
	ctx context.Context,
	weaging *partialweagingdomain.PartialWeaging,
	farrowing *farrowingdomain.Farrowing,
	deletedBy uuid.UUID,
	now time.Time,
) error {
	hasLaterEvents, err := s.repository.HasLaterNurseEvents(ctx, weaging.FarrowingID, weaging.WeagingDate)
	if err != nil {
		return err
	}
	if hasLaterEvents {
		return partialweagingdomain.ErrHasLaterNurseEvents
	}

	previousNurseDate, err := s.repository.GetPreviousNurseDate(ctx, weaging.FarrowingID, weaging.WeagingDate)
	if err != nil {
		return err
	}

	return farrowing.RestoreNurseState(previousNurseDate, deletedBy, now)
}

// restoreSowState blocks a non-Nodriza delete when the sow has later mounts or removals.
// It returns the sow moved back to Lactando when the deletion can proceed.
func (s *DeletePartialWeagingService) restoreSowState(
	ctx context.Context,
	weaging *partialweagingdomain.PartialWeaging,
	deletedBy uuid.UUID,
	now time.Time,
) (*sowdomain.Sow, error) {
	hasFutureEvents, err := s.repository.HasFutureEvents(ctx, weaging.SowID, weaging.WeagingDate)
	if err != nil {
		return nil, err
	}
	if hasFutureEvents {
		return nil, partialweagingdomain.ErrHasFutureEvents
	}

	sow, err := s.repository.GetSow(ctx, weaging.SowID)
	if err != nil {
		return nil, err
	}
	if err := sow.ChangeState(sowdomain.StateLactating, deletedBy, now); err != nil {
		return nil, err
	}
	return sow, nil
}
