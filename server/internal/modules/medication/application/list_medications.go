package application

import (
	"context"

	medicationdomain "server/internal/modules/medication/domain"
	"server/internal/modules/medication/ports"
)

type ListMedicationsService struct {
	repository ports.MedicationRepository
}

// NewListMedicationsService builds a list-medications use case with its repository.
// It returns a service ready to return medications matching the given filter.
func NewListMedicationsService(repository ports.MedicationRepository) *ListMedicationsService {
	return &ListMedicationsService{repository: repository}
}

// Execute returns the medications matching the filter, ordered by name.
// A zero-value filter returns every registered medication.
func (s *ListMedicationsService) Execute(ctx context.Context, filter ports.MedicationFilter) ([]*medicationdomain.Medication, error) {
	return s.repository.List(ctx, filter)
}
