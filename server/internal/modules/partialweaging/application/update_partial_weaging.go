package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	partialweagingdomain "server/internal/modules/partialweaging/domain"
	"server/internal/modules/partialweaging/ports"
)

// UpdatePartialWeagingCommand carries the editable fields of a partial weaging update.
// A nil field is left unchanged and ClearTotalWeight removes the weight; the sow,
// farrowing, date, quantity and type are never editable.
type UpdatePartialWeagingCommand struct {
	TotalWeight      *float64
	ClearTotalWeight bool
	Note             *string
	UpdatedBy        uuid.UUID
}

type UpdatePartialWeagingService struct {
	repository ports.PartialWeagingRepository
	clock      func() time.Time
}

// NewUpdatePartialWeagingService builds an update-partial-weaging use case with its
// repository and clock, ready to execute UpdatePartialWeagingCommand values.
func NewUpdatePartialWeagingService(repository ports.PartialWeagingRepository, clock func() time.Time) *UpdatePartialWeagingService {
	return &UpdatePartialWeagingService{repository: repository, clock: clock}
}

// Execute applies the editable fields to a partial weaging.
// It validates the update and persists the weight, note and audit fields only.
func (s *UpdatePartialWeagingService) Execute(ctx context.Context, id uuid.UUID, command UpdatePartialWeagingCommand) error {
	weaging, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := s.clock()
	if err := weaging.Update(partialweagingdomain.UpdatePartialWeagingParams{
		TotalWeight:      command.TotalWeight,
		ClearTotalWeight: command.ClearTotalWeight,
		Note:             command.Note,
	}, command.UpdatedBy, now); err != nil {
		return err
	}

	return s.repository.Update(ctx, weaging)
}
