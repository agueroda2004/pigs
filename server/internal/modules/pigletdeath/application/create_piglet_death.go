package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	"server/internal/modules/pigletdeath/ports"
	sowdomain "server/internal/modules/sow/domain"
)

var ErrSowNotLactating = errors.New("La cerda no está lactando")

// CreatePigletDeathCommand carries the fields required to register a piglet death.
// The farrowing is derived by the use case from the sow's latest farrowing.
type CreatePigletDeathCommand struct {
	SowID      uuid.UUID
	OperatorID uuid.UUID
	DeathDate  time.Time
	Quantity   int
	Weight     *float64
	Cause      pigletdeathdomain.PigletDeathCause
	Turn       pigletdeathdomain.Turn
	Note       *string
	CreatedBy  uuid.UUID
}

type CreatePigletDeathService struct {
	repository ports.PigletDeathRepository
	clock      func() time.Time
}

// NewCreatePigletDeathService builds a create-piglet-death use case with its repository and clock.
// It returns a service ready to execute CreatePigletDeathCommand values.
func NewCreatePigletDeathService(repository ports.PigletDeathRepository, clock func() time.Time) *CreatePigletDeathService {
	return &CreatePigletDeathService{repository: repository, clock: clock}
}

// Execute registers a piglet death and reduces the farrowing's current piglets.
// It validates that the sow is lactating, resolves its latest farrowing and the
// operator, then persists the death and the reduced balance in one transaction.
func (s *CreatePigletDeathService) Execute(ctx context.Context, command CreatePigletDeathCommand) (*pigletdeathdomain.PigletDeath, error) {
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

	operator, err := s.repository.GetOperator(ctx, command.OperatorID)
	if err != nil {
		return nil, err
	}

	now := s.clock()
	newDeath, err := pigletdeathdomain.NewPigletDeath(pigletdeathdomain.NewPigletDeathParams{
		ID:           uuid.New(),
		FarrowingID:  farrowing.ID,
		SowID:        command.SowID,
		OperatorID:   command.OperatorID,
		OperatorName: operator.Name,
		DeathDate:    command.DeathDate,
		Quantity:     command.Quantity,
		Weight:       command.Weight,
		Cause:        command.Cause,
		Turn:         command.Turn,
		Note:         command.Note,
		CreatedBy:    command.CreatedBy,
	}, farrowing.FarrowDate, now)
	if err != nil {
		return nil, err
	}

	if err := farrowing.ReduceCurrentPiglets(newDeath.Quantity, command.CreatedBy, now); err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, newDeath, farrowing); err != nil {
		return nil, err
	}
	return newDeath, nil
}
