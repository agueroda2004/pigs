package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	boarremovaldomain "server/internal/modules/boarremoval/domain"
	"server/internal/modules/boarremoval/ports"
)

var ErrBoarNotRemovable = errors.New("El verraco no está en un estado válido para la baja")

// CreateBoarRemovalCommand carries the fields required to register a removal.
// The last state and the reference dates are derived by the use case from the boar.
type CreateBoarRemovalCommand struct {
	BoarID      uuid.UUID
	RemovalDate time.Time
	Type        boarremovaldomain.Type
	Reason      boarremovaldomain.Reason
	Note        *string
	CreatedBy   uuid.UUID
}

type CreateBoarRemovalService struct {
	repository ports.BoarRemovalRepository
	clock      func() time.Time
}

// NewCreateBoarRemovalService builds a create-removal use case with its repository and clock.
// It returns a service ready to execute CreateBoarRemovalCommand values.
func NewCreateBoarRemovalService(repository ports.BoarRemovalRepository, clock func() time.Time) *CreateBoarRemovalService {
	return &CreateBoarRemovalService{repository: repository, clock: clock}
}

// Execute registers a boar removal and deactivates the boar.
// It validates the boar state and the removal dates, then persists the removal
// and the boar state with active false in one transaction.
func (s *CreateBoarRemovalService) Execute(ctx context.Context, command CreateBoarRemovalCommand) (*boarremovaldomain.BoarRemoval, error) {
	boar, err := s.repository.GetBoar(ctx, command.BoarID)
	if err != nil {
		return nil, err
	}
	if !isRemovableBoarState(boar.State) {
		return nil, ErrBoarNotRemovable
	}

	lastMountDate, err := s.repository.GetLastMountDate(ctx, command.BoarID)
	if err != nil {
		return nil, err
	}

	targetState, err := command.Type.BoarState()
	if err != nil {
		return nil, err
	}

	now := s.clock()
	newRemoval, err := boarremovaldomain.NewBoarRemoval(boarremovaldomain.NewBoarRemovalParams{
		ID:          uuid.New(),
		BoarID:      command.BoarID,
		RemovalDate: command.RemovalDate,
		Type:        command.Type,
		Reason:      command.Reason,
		Note:        command.Note,
		LastState:   string(boar.State),
		CreatedBy:   command.CreatedBy,
	}, boarremovaldomain.Reference{EntryDate: boar.EntryDate, LastMountDate: lastMountDate}, now)
	if err != nil {
		return nil, err
	}

	inactive := false
	if err := boar.Update(boardomain.UpdateBoarParams{Active: &inactive}, command.CreatedBy, now); err != nil {
		return nil, err
	}
	if err := boar.ChangeState(boardomain.State(targetState), command.CreatedBy, now); err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, newRemoval, boar); err != nil {
		return nil, err
	}
	return newRemoval, nil
}

// isRemovableBoarState reports whether a boar state allows registering a removal.
// Only alive boars may be removed.
func isRemovableBoarState(state boardomain.State) bool {
	return state == boardomain.StateAlive
}
