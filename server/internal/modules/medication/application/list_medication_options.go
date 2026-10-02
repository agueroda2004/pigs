package application

import (
	"context"

	medicationdomain "server/internal/modules/medication/domain"
	"server/internal/modules/medication/ports"
)

type ListMedicationOptionsService struct {
	repository ports.MedicationRepository
}

// NewListMedicationOptionsService builds a list-medication-options use case with its repository.
// It returns a service ready to return lightweight medications for selection lists.
func NewListMedicationOptionsService(repository ports.MedicationRepository) *ListMedicationOptionsService {
	return &ListMedicationOptionsService{repository: repository}
}

// Execute returns the id and name of the medications matching the active filter.
// A true active restricts to active medications while nil returns every medication.
func (s *ListMedicationOptionsService) Execute(ctx context.Context, active *bool) ([]medicationdomain.MedicationOption, error) {
	return s.repository.ListOptions(ctx, active)
}
