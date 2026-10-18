package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	abortiondomain "server/internal/modules/abortion/domain"
	"server/internal/modules/abortion/ports"
	servicedomain "server/internal/modules/service/domain"
)

// UpdateAbortionCommand carries the editable fields of an abortion update.
// A nil field is left unchanged; the sow and the service are never editable.
type UpdateAbortionCommand struct {
	AbortionDate *time.Time
	Cause        *abortiondomain.Cause
	Note         *string
	UpdatedBy    uuid.UUID
}

type UpdateAbortionService struct {
	repository ports.AbortionRepository
	clock      func() time.Time
}

// NewUpdateAbortionService builds an update-abortion use case with its repository
// and clock, ready to execute UpdateAbortionCommand values.
func NewUpdateAbortionService(repository ports.AbortionRepository, clock func() time.Time) *UpdateAbortionService {
	return &UpdateAbortionService{repository: repository, clock: clock}
}

// Execute applies the editable fields to an abortion. It reloads the sow entry
// date, the service mounts and the next service so a changed date is validated
// against the same bounds used on creation plus the following service, then
// persists the abortion.
func (s *UpdateAbortionService) Execute(ctx context.Context, id uuid.UUID, command UpdateAbortionCommand) error {
	abortion, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	sow, err := s.repository.GetSow(ctx, abortion.SowID)
	if err != nil {
		return err
	}

	service, err := s.repository.GetService(ctx, abortion.ServiceID)
	if err != nil {
		return err
	}

	nextService, err := s.repository.GetNextService(ctx, abortion.SowID, abortion.ServiceID)
	if err != nil && !errors.Is(err, ports.ErrServiceNotFound) {
		return err
	}

	reference := abortiondomain.Reference{
		EntryDate:     sow.EntryDate,
		LastMountDate: latestMountDate(service),
	}
	if nextService != nil {
		reference.NextMountDate = earliestMountDate(nextService)
	}

	if err := abortion.Update(abortiondomain.UpdateAbortionParams{
		AbortionDate: command.AbortionDate,
		Cause:        command.Cause,
		Note:         command.Note,
	}, reference, command.UpdatedBy, s.clock()); err != nil {
		return err
	}

	return s.repository.Update(ctx, abortion)
}

// earliestMountDate returns the first mount date of a service.
// It returns the zero time when the service has no mounts.
func earliestMountDate(service *servicedomain.Service) time.Time {
	var earliest time.Time
	for _, mount := range service.Mounts {
		if earliest.IsZero() || mount.MountDate.Before(earliest) {
			earliest = mount.MountDate
		}
	}
	return earliest
}
