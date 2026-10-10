package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
	sowdomain "server/internal/modules/sow/domain"
)

var (
	ErrSowNotEligible       = errors.New("La cerda no está en un estado válido para el servicio")
	ErrBoarNotEligible      = errors.New("El verraco no está disponible para el servicio")
	ErrOperatorNotAvailable = errors.New("El operador no está disponible")
	ErrMountBeforeBoarEntry = errors.New("La fecha de monta no puede ser anterior al ingreso del verraco")
)

// CreateMountCommand carries the fields required to register a single mount.
// The type is optional and defaults to MountTypeArtificial inside the domain.
type CreateMountCommand struct {
	BoarID     uuid.UUID
	OperatorID uuid.UUID
	MountDate  time.Time
	Type       servicedomain.MountType
	Note       *string
}

type CreateServiceCommand struct {
	SowID     uuid.UUID
	Note      *string
	Location  *string
	Mounts    []CreateMountCommand
	CreatedBy uuid.UUID
}

type CreateServiceService struct {
	repository ports.ServiceRepository
	clock      func() time.Time
}

// NewCreateServiceService builds a create-service use case with its repository and clock.
// It returns a service ready to execute CreateServiceCommand values.
func NewCreateServiceService(repository ports.ServiceRepository, clock func() time.Time) *CreateServiceService {
	return &CreateServiceService{repository: repository, clock: clock}
}

// Execute registers a service with its mounts and moves the sow to StatePregnant.
// It validates the sow state, every boar and operator availability, builds the
// domain aggregate and persists the service, mounts and sow state in one write.
func (s *CreateServiceService) Execute(ctx context.Context, command CreateServiceCommand) (*servicedomain.Service, error) {
	sow, err := s.repository.GetSow(ctx, command.SowID)
	if err != nil {
		return nil, err
	}
	if !isServiceableSowState(sow.State) {
		return nil, ErrSowNotEligible
	}

	mountParams := make([]servicedomain.NewMountParams, 0, len(command.Mounts))
	for _, mount := range command.Mounts {
		boar, err := s.repository.GetBoar(ctx, mount.BoarID)
		if err != nil {
			return nil, err
		}
		if !boar.Active || boar.State != boardomain.StateAlive {
			return nil, ErrBoarNotEligible
		}
		if truncateToDay(mount.MountDate).Before(truncateToDay(boar.EntryDate)) {
			return nil, ErrMountBeforeBoarEntry
		}

		operator, err := s.repository.GetOperator(ctx, mount.OperatorID)
		if err != nil {
			return nil, err
		}
		if !operator.Active {
			return nil, ErrOperatorNotAvailable
		}

		mountParams = append(mountParams, servicedomain.NewMountParams{
			ID:         uuid.New(),
			BoarID:     mount.BoarID,
			OperatorID: mount.OperatorID,
			MountDate:  mount.MountDate,
			Type:       mount.Type,
			Note:       mount.Note,
		})
	}

	reference, err := s.buildReference(ctx, sow)
	if err != nil {
		return nil, err
	}

	now := s.clock()
	newService, err := servicedomain.NewService(servicedomain.NewServiceParams{
		ID:        uuid.New(),
		SowID:     command.SowID,
		Note:      command.Note,
		Location:  command.Location,
		LastState: string(sow.State),
		CreatedBy: command.CreatedBy,
	}, mountParams, reference, now)
	if err != nil {
		return nil, err
	}

	if err := sow.ChangeState(sowdomain.StatePregnant, command.CreatedBy, now); err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, newService, sow); err != nil {
		return nil, err
	}
	return newService, nil
}

// isServiceableSowState reports whether a sow may be serviced.
// Only sows that are alive, weaned, aborted or already pregnant are accepted.
func isServiceableSowState(state sowdomain.State) bool {
	switch state {
	case sowdomain.StateAlive, sowdomain.StateWeaned, sowdomain.StateAborted, sowdomain.StatePregnant:
		return true
	default:
		return false
	}
}

// truncateToDay removes the time portion from a timestamp in UTC.
// It is used to compare mount dates with the boar entry date at day granularity.
func truncateToDay(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

// buildReference gathers the dates a new service must respect for a sow.
// It reads the sow entry date, the previous service's latest mount date and the
// last abortion date; missing previous events are left as zero values.
func (s *CreateServiceService) buildReference(ctx context.Context, sow *sowdomain.Sow) (servicedomain.Reference, error) {
	reference := servicedomain.Reference{EntryDate: sow.EntryDate}

	previous, err := s.repository.GetLastService(ctx, sow.ID)
	if err != nil && !errors.Is(err, ports.ErrServiceNotFound) {
		return servicedomain.Reference{}, err
	}
	if err == nil && previous != nil {
		reference.PreviousMountDate = latestMountDate(previous)
	}

	lastAbortion, err := s.repository.GetLastAbortionDate(ctx, sow.ID)
	if err != nil {
		return servicedomain.Reference{}, err
	}
	if lastAbortion != nil {
		reference.LastAbortionDate = *lastAbortion
	}

	return reference, nil
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
