package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	medicationdomain "server/internal/modules/medication/domain"
	"server/internal/modules/medication/ports"
)

type UpdateMedicationCommand struct {
	Name      *string
	Active    *bool
	UpdatedBy uuid.UUID
}

type UpdateMedicationService struct {
	repository ports.MedicationRepository
	clock      func() time.Time
}

// NewUpdateMedicationService builds an update-medication use case with its repository and clock.
// It returns a service ready to execute UpdateMedicationCommand values.
func NewUpdateMedicationService(repository ports.MedicationRepository, clock func() time.Time) *UpdateMedicationService {
	return &UpdateMedicationService{repository: repository, clock: clock}
}

// Execute applies the provided fields to an existing medication by its identifier.
// It loads the medication, validates the change and persists the updated entity.
func (s *UpdateMedicationService) Execute(
	ctx context.Context,
	medicationID uuid.UUID,
	command UpdateMedicationCommand,
) (*medicationdomain.Medication, error) {
	currentMedication, err := s.repository.GetByID(ctx, medicationID)
	if err != nil {
		return nil, err
	}

	if err := currentMedication.Update(command.Name, command.Active, command.UpdatedBy, s.clock()); err != nil {
		return nil, err
	}

	if err := s.repository.Update(ctx, currentMedication); err != nil {
		return nil, err
	}
	return currentMedication, nil
}
