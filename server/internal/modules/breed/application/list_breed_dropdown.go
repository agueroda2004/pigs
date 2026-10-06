package application

import (
	"context"

	breeddomain "server/internal/modules/breed/domain"
	"server/internal/modules/breed/ports"
)

type ListBreedDropdownService struct {
	repository ports.BreedRepository
}

// NewListBreedDropdownService builds a list-breed-dropdown use case with its repository.
// It returns a service ready to return breeds for selection lists.
func NewListBreedDropdownService(repository ports.BreedRepository) *ListBreedDropdownService {
	return &ListBreedDropdownService{repository: repository}
}

// Execute returns the id, name and active flag of the breeds for a selection list.
// When active is true it only returns active breeds; when false it returns every breed.
func (s *ListBreedDropdownService) Execute(ctx context.Context, active bool) ([]breeddomain.BreedDropdown, error) {
	return s.repository.ListDropdown(ctx, active)
}
