package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sow/ports"
)

type UpdateSowCommand struct {
	Code           *string
	Location       *string
	EntryDate      *time.Time
	BirthDate      *time.Time
	ClearBirthDate bool
	Note           *string
	Origin         *sowdomain.Origin
	BreedID        *uuid.UUID
	Parity         *int
	UpdatedBy      uuid.UUID
}

type UpdateSowService struct {
	repository ports.SowRepository
	clock      func() time.Time
}

// NewUpdateSowService builds an update-sow use case with its repository and clock.
// It returns a service ready to execute UpdateSowCommand values.
func NewUpdateSowService(repository ports.SowRepository, clock func() time.Time) *UpdateSowService {
	return &UpdateSowService{repository: repository, clock: clock}
}

// Execute applies the provided fields to an existing sow by its identifier.
// It loads the sow, resolves the most recent service date when a date changes so
// the entry date can be validated, and persists the change without touching state.
func (s *UpdateSowService) Execute(
	ctx context.Context,
	sowID uuid.UUID,
	command UpdateSowCommand,
) (*sowdomain.Sow, error) {
	currentSow, err := s.repository.GetByID(ctx, sowID)
	if err != nil {
		return nil, err
	}

	params := sowdomain.UpdateSowParams{
		Code:           command.Code,
		Location:       command.Location,
		EntryDate:      command.EntryDate,
		BirthDate:      command.BirthDate,
		ClearBirthDate: command.ClearBirthDate,
		Note:           command.Note,
		Origin:         command.Origin,
		BreedID:        command.BreedID,
		Parity:         command.Parity,
	}

	if command.EntryDate != nil || command.BirthDate != nil || command.ClearBirthDate {
		lastServiceDate, err := s.repository.LastServiceDate(ctx, sowID)
		if err != nil {
			return nil, err
		}
		params.LastServiceDate = lastServiceDate
	}

	if err := currentSow.Update(params, command.UpdatedBy, s.clock()); err != nil {
		return nil, err
	}

	if err := s.repository.Update(ctx, currentSow); err != nil {
		return nil, err
	}
	return currentSow, nil
}
