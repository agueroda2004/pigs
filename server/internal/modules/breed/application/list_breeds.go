package application

import (
	"context"

	breeddomain "server/internal/modules/breed/domain"
	"server/internal/modules/breed/ports"
)

type ListBreedsService struct {
	repository ports.BreedRepository
}

// NewListBreedsService builds a list-breeds use case with its repository.
// It returns a service ready to return every registered breed.
func NewListBreedsService(repository ports.BreedRepository) *ListBreedsService {
	return &ListBreedsService{repository: repository}
}

// Execute returns every registered breed without pagination or filtering.
// It is intended for the small catalog of breeds managed by the farm.
func (s *ListBreedsService) Execute(ctx context.Context) ([]*breeddomain.Breed, error) {
	return s.repository.List(ctx)
}
