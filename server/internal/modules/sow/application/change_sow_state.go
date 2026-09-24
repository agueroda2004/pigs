package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sow/ports"
)

type ChangeSowStateCommand struct {
	State     sowdomain.State
	UpdatedBy uuid.UUID
}

type ChangeSowStateService struct {
	repository ports.SowRepository
	clock      func() time.Time
}

// NewChangeSowStateService builds a change-state use case with its repository and clock.
// It returns a service reserved for domain event handlers, not for the HTTP update flow.
func NewChangeSowStateService(repository ports.SowRepository, clock func() time.Time) *ChangeSowStateService {
	return &ChangeSowStateService{repository: repository, clock: clock}
}

// Execute applies an event-driven state change to an existing sow.
// It loads the sow, delegates validation to the domain and persists the new state.
func (s *ChangeSowStateService) Execute(
	ctx context.Context,
	sowID uuid.UUID,
	command ChangeSowStateCommand,
) (*sowdomain.Sow, error) {
	currentSow, err := s.repository.GetByID(ctx, sowID)
	if err != nil {
		return nil, err
	}

	if err := currentSow.ChangeState(command.State, command.UpdatedBy, s.clock()); err != nil {
		return nil, err
	}

	if err := s.repository.UpdateState(ctx, currentSow); err != nil {
		return nil, err
	}
	return currentSow, nil
}
