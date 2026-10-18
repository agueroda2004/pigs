package application

import (
	"context"

	"github.com/google/uuid"

	"server/internal/modules/medication/ports"
)

type DeleteMedicationService struct {
	repository ports.MedicationRepository
}

// NewDeleteMedicationService builds a delete-medication use case with its repository.
// It returns a service ready to execute medication deletions.
func NewDeleteMedicationService(repository ports.MedicationRepository) *DeleteMedicationService {
	return &DeleteMedicationService{repository: repository}
}

// Execute deletes a medication by its identifier.
// It returns ErrMedicationNotFound when the medication does not exist and
// ErrMedicationInUse when the medication still has records linked to it.
func (s *DeleteMedicationService) Execute(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}
