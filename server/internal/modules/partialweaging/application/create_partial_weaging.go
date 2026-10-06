package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	partialweagingdomain "server/internal/modules/partialweaging/domain"
	"server/internal/modules/partialweaging/ports"
	sowdomain "server/internal/modules/sow/domain"
)

var ErrSowNotLactating = errors.New("La cerda no está lactando")

// CreatePartialWeagingCommand carries the fields required to register a partial weaging.
// The farrowing is derived by the use case from the sow's latest farrowing.
type CreatePartialWeagingCommand struct {
	SowID       uuid.UUID
	WeagingDate time.Time
	Quantity    int
	TotalWeight *float64
	Type        partialweagingdomain.Type
	Note        *string
	CreatedBy   uuid.UUID
}

type CreatePartialWeagingService struct {
	repository ports.PartialWeagingRepository
	clock      func() time.Time
}

// NewCreatePartialWeagingService builds a create-partial-weaging use case with its repository and clock.
// It returns a service ready to execute CreatePartialWeagingCommand values.
func NewCreatePartialWeagingService(repository ports.PartialWeagingRepository, clock func() time.Time) *CreatePartialWeagingService {
	return &CreatePartialWeagingService{repository: repository, clock: clock}
}

// Execute registers a partial weaging and reduces the farrowing's current piglets.
// It validates that the sow is lactating, resolves its latest farrowing and the latest
// related event date, then for a Nodriza type flags the farrowing as a nurse while for
// the other types it moves the sow to weaned. Everything is persisted in one transaction.
func (s *CreatePartialWeagingService) Execute(ctx context.Context, command CreatePartialWeagingCommand) (*partialweagingdomain.PartialWeaging, error) {
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
	newWeaging, err := partialweagingdomain.NewPartialWeaging(partialweagingdomain.NewPartialWeagingParams{
		ID:          uuid.New(),
		FarrowingID: farrowing.ID,
		SowID:       sow.ID,
		WeagingDate: command.WeagingDate,
		Quantity:    command.Quantity,
		TotalWeight: command.TotalWeight,
		Type:        command.Type,
		Note:        command.Note,
		CreatedBy:   command.CreatedBy,
	}, farrowing.FarrowDate, lastEventDate, now)
	if err != nil {
		return nil, err
	}

	if err := farrowing.ReduceCurrentPiglets(newWeaging.Quantity, command.CreatedBy, now); err != nil {
		return nil, err
	}

	if newWeaging.Type == partialweagingdomain.TypeNodriza {
		if err := farrowing.MarkAsNurse(newWeaging.WeagingDate, command.CreatedBy, now); err != nil {
			return nil, err
		}
	} else {
		if err := sow.ChangeState(sowdomain.StateWeaned, command.CreatedBy, now); err != nil {
			return nil, err
		}
	}

	if err := s.repository.Create(ctx, newWeaging, farrowing, sow); err != nil {
		return nil, err
	}
	return newWeaging, nil
}
