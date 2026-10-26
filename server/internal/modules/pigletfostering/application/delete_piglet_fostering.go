package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	"server/internal/modules/pigletfostering/ports"
)

// DeletePigletFosteringCommand identifies the actor that removes a fostering.
// The farrowings and the moved quantity are derived from the stored record.
type DeletePigletFosteringCommand struct {
	DeletedBy uuid.UUID
}

type DeletePigletFosteringService struct {
	repository ports.PigletFosteringRepository
	clock      func() time.Time
}

// NewDeletePigletFosteringService builds a delete-fostering use case with its
// repository and clock, ready to execute DeletePigletFosteringCommand values.
func NewDeletePigletFosteringService(repository ports.PigletFosteringRepository, clock func() time.Time) *DeletePigletFosteringService {
	return &DeletePigletFosteringService{repository: repository, clock: clock}
}

// Execute deletes a fostering and returns the moved piglets to the donor farrowing.
// It blocks the deletion when either farrowing has a full or partial weaging after
// the movement date, then takes the piglets back from the receiver, gives them to
// the donor and removes the record in one transaction.
func (s *DeletePigletFosteringService) Execute(
	ctx context.Context,
	id uuid.UUID,
	command DeletePigletFosteringCommand,
) error {
	fostering, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	donor, err := s.repository.GetFarrowingByID(ctx, fostering.DonorFarrowingID)
	if err != nil {
		return err
	}

	receiver, err := s.repository.GetFarrowingByID(ctx, fostering.ReceiverFarrowingID)
	if err != nil {
		return err
	}

	bounds, err := s.deleteBounds(ctx, donor.ID, receiver.ID)
	if err != nil {
		return err
	}

	if err := fostering.ValidateDelete(bounds, command.DeletedBy); err != nil {
		return err
	}

	now := s.clock()
	if err := receiver.ReduceCurrentPiglets(fostering.Quantity, command.DeletedBy, now); err != nil {
		return err
	}
	if err := donor.AddCurrentPiglets(fostering.Quantity, command.DeletedBy, now); err != nil {
		return err
	}

	return s.repository.Delete(ctx, fostering, donor, receiver)
}

// deleteBounds resolves the latest full and partial weaging dates of both
// farrowings. It always queries both sides so a later weaging always blocks.
func (s *DeletePigletFosteringService) deleteBounds(ctx context.Context, donorFarrowingID, receiverFarrowingID uuid.UUID) (pigletfosteringdomain.FosteringDeleteBounds, error) {
	donorWeagingDate, err := s.repository.GetLastWeagingDate(ctx, donorFarrowingID)
	if err != nil {
		return pigletfosteringdomain.FosteringDeleteBounds{}, err
	}

	donorPartialWeagingDate, err := s.repository.GetLastPartialWeagingDate(ctx, donorFarrowingID)
	if err != nil {
		return pigletfosteringdomain.FosteringDeleteBounds{}, err
	}

	receiverWeagingDate, err := s.repository.GetLastWeagingDate(ctx, receiverFarrowingID)
	if err != nil {
		return pigletfosteringdomain.FosteringDeleteBounds{}, err
	}

	receiverPartialWeagingDate, err := s.repository.GetLastPartialWeagingDate(ctx, receiverFarrowingID)
	if err != nil {
		return pigletfosteringdomain.FosteringDeleteBounds{}, err
	}

	return pigletfosteringdomain.FosteringDeleteBounds{
		DonorPartialWeagingDate:    donorPartialWeagingDate,
		DonorWeagingDate:           donorWeagingDate,
		ReceiverPartialWeagingDate: receiverPartialWeagingDate,
		ReceiverWeagingDate:        receiverWeagingDate,
	}, nil
}
