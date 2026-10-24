package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	"server/internal/modules/pigletdeath/ports"
)

// UpdatePigletDeathCommand carries the editable fields of a piglet death update.
// The actor is required and the operator name is resolved by the use case.
type UpdatePigletDeathCommand struct {
	Params    pigletdeathdomain.UpdatePigletDeathParams
	UpdatedBy uuid.UUID
}

type UpdatePigletDeathService struct {
	repository ports.PigletDeathRepository
	clock      func() time.Time
}

// NewUpdatePigletDeathService builds an update-piglet-death use case with its
// repository and clock, ready to execute UpdatePigletDeathCommand values.
func NewUpdatePigletDeathService(repository ports.PigletDeathRepository, clock func() time.Time) *UpdatePigletDeathService {
	return &UpdatePigletDeathService{repository: repository, clock: clock}
}

// Execute applies the provided fields to an existing piglet death.
// It loads the death, resolves the new operator when one is provided, validates
// the change and persists the updated aggregate.
func (s *UpdatePigletDeathService) Execute(
	ctx context.Context,
	id uuid.UUID,
	command UpdatePigletDeathCommand,
) (*pigletdeathdomain.PigletDeath, error) {
	death, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if command.Params.OperatorID != nil {
		operator, err := s.repository.GetOperator(ctx, *command.Params.OperatorID)
		if err != nil {
			return nil, err
		}
		command.Params.OperatorName = &operator.Name
	}

	if err := death.Update(command.Params, command.UpdatedBy, s.clock()); err != nil {
		return nil, err
	}

	if err := s.repository.Update(ctx, death); err != nil {
		return nil, err
	}
	return death, nil
}
