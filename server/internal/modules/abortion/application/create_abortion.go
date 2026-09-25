package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	abortiondomain "server/internal/modules/abortion/domain"
	"server/internal/modules/abortion/ports"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

var (
	ErrSowNotGestating     = errors.New("La cerda no está gestando")
	ErrServiceNotConfirmed = errors.New("El último servicio no está confirmado")
)

// CreateAbortionCommand carries the fields required to register an abortion.
// The service is derived by the use case from the sow's last service.
type CreateAbortionCommand struct {
	SowID        uuid.UUID
	AbortionDate time.Time
	Cause        abortiondomain.Cause
	Note         *string
	CreatedBy    uuid.UUID
}

type CreateAbortionService struct {
	repository ports.AbortionRepository
	clock      func() time.Time
}

// NewCreateAbortionService builds a create-abortion use case with its repository and clock.
// It returns a service ready to execute CreateAbortionCommand values.
func NewCreateAbortionService(repository ports.AbortionRepository, clock func() time.Time) *CreateAbortionService {
	return &CreateAbortionService{repository: repository, clock: clock}
}

// Execute registers an abortion and moves the sow and its last service to aborted.
// It validates that the sow is gestating and its last service is confirmed, then
// persists the abortion, the sow state and the service state in one transaction.
func (s *CreateAbortionService) Execute(ctx context.Context, command CreateAbortionCommand) (*abortiondomain.Abortion, error) {
	sow, err := s.repository.GetSow(ctx, command.SowID)
	if err != nil {
		return nil, err
	}
	if sow.State != sowdomain.StatePregnant {
		return nil, ErrSowNotGestating
	}

	service, err := s.repository.GetLastService(ctx, command.SowID)
	if err != nil {
		return nil, err
	}
	if service.State != servicedomain.StateConfirmed {
		return nil, ErrServiceNotConfirmed
	}

	now := s.clock()
	newAbortion, err := abortiondomain.NewAbortion(abortiondomain.NewAbortionParams{
		ID:           uuid.New(),
		SowID:        command.SowID,
		ServiceID:    service.ID,
		AbortionDate: command.AbortionDate,
		Cause:        command.Cause,
		Note:         command.Note,
		CreatedBy:    command.CreatedBy,
	}, latestMountDate(service), now)
	if err != nil {
		return nil, err
	}

	if err := sow.ChangeState(sowdomain.StateAborted, command.CreatedBy, now); err != nil {
		return nil, err
	}
	if err := service.ChangeState(servicedomain.StateAborted, command.CreatedBy, now); err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, newAbortion, sow, service); err != nil {
		return nil, err
	}
	return newAbortion, nil
}

// latestMountDate returns the most recent mount date of a service.
// It returns the zero time when the service has no mounts.
func latestMountDate(service *servicedomain.Service) time.Time {
	var latest time.Time
	for _, mount := range service.Mounts {
		if mount.MountDate.After(latest) {
			latest = mount.MountDate
		}
	}
	return latest
}
