package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	"server/internal/modules/boarremoval/ports"
)

var ErrBoarStateMismatch = errors.New("El verraco no está en el estado esperado para eliminar la baja")

// DeleteBoarRemovalCommand identifies the removal to delete and its actor.
// The removal and boar states are derived by the use case.
type DeleteBoarRemovalCommand struct {
	ID        uuid.UUID
	DeletedBy uuid.UUID
}

type DeleteBoarRemovalService struct {
	repository ports.BoarRemovalRepository
	clock      func() time.Time
}

// NewDeleteBoarRemovalService builds a delete-removal use case with its
// repository and clock, ready to execute DeleteBoarRemovalCommand values.
func NewDeleteBoarRemovalService(repository ports.BoarRemovalRepository, clock func() time.Time) *DeleteBoarRemovalService {
	return &DeleteBoarRemovalService{repository: repository, clock: clock}
}

// Execute deletes a removal and undoes its effects on the boar.
// It reactivates the boar, restores its last state and persists both changes in
// one transaction, without touching any service.
func (s *DeleteBoarRemovalService) Execute(ctx context.Context, command DeleteBoarRemovalCommand) error {
	removal, err := s.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	boar, err := s.repository.GetBoar(ctx, removal.BoarID)
	if err != nil {
		return err
	}

	targetState, err := removal.Type.BoarState()
	if err != nil {
		return err
	}
	if boar.Active || string(boar.State) != targetState {
		return ErrBoarStateMismatch
	}

	lastState, err := boardomain.ParseState(removal.LastState)
	if err != nil {
		return err
	}

	now := s.clock()
	active := true
	if err := boar.Update(boardomain.UpdateBoarParams{Active: &active}, command.DeletedBy, now); err != nil {
		return err
	}
	if err := boar.ChangeState(lastState, command.DeletedBy, now); err != nil {
		return err
	}

	return s.repository.Delete(ctx, removal, boar)
}
