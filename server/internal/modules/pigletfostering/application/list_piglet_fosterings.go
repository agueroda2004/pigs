package application

import (
	"context"

	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	"server/internal/modules/pigletfostering/ports"
)

type ListPigletFosteringsService struct {
	repository ports.PigletFosteringRepository
}

// NewListPigletFosteringsService builds a list-fosterings use case with its repository.
// It returns a service ready to return fosterings matching the given filter.
func NewListPigletFosteringsService(repository ports.PigletFosteringRepository) *ListPigletFosteringsService {
	return &ListPigletFosteringsService{repository: repository}
}

// Execute returns the fosterings matching the filter.
// A zero-value filter returns every registered fostering.
func (s *ListPigletFosteringsService) Execute(ctx context.Context, filter ports.PigletFosteringFilter) ([]*pigletfosteringdomain.PigletFostering, error) {
	return s.repository.List(ctx, filter)
}
