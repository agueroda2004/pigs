package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	"server/internal/modules/boar/ports"
)

type ChangeBoarStateCommand struct {
	State     boardomain.State
	UpdatedBy uuid.UUID
}

type ChangeBoarStateService struct {
	repository ports.BoarRepository
	clock      func() time.Time
}

// NewChangeBoarStateService builds a change-state use case with its repository and clock.
// It returns a service reserved for domain event handlers, not for the HTTP update flow.
func NewChangeBoarStateService(repository ports.BoarRepository, clock func() time.Time) *ChangeBoarStateService {
	return &ChangeBoarStateService{repository: repository, clock: clock}
}

// Execute applies an event-driven state change to an existing boar.
// It loads the boar, delegates validation to the domain and persists the new state.
func (s *ChangeBoarStateService) Execute(
	ctx context.Context,
	boarID uuid.UUID,
	command ChangeBoarStateCommand,
) (*boardomain.Boar, error) {
	currentBoar, err := s.repository.GetByID(ctx, boarID)
	if err != nil {
		return nil, err
	}

	if err := currentBoar.ChangeState(command.State, command.UpdatedBy, s.clock()); err != nil {
		return nil, err
	}

	if err := s.repository.UpdateState(ctx, currentBoar); err != nil {
		return nil, err
	}
	return currentBoar, nil
}
