package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	"server/internal/modules/pigletdeath/ports"
)

// DeletePigletDeathCommand identifies the piglet death to delete and its actor.
// The farrowing and the restored quantity are derived from the stored death.
type DeletePigletDeathCommand struct {
	ID        uuid.UUID
	DeletedBy uuid.UUID
}

type DeletePigletDeathService struct {
	repository ports.PigletDeathRepository
	clock      func() time.Time
}

// NewDeletePigletDeathService builds a delete-piglet-death use case with its
// repository and clock, ready to execute DeletePigletDeathCommand values.
func NewDeletePigletDeathService(repository ports.PigletDeathRepository, clock func() time.Time) *DeletePigletDeathService {
	return &DeletePigletDeathService{repository: repository, clock: clock}
}

// Execute deletes a piglet death and gives its quantity back to the farrowing balance.
// It blocks the deletion when a full or partial weaging was registered after the
// death date, then restores the piglets and removes the record in one transaction.
func (s *DeletePigletDeathService) Execute(ctx context.Context, command DeletePigletDeathCommand) error {
	death, err := s.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	farrowing, err := s.repository.GetFarrowingByID(ctx, death.FarrowingID)
	if err != nil {
		return err
	}

	hasWeaging, err := s.repository.HasWeagingAfter(ctx, death.FarrowingID, death.DeathDate)
	if err != nil {
		return err
	}
	if hasWeaging {
		return pigletdeathdomain.ErrHasLaterWeaging
	}

	hasPartialWeaging, err := s.repository.HasPartialWeagingAfter(ctx, death.FarrowingID, death.DeathDate)
	if err != nil {
		return err
	}
	if hasPartialWeaging {
		return pigletdeathdomain.ErrHasLaterPartialWeaging
	}

	now := s.clock()
	if err := farrowing.AddCurrentPiglets(death.Quantity, command.DeletedBy, now); err != nil {
		return err
	}

	return s.repository.Delete(ctx, death, farrowing)
}
