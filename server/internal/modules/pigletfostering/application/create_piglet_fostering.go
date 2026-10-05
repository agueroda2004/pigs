package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	"server/internal/modules/pigletfostering/ports"
	sowdomain "server/internal/modules/sow/domain"
)

var ErrSowNotLactating = errors.New("La cerda no está lactando")

// CreatePigletFosteringCommand carries the fields required to register a fostering.
// The farrowings are derived by the use case from each sow's latest farrowing.
type CreatePigletFosteringCommand struct {
	DonorSowID    uuid.UUID
	ReceiverSowID uuid.UUID
	MovementDate  time.Time
	Quantity      int
	Note          *string
	CreatedBy     uuid.UUID
}

type CreatePigletFosteringService struct {
	repository ports.PigletFosteringRepository
	clock      func() time.Time
}

// NewCreatePigletFosteringService builds a create-fostering use case with its repository and clock.
// It returns a service ready to execute CreatePigletFosteringCommand values.
func NewCreatePigletFosteringService(repository ports.PigletFosteringRepository, clock func() time.Time) *CreatePigletFosteringService {
	return &CreatePigletFosteringService{repository: repository, clock: clock}
}

// Execute registers a fostering and moves the piglets between both balances.
// It validates that both sows are lactating, resolves each latest farrowing,
// then decreases the donor and increases the receiver in one transaction.
func (s *CreatePigletFosteringService) Execute(ctx context.Context, command CreatePigletFosteringCommand) (*pigletfosteringdomain.PigletFostering, error) {
	donorSow, err := s.repository.GetSow(ctx, command.DonorSowID)
	if err != nil {
		return nil, err
	}
	if donorSow.State != sowdomain.StateLactating {
		return nil, ErrSowNotLactating
	}

	receiverSow, err := s.repository.GetSow(ctx, command.ReceiverSowID)
	if err != nil {
		return nil, err
	}
	if receiverSow.State != sowdomain.StateLactating {
		return nil, ErrSowNotLactating
	}

	donor, err := s.repository.GetLastFarrowing(ctx, command.DonorSowID)
	if err != nil {
		return nil, err
	}

	receiver, err := s.repository.GetLastFarrowing(ctx, command.ReceiverSowID)
	if err != nil {
		return nil, err
	}

	now := s.clock()
	newFostering, err := pigletfosteringdomain.NewPigletFostering(pigletfosteringdomain.NewPigletFosteringParams{
		ID:                  uuid.New(),
		DonorFarrowingID:    donor.ID,
		ReceiverFarrowingID: receiver.ID,
		DonorSowID:          donorSow.ID,
		ReceiverSowID:       receiverSow.ID,
		MovementDate:        command.MovementDate,
		Quantity:            command.Quantity,
		Note:                command.Note,
		CreatedBy:           command.CreatedBy,
	}, donor.FarrowDate, receiver.FarrowDate, now)
	if err != nil {
		return nil, err
	}

	if err := donor.ReduceCurrentPiglets(newFostering.Quantity, command.CreatedBy, now); err != nil {
		return nil, err
	}
	if err := receiver.AddCurrentPiglets(newFostering.Quantity, command.CreatedBy, now); err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, newFostering, donor, receiver); err != nil {
		return nil, err
	}
	return newFostering, nil
}
