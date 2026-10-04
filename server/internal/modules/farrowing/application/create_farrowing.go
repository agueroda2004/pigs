package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	"server/internal/modules/farrowing/ports"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

var (
	ErrSowNotGestating     = errors.New("La cerda no está gestando")
	ErrServiceNotConfirmed = errors.New("El servicio no está confirmado")
)

// CreateFarrowingOperatorInput carries the operator selected for a farrowing.
// It only needs the operator identifier referenced by the link row.
type CreateFarrowingOperatorInput struct {
	OperatorID uuid.UUID
}

// CreateFarrowingMedicationInput carries a medication applied during a farrowing.
// It records the medication, its dose and the operator that applied it.
type CreateFarrowingMedicationInput struct {
	MedicationID uuid.UUID
	Dose         float64
	AppliedBy    uuid.UUID
}

// CreateFarrowingCommand carries the fields required to register a farrowing.
// The sow is provided by the caller and the service is derived by the use case
// from the sow's last service, validating the reproductive result plus the
// operators and medications involved in the event.
type CreateFarrowingCommand struct {
	SowID           uuid.UUID
	FarrowDate      time.Time
	StartTime       *string
	EndTime         *string
	Location        *string
	LiveBorn        int
	Stillborn       int
	Mummified       int
	LitterWeight    *float64
	StillbornWeight *float64
	IsManipulated   bool
	Note            *string
	Operators       []CreateFarrowingOperatorInput
	Medications     []CreateFarrowingMedicationInput
	CreatedBy       uuid.UUID
}

type CreateFarrowingService struct {
	repository ports.FarrowingRepository
	clock      func() time.Time
}

// NewCreateFarrowingService builds a create-farrowing use case with its repository and clock.
// It returns a service ready to execute CreateFarrowingCommand values.
func NewCreateFarrowingService(repository ports.FarrowingRepository, clock func() time.Time) *CreateFarrowingService {
	return &CreateFarrowingService{repository: repository, clock: clock}
}

// Execute registers a farrowing and moves the sow and its service to their next state.
// It validates that the sow is gestating and its last service is confirmed, resolves
// every operator and medication, then persists the farrowing, its join rows, the sow
// state and the service state in one transaction.
func (s *CreateFarrowingService) Execute(ctx context.Context, command CreateFarrowingCommand) (*farrowingdomain.Farrowing, error) {
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

	if err := s.checkOperatorsExist(ctx, command.Operators, command.Medications); err != nil {
		return nil, err
	}

	now := s.clock()
	newFarrowing, err := farrowingdomain.NewFarrowing(
		farrowingdomain.NewFarrowingParams{
			ID:              uuid.New(),
			ServiceID:       service.ID,
			SowID:           command.SowID,
			FarrowDate:      command.FarrowDate,
			StartTime:       command.StartTime,
			EndTime:         command.EndTime,
			Location:        command.Location,
			LiveBorn:        command.LiveBorn,
			Stillborn:       command.Stillborn,
			Mummified:       command.Mummified,
			LitterWeight:    command.LitterWeight,
			StillbornWeight: command.StillbornWeight,
			IsManipulated:   command.IsManipulated,
			Note:            command.Note,
			CreatedBy:       command.CreatedBy,
		},
		buildOperatorParams(command.Operators),
		buildMedicationParams(command.Medications),
		latestMountDate(service),
		now,
	)
	if err != nil {
		return nil, err
	}

	if err := sow.ChangeState(sowdomain.StateLactating, command.CreatedBy, now); err != nil {
		return nil, err
	}
	if err := service.ChangeState(servicedomain.StateFinished, command.CreatedBy, now); err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, newFarrowing, sow, service); err != nil {
		return nil, err
	}
	return newFarrowing, nil
}

// checkOperatorsExist verifies that every referenced operator and medication exists.
// It also verifies each medication's applying operator, returning the repository
// not-found error of the missing entity.
func (s *CreateFarrowingService) checkOperatorsExist(
	ctx context.Context,
	operators []CreateFarrowingOperatorInput,
	medications []CreateFarrowingMedicationInput,
) error {
	seen := make(map[uuid.UUID]struct{}, len(operators))
	for _, input := range operators {
		if _, checked := seen[input.OperatorID]; checked {
			continue
		}
		seen[input.OperatorID] = struct{}{}
		if _, err := s.repository.GetOperator(ctx, input.OperatorID); err != nil {
			return err
		}
	}

	seenMedications := make(map[uuid.UUID]struct{}, len(medications))
	for _, input := range medications {
		if _, checked := seenMedications[input.MedicationID]; !checked {
			seenMedications[input.MedicationID] = struct{}{}
			if _, err := s.repository.GetMedication(ctx, input.MedicationID); err != nil {
				return err
			}
		}
		if _, checked := seen[input.AppliedBy]; checked {
			continue
		}
		seen[input.AppliedBy] = struct{}{}
		if _, err := s.repository.GetOperator(ctx, input.AppliedBy); err != nil {
			return err
		}
	}
	return nil
}

// buildOperatorParams assigns a fresh identifier to each operator link.
// It returns the params required by the farrowing domain constructor.
func buildOperatorParams(operators []CreateFarrowingOperatorInput) []farrowingdomain.NewFarrowingOperatorParams {
	params := make([]farrowingdomain.NewFarrowingOperatorParams, 0, len(operators))
	for _, input := range operators {
		params = append(params, farrowingdomain.NewFarrowingOperatorParams{
			ID:         uuid.New(),
			OperatorID: input.OperatorID,
		})
	}
	return params
}

// buildMedicationParams assigns a fresh identifier to each medication link.
// It returns the params required by the farrowing domain constructor.
func buildMedicationParams(medications []CreateFarrowingMedicationInput) []farrowingdomain.NewFarrowingMedicationParams {
	params := make([]farrowingdomain.NewFarrowingMedicationParams, 0, len(medications))
	for _, input := range medications {
		params = append(params, farrowingdomain.NewFarrowingMedicationParams{
			ID:           uuid.New(),
			MedicationID: input.MedicationID,
			Dose:         input.Dose,
			AppliedBy:    input.AppliedBy,
		})
	}
	return params
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
