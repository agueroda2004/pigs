package application

import (
	"context"

	breeddomain "server/internal/modules/breed/domain"
	"server/internal/modules/breed/ports"
)

type ListBreedOptionsService struct {
	repository ports.BreedRepository
}

// NewListBreedOptionsService builds a list-breed-options use case with its repository.
// It returns a service ready to return the active breeds for selection lists.
func NewListBreedOptionsService(repository ports.BreedRepository) *ListBreedOptionsService {
	return &ListBreedOptionsService{repository: repository}
}

// Execute returns the id and name of every active breed ordered by name.
// It is intended for dropdowns that only need a lightweight read model.
func (s *ListBreedOptionsService) Execute(ctx context.Context) ([]breeddomain.BreedOption, error) {
	return s.repository.ListActiveOptions(ctx)
}
