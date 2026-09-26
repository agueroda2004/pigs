package application

import (
	"context"

	sowremovaldomain "server/internal/modules/sowremoval/domain"
	"server/internal/modules/sowremoval/ports"
)

type ListSowRemovalsService struct {
	repository ports.SowRemovalRepository
}

// NewListSowRemovalsService builds a list-removals use case with its repository.
// It returns a service ready to return removals matching the given filter.
func NewListSowRemovalsService(repository ports.SowRemovalRepository) *ListSowRemovalsService {
	return &ListSowRemovalsService{repository: repository}
}

// Execute returns the removals matching the filter.
// A zero-value filter returns every registered removal.
func (s *ListSowRemovalsService) Execute(ctx context.Context, filter ports.SowRemovalFilter) ([]*sowremovaldomain.SowRemoval, error) {
	return s.repository.List(ctx, filter)
}
