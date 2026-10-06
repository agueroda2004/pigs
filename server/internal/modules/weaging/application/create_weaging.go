package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	sowdomain "server/internal/modules/sow/domain"
	weagingdomain "server/internal/modules/weaging/domain"
	"server/internal/modules/weaging/ports"
)

var ErrSowNotLactating = errors.New("La cerda no está lactando")

// CreateWeagingCommand carries the fields required to register a weaging.
// The farrowing is derived by the use case from the sow's latest farrowing.
type CreateWeagingCommand struct {
	SowID       uuid.UUID
	WeagingDate time.Time
	Quantity    int
	TotalWeight *float64
	Destination *string
	Note        *string
	CreatedBy   uuid.UUID
}

type CreateWeagingService struct {
	repository ports.WeagingRepository
	clock      func() time.Time
}

// NewCreateWeagingService builds a create-weaging use case with its repository and clock.
// It returns a service ready to execute CreateWeagingCommand values.
func NewCreateWeagingService(repository ports.WeagingRepository, clock func() time.Time) *CreateWeagingService {
	return &CreateWeagingService{repository: repository, clock: clock}
}

// Execute registers a weaging, zeroes the farrowing piglets and weans the sow.
// It validates that the sow is lactating and the weaging date follows every related
// event, then closes the farrowing balance and moves the sow to weaned in one transaction.
func (s *CreateWeagingService) Execute(ctx context.Context, command CreateWeagingCommand) (*weagingdomain.Weaging, error) {
	sow, err := s.repository.GetSow(ctx, command.SowID)
	if err != nil {
		return nil, err
	}
	if sow.State != sowdomain.StateLactating {
		return nil, ErrSowNotLactating
	}

	farrowing, err := s.repository.GetLastFarrowing(ctx, command.SowID)
	if err != nil {
		return nil, err
	}

	lastEventDate, err := s.repository.GetLatestEventDate(ctx, farrowing.ID)
	if err != nil {
		return nil, err
	}

	now := s.clock()
	newWeaging, err := weagingdomain.NewWeaging(weagingdomain.NewWeagingParams{
		ID:          uuid.New(),
		FarrowingID: farrowing.ID,
		SowID:       sow.ID,
		WeagingDate: command.WeagingDate,
		Quantity:    command.Quantity,
		TotalWeight: command.TotalWeight,
		Destination: command.Destination,
		Note:        command.Note,
		CreatedBy:   command.CreatedBy,
	}, farrowing.FarrowDate, lastEventDate, now)
	if err != nil {
		return nil, err
	}

	if err := farrowing.WeanAll(newWeaging.Quantity, command.CreatedBy, now); err != nil {
		return nil, err
	}

	if err := sow.ChangeState(sowdomain.StateWeaned, command.CreatedBy, now); err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, newWeaging, farrowing, sow); err != nil {
		return nil, err
	}
	return newWeaging, nil
}
