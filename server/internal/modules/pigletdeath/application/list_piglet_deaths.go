package application

import (
	"context"

	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	"server/internal/modules/pigletdeath/ports"
)

type ListPigletDeathsService struct {
	repository ports.PigletDeathRepository
}

// NewListPigletDeathsService builds a list-piglet-deaths use case with its repository.
// It returns a service ready to return piglet deaths matching the given filter.
func NewListPigletDeathsService(repository ports.PigletDeathRepository) *ListPigletDeathsService {
	return &ListPigletDeathsService{repository: repository}
}

// Execute returns the piglet deaths matching the filter.
// A zero-value filter returns every registered piglet death.
func (s *ListPigletDeathsService) Execute(ctx context.Context, filter ports.PigletDeathFilter) ([]*pigletdeathdomain.PigletDeath, error) {
	return s.repository.List(ctx, filter)
}
