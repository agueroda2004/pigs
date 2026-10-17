package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
	sowremovaldomain "server/internal/modules/sowremoval/domain"
	"server/internal/modules/sowremoval/ports"
)

var ErrSowNotRemovable = errors.New("La cerda no está en un estado válido para la baja")

// CreateSowRemovalCommand carries the fields required to register a removal.
// The last state and the reference dates are derived by the use case from the sow.
type CreateSowRemovalCommand struct {
	SowID       uuid.UUID
	RemovalDate time.Time
	Type        sowremovaldomain.Type
	Reason      sowremovaldomain.Reason
	Note        *string
	CreatedBy   uuid.UUID
}

type CreateSowRemovalService struct {
	repository ports.SowRemovalRepository
	clock      func() time.Time
}

// NewCreateSowRemovalService builds a create-removal use case with its repository and clock.
// It returns a service ready to execute CreateSowRemovalCommand values.
func NewCreateSowRemovalService(repository ports.SowRemovalRepository, clock func() time.Time) *CreateSowRemovalService {
	return &CreateSowRemovalService{repository: repository, clock: clock}
}

// Execute registers a sow removal, deactivates the sow and fails its active service.
// It validates the sow state and the removal dates, then persists the removal, the
// sow state with active false and the failed service in one transaction.
func (s *CreateSowRemovalService) Execute(ctx context.Context, command CreateSowRemovalCommand) (*sowremovaldomain.SowRemoval, error) {
	sow, err := s.repository.GetSow(ctx, command.SowID)
	if err != nil {
		return nil, err
	}
	if !isRemovableSowState(sow.State) {
		return nil, ErrSowNotRemovable
	}

	reference, service, err := removalReference(ctx, s.repository, sow, sow.State)
	if err != nil {
		return nil, err
	}

	targetState, err := command.Type.SowState()
	if err != nil {
		return nil, err
	}

	now := s.clock()
	newRemoval, err := sowremovaldomain.NewSowRemoval(sowremovaldomain.NewSowRemovalParams{
		ID:          uuid.New(),
		SowID:       command.SowID,
		RemovalDate: command.RemovalDate,
		Type:        command.Type,
		Reason:      command.Reason,
		Note:        command.Note,
		LastState:   string(sow.State),
		CreatedBy:   command.CreatedBy,
	}, reference, now)
	if err != nil {
		return nil, err
	}

	if err := sow.SetActive(false, command.CreatedBy, now); err != nil {
		return nil, err
	}
	if err := sow.ChangeState(sowdomain.State(targetState), command.CreatedBy, now); err != nil {
		return nil, err
	}

	var activeService *servicedomain.Service
	if service != nil && service.State == servicedomain.StateConfirmed {
		if err := service.ChangeState(servicedomain.StateFailed, command.CreatedBy, now); err != nil {
			return nil, err
		}
		activeService = service
	}

	if err := s.repository.Create(ctx, newRemoval, sow, activeService); err != nil {
		return nil, err
	}
	return newRemoval, nil
}

// isRemovableSowState reports whether a sow state allows registering a removal.
// Only alive, weaned, aborted and gestating sows may be removed.
func isRemovableSowState(state sowdomain.State) bool {
	switch state {
	case sowdomain.StateAlive, sowdomain.StateWeaned, sowdomain.StateAborted, sowdomain.StatePregnant:
		return true
	default:
		return false
	}
}
