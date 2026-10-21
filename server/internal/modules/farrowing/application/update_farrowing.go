package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	"server/internal/modules/farrowing/ports"
)

// UpdateFarrowingCommand carries the editable fields of a farrowing update.
// A nil pointer means the field is omitted and stays unchanged, nullable string
// fields are cleared with an empty string, and the farrow date plus the piglet
// counts are only accepted when they match the stored values. The operator and
// medication lists are only replaced when their Update flag is set, and the sow,
// service and nurse fields are never editable here.
type UpdateFarrowingCommand struct {
	FarrowDate        *time.Time
	StartTime         *string
	EndTime           *string
	Location          *string
	LiveBorn          *int
	Stillborn         *int
	Mummified         *int
	LitterWeight      *float64
	StillbornWeight   *float64
	IsManipulated     *bool
	Note              *string
	Operators         []CreateFarrowingOperatorInput
	UpdateOperators   bool
	Medications       []CreateFarrowingMedicationInput
	UpdateMedications bool
	UpdatedBy         uuid.UUID
}

type UpdateFarrowingService struct {
	repository ports.FarrowingRepository
	clock      func() time.Time
}

// NewUpdateFarrowingService builds an update-farrowing use case with its
// repository and clock, ready to execute UpdateFarrowingCommand values.
func NewUpdateFarrowingService(repository ports.FarrowingRepository, clock func() time.Time) *UpdateFarrowingService {
	return &UpdateFarrowingService{repository: repository, clock: clock}
}

// Execute applies the editable fields to a farrowing.
// It reloads the farrowing, validates any replaced operator or medication and then
// persists the farrowing and its links atomically.
func (s *UpdateFarrowingService) Execute(ctx context.Context, id uuid.UUID, command UpdateFarrowingCommand) error {
	farrowing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if command.UpdateOperators || command.UpdateMedications {
		if err := ensureReferencesExist(ctx, s.repository, command.Operators, command.Medications); err != nil {
			return err
		}
	}

	if err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{
		FarrowDate:        command.FarrowDate,
		StartTime:         command.StartTime,
		EndTime:           command.EndTime,
		Location:          command.Location,
		LiveBorn:          command.LiveBorn,
		Stillborn:         command.Stillborn,
		Mummified:         command.Mummified,
		LitterWeight:      command.LitterWeight,
		StillbornWeight:   command.StillbornWeight,
		IsManipulated:     command.IsManipulated,
		Note:              command.Note,
		Operators:         buildOperatorParams(command.Operators),
		UpdateOperators:   command.UpdateOperators,
		Medications:       buildMedicationParams(command.Medications),
		UpdateMedications: command.UpdateMedications,
	}, command.UpdatedBy, s.clock()); err != nil {
		return err
	}

	return s.repository.Update(ctx, farrowing, command.UpdateOperators, command.UpdateMedications)
}
