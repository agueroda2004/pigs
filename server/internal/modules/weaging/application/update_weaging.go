package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	weagingdomain "server/internal/modules/weaging/domain"
	"server/internal/modules/weaging/ports"
)

// UpdateWeagingCommand carries the editable fields of a weaging update.
// A nil field is left unchanged and ClearTotalWeight removes the weight; the sow,
// farrowing, date and quantity are never editable.
type UpdateWeagingCommand struct {
	TotalWeight      *float64
	ClearTotalWeight bool
	Destination      *string
	Note             *string
	UpdatedBy        uuid.UUID
}

type UpdateWeagingService struct {
	repository ports.WeagingRepository
	clock      func() time.Time
}

// NewUpdateWeagingService builds an update-weaging use case with its repository and clock.
// It returns a service ready to execute UpdateWeagingCommand values.
func NewUpdateWeagingService(repository ports.WeagingRepository, clock func() time.Time) *UpdateWeagingService {
	return &UpdateWeagingService{repository: repository, clock: clock}
}

// Execute applies the editable fields to a weaging.
// It validates the update and persists the weight, destination, note and audit fields only.
func (s *UpdateWeagingService) Execute(ctx context.Context, id uuid.UUID, command UpdateWeagingCommand) error {
	weaging, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := s.clock()
	if err := weaging.Update(weagingdomain.UpdateWeagingParams{
		TotalWeight:      command.TotalWeight,
		ClearTotalWeight: command.ClearTotalWeight,
		Destination:      command.Destination,
		Note:             command.Note,
	}, command.UpdatedBy, now); err != nil {
		return err
	}

	return s.repository.Update(ctx, weaging)
}
