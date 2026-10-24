package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	"server/internal/modules/pigletfostering/ports"
)

// UpdatePigletFosteringCommand carries the editable fields of a fostering update.
// The donor and receiver farrowings are fixed, so only the date, quantity and note
// can change and the actor is required.
type UpdatePigletFosteringCommand struct {
	Params    pigletfosteringdomain.UpdatePigletFosteringParams
	UpdatedBy uuid.UUID
}

type UpdatePigletFosteringService struct {
	repository ports.PigletFosteringRepository
	clock      func() time.Time
}

// NewUpdatePigletFosteringService builds an update-fostering use case with its
// repository and clock, ready to execute UpdatePigletFosteringCommand values.
func NewUpdatePigletFosteringService(repository ports.PigletFosteringRepository, clock func() time.Time) *UpdatePigletFosteringService {
	return &UpdatePigletFosteringService{repository: repository, clock: clock}
}

// Execute applies the provided fields to an existing fostering.
// It loads the fostering and both stored farrowings, resolves each nurse partial
// weaging, revalidates the change and rebalances the piglets before persisting.
func (s *UpdatePigletFosteringService) Execute(
	ctx context.Context,
	id uuid.UUID,
	command UpdatePigletFosteringCommand,
) (*pigletfosteringdomain.PigletFostering, error) {
	fostering, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	donor, err := s.repository.GetFarrowingByID(ctx, fostering.DonorFarrowingID)
	if err != nil {
		return nil, err
	}

	receiver, err := s.repository.GetFarrowingByID(ctx, fostering.ReceiverFarrowingID)
	if err != nil {
		return nil, err
	}

	donorPartialWeagingDate, err := s.lastPartialWeagingDate(ctx, donor)
	if err != nil {
		return nil, err
	}

	receiverPartialWeagingDate, err := s.lastPartialWeagingDate(ctx, receiver)
	if err != nil {
		return nil, err
	}

	previousQuantity := fostering.Quantity
	now := s.clock()
	if err := fostering.Update(command.Params, pigletfosteringdomain.FosteringDateBounds{
		DonorFarrowDate:            donor.FarrowDate,
		ReceiverFarrowDate:         receiver.FarrowDate,
		DonorPartialWeagingDate:    donorPartialWeagingDate,
		ReceiverPartialWeagingDate: receiverPartialWeagingDate,
	}, command.UpdatedBy, now); err != nil {
		return nil, err
	}

	if err := rebalancePiglets(donor, receiver, fostering.Quantity-previousQuantity, command.UpdatedBy, now); err != nil {
		return nil, err
	}

	if err := s.repository.Update(ctx, fostering); err != nil {
		return nil, err
	}
	return fostering, nil
}

// rebalancePiglets moves the difference between the previous and the new quantity
// between both farrowings. A positive delta subtracts from the donor and a
// negative delta gives the piglets back to it.
func rebalancePiglets(
	donor *farrowingdomain.Farrowing,
	receiver *farrowingdomain.Farrowing,
	delta int,
	updatedBy uuid.UUID,
	now time.Time,
) error {
	switch {
	case delta > 0:
		if err := donor.ReduceCurrentPiglets(delta, updatedBy, now); err != nil {
			return err
		}
		return receiver.AddCurrentPiglets(delta, updatedBy, now)
	case delta < 0:
		if err := receiver.ReduceCurrentPiglets(-delta, updatedBy, now); err != nil {
			return err
		}
		return donor.AddCurrentPiglets(-delta, updatedBy, now)
	default:
		return nil
	}
}

// lastPartialWeagingDate resolves the latest partial weaging of a farrowing when
// it is a nurse, and returns the zero time for non-nurse farrowings.
func (s *UpdatePigletFosteringService) lastPartialWeagingDate(ctx context.Context, farrowing *farrowingdomain.Farrowing) (time.Time, error) {
	if !farrowing.IsNurse {
		return time.Time{}, nil
	}
	return s.repository.GetLastPartialWeagingDate(ctx, farrowing.ID)
}
