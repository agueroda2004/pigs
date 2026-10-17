package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	sowdomain "server/internal/modules/sow/domain"
	sowremovaldomain "server/internal/modules/sowremoval/domain"
	"server/internal/modules/sowremoval/ports"
)

// UpdateSowRemovalCommand carries the mutable fields of a removal update.
// A nil field is left unchanged; sow_id is intentionally absent because a
// removal can never change the sow it belongs to.
type UpdateSowRemovalCommand struct {
	RemovalDate *time.Time
	Type        *sowremovaldomain.Type
	Reason      *sowremovaldomain.Reason
	Note        *string
	UpdatedBy   uuid.UUID
}

type UpdateSowRemovalService struct {
	repository ports.SowRemovalRepository
	clock      func() time.Time
}

// NewUpdateSowRemovalService builds an update-removal use case with its
// repository and clock, ready to execute UpdateSowRemovalCommand values.
func NewUpdateSowRemovalService(repository ports.SowRemovalRepository, clock func() time.Time) *UpdateSowRemovalService {
	return &UpdateSowRemovalService{repository: repository, clock: clock}
}

// Execute applies the provided fields to an existing removal by its identifier.
// It reloads the date references and, when the type changes, the resulting sow
// state, then persists the removal and the sow in one transaction.
func (s *UpdateSowRemovalService) Execute(ctx context.Context, removalID uuid.UUID, command UpdateSowRemovalCommand) (*sowremovaldomain.SowRemoval, error) {
	removal, err := s.repository.GetByID(ctx, removalID)
	if err != nil {
		return nil, err
	}

	var sow *sowdomain.Sow
	if command.RemovalDate != nil || command.Type != nil {
		sow, err = s.repository.GetSow(ctx, removal.SowID)
		if err != nil {
			return nil, err
		}
	}

	reference := sowremovaldomain.Reference{}
	if command.RemovalDate != nil {
		reference, _, err = removalReference(ctx, s.repository, sow, sowdomain.State(removal.LastState))
		if err != nil {
			return nil, err
		}
	}

	now := s.clock()
	if err := removal.Update(sowremovaldomain.UpdateSowRemovalParams{
		RemovalDate: command.RemovalDate,
		Type:        command.Type,
		Reason:      command.Reason,
		Note:        command.Note,
	}, reference, command.UpdatedBy, now); err != nil {
		return nil, err
	}

	var updatedSow *sowdomain.Sow
	if command.Type != nil {
		targetState, err := command.Type.SowState()
		if err != nil {
			return nil, err
		}
		if sow.State != sowdomain.State(targetState) {
			if err := sow.ChangeState(sowdomain.State(targetState), command.UpdatedBy, now); err != nil {
				return nil, err
			}
			updatedSow = sow
		}
	}

	if err := s.repository.Update(ctx, removal, updatedSow); err != nil {
		return nil, err
	}
	return removal, nil
}
