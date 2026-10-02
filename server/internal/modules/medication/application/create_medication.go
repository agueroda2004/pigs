package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	medicationdomain "server/internal/modules/medication/domain"
	"server/internal/modules/medication/ports"
)

var ErrMedicationNameAlreadyExists = ports.ErrMedicationNameAlreadyUsed

type CreateMedicationCommand struct {
	Name      string
	CreatedBy uuid.UUID
}

type CreateMedicationService struct {
	repository ports.MedicationRepository
	clock      func() time.Time
}

// NewCreateMedicationService builds a create-medication use case with its repository and clock.
// It returns a service ready to execute CreateMedicationCommand values.
func NewCreateMedicationService(repository ports.MedicationRepository, clock func() time.Time) *CreateMedicationService {
	return &CreateMedicationService{repository: repository, clock: clock}
}

// Execute creates a medication after ensuring its name is free.
// It returns ErrMedicationNameAlreadyExists when the name is already taken.
func (s *CreateMedicationService) Execute(ctx context.Context, command CreateMedicationCommand) (*medicationdomain.Medication, error) {
	name := strings.TrimSpace(command.Name)
	exists, err := s.repository.ExistsByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrMedicationNameAlreadyExists
	}

	newMedication, err := medicationdomain.NewMedication(uuid.New(), name, command.CreatedBy, s.clock())
	if err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, newMedication); err != nil {
		return nil, err
	}
	return newMedication, nil
}
