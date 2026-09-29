package application

import (
	"context"

	boarremovaldomain "server/internal/modules/boarremoval/domain"
	"server/internal/modules/boarremoval/ports"
)

type ListBoarRemovalsService struct {
	repository ports.BoarRemovalRepository
}

// NewListBoarRemovalsService builds a list-removals use case with its repository.
// It returns a service ready to return removals matching the given filter.
func NewListBoarRemovalsService(repository ports.BoarRemovalRepository) *ListBoarRemovalsService {
	return &ListBoarRemovalsService{repository: repository}
}

// Execute returns the removals matching the filter.
// A zero-value filter returns every registered removal.
func (s *ListBoarRemovalsService) Execute(ctx context.Context, filter ports.BoarRemovalFilter) ([]*boarremovaldomain.BoarRemoval, error) {
	return s.repository.List(ctx, filter)
}
