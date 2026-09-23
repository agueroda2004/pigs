package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	"server/internal/modules/boar/ports"
)

type UpdateBoarCommand struct {
	Code           *string
	Location       *string
	Active         *bool
	EntryDate      *time.Time
	BirthDate      *time.Time
	ClearBirthDate bool
	Note           *string
	Origin         *boardomain.Origin
	BreedID        *uuid.UUID
	UpdatedBy      uuid.UUID
}

type UpdateBoarService struct {
	repository ports.BoarRepository
	clock      func() time.Time
}

// NewUpdateBoarService builds an update-boar use case with its repository and clock.
// It returns a service ready to execute UpdateBoarCommand values.
func NewUpdateBoarService(repository ports.BoarRepository, clock func() time.Time) *UpdateBoarService {
	return &UpdateBoarService{repository: repository, clock: clock}
}

// Execute applies the provided fields to an existing boar by its identifier.
// It loads the boar, validates the change and persists it without touching state.
func (s *UpdateBoarService) Execute(
	ctx context.Context,
	boarID uuid.UUID,
	command UpdateBoarCommand,
) (*boardomain.Boar, error) {
	currentBoar, err := s.repository.GetByID(ctx, boarID)
	if err != nil {
		return nil, err
	}

	if err := currentBoar.Update(boardomain.UpdateBoarParams{
		Code:           command.Code,
		Location:       command.Location,
		Active:         command.Active,
		EntryDate:      command.EntryDate,
		BirthDate:      command.BirthDate,
		ClearBirthDate: command.ClearBirthDate,
		Note:           command.Note,
		Origin:         command.Origin,
		BreedID:        command.BreedID,
	}, command.UpdatedBy, s.clock()); err != nil {
		return nil, err
	}

	if err := s.repository.Update(ctx, currentBoar); err != nil {
		return nil, err
	}
	return currentBoar, nil
}
