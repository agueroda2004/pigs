package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sowremoval/ports"
)

var ErrSowStateMismatch = errors.New("La cerda no está en el estado esperado para eliminar la baja")

// DeleteSowRemovalCommand identifies the removal to delete and its actor.
// The removal, sow and service states are derived by the use case.
type DeleteSowRemovalCommand struct {
	ID        uuid.UUID
	DeletedBy uuid.UUID
}

type DeleteSowRemovalService struct {
	repository ports.SowRemovalRepository
	clock      func() time.Time
}

// NewDeleteSowRemovalService builds a delete-removal use case with its
// repository and clock, ready to execute DeleteSowRemovalCommand values.
func NewDeleteSowRemovalService(repository ports.SowRemovalRepository, clock func() time.Time) *DeleteSowRemovalService {
	return &DeleteSowRemovalService{repository: repository, clock: clock}
}

// Execute deletes a removal and undoes its effects on the sow and its service.
// It reactivates the sow, restores its last state, reverts a failed service to
// confirmed and persists all changes in one transaction.
func (s *DeleteSowRemovalService) Execute(ctx context.Context, command DeleteSowRemovalCommand) error {
	removal, err := s.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	sow, err := s.repository.GetSow(ctx, removal.SowID)
	if err != nil {
		return err
	}

	targetState, err := removal.Type.SowState()
	if err != nil {
		return err
	}
	if sow.Active || string(sow.State) != targetState {
		return ErrSowStateMismatch
	}

	lastState, err := sowdomain.ParseState(removal.LastState)
	if err != nil {
		return err
	}

	service, err := s.repository.GetLastService(ctx, removal.SowID)
	if err != nil && !errors.Is(err, ports.ErrServiceNotFound) {
		return err
	}
	if errors.Is(err, ports.ErrServiceNotFound) {
		service = nil
	}

	now := s.clock()
	if err := sow.SetActive(true, command.DeletedBy, now); err != nil {
		return err
	}
	if err := sow.ChangeState(lastState, command.DeletedBy, now); err != nil {
		return err
	}

	var restoredService *servicedomain.Service
	if service != nil && service.State == servicedomain.StateFailed {
		if err := service.ChangeState(servicedomain.StateConfirmed, command.DeletedBy, now); err != nil {
			return err
		}
		restoredService = service
	}

	return s.repository.Delete(ctx, removal, sow, restoredService)
}
