package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	boarremovaldomain "server/internal/modules/boarremoval/domain"
	"server/internal/modules/boarremoval/ports"
)

// UpdateBoarRemovalCommand carries the mutable fields of a removal update.
// A nil field is left unchanged; boar_id is intentionally absent because a
// removal can never change the boar it belongs to.
type UpdateBoarRemovalCommand struct {
	RemovalDate *time.Time
	Type        *boarremovaldomain.Type
	Reason      *boarremovaldomain.Reason
	Note        *string
	UpdatedBy   uuid.UUID
}

type UpdateBoarRemovalService struct {
	repository ports.BoarRemovalRepository
	clock      func() time.Time
}

// NewUpdateBoarRemovalService builds an update-removal use case with its
// repository and clock, ready to execute UpdateBoarRemovalCommand values.
func NewUpdateBoarRemovalService(repository ports.BoarRemovalRepository, clock func() time.Time) *UpdateBoarRemovalService {
	return &UpdateBoarRemovalService{repository: repository, clock: clock}
}

// Execute applies the provided fields to an existing removal by its identifier.
// It reloads the date references and, when the type changes, the resulting boar
// state, then persists the removal and the boar in one transaction.
func (s *UpdateBoarRemovalService) Execute(ctx context.Context, removalID uuid.UUID, command UpdateBoarRemovalCommand) (*boarremovaldomain.BoarRemoval, error) {
	removal, err := s.repository.GetByID(ctx, removalID)
	if err != nil {
		return nil, err
	}

	var boar *boardomain.Boar
	if command.RemovalDate != nil || command.Type != nil {
		boar, err = s.repository.GetBoar(ctx, removal.BoarID)
		if err != nil {
			return nil, err
		}
	}

	reference := boarremovaldomain.Reference{}
	if boar != nil {
		reference.EntryDate = boar.EntryDate
	}
	if command.RemovalDate != nil {
		lastMountDate, err := s.repository.GetLastMountDate(ctx, removal.BoarID)
		if err != nil {
			return nil, err
		}
		reference.LastMountDate = lastMountDate
	}

	now := s.clock()
	if err := removal.Update(boarremovaldomain.UpdateBoarRemovalParams{
		RemovalDate: command.RemovalDate,
		Type:        command.Type,
		Reason:      command.Reason,
		Note:        command.Note,
	}, reference, command.UpdatedBy, now); err != nil {
		return nil, err
	}

	var updatedBoar *boardomain.Boar
	if command.Type != nil {
		targetState, err := command.Type.BoarState()
		if err != nil {
			return nil, err
		}
		if boar.State != boardomain.State(targetState) {
			if err := boar.ChangeState(boardomain.State(targetState), command.UpdatedBy, now); err != nil {
				return nil, err
			}
			updatedBoar = boar
		}
	}

	if err := s.repository.Update(ctx, removal, updatedBoar); err != nil {
		return nil, err
	}
	return removal, nil
}
