package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	"server/internal/modules/pigletfostering/ports"
)

// UpdatePigletFosteringCommand carries the editable fields of a fostering update.
// Only the note can change after creation and the actor is required.
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

// Execute applies the note to an existing fostering.
// It loads the fostering, validates the note and persists the change; the date,
// quantity and both farrowings are never modified.
func (s *UpdatePigletFosteringService) Execute(
	ctx context.Context,
	id uuid.UUID,
	command UpdatePigletFosteringCommand,
) (*pigletfosteringdomain.PigletFostering, error) {
	fostering, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := fostering.Update(command.Params, command.UpdatedBy, s.clock()); err != nil {
		return nil, err
	}

	if err := s.repository.Update(ctx, fostering); err != nil {
		return nil, err
	}
	return fostering, nil
}
