package application

import (
	"context"

	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sow/ports"
)

type ListSowsService struct {
	repository ports.SowRepository
}

// NewListSowsService builds a list-sows use case with its repository.
// It returns a service ready to return sows matching the given filter.
func NewListSowsService(repository ports.SowRepository) *ListSowsService {
	return &ListSowsService{repository: repository}
}

// Execute returns the sows matching the filter, ordered by code.
// A zero-value filter returns every registered sow.
func (s *ListSowsService) Execute(ctx context.Context, filter ports.SowFilter) ([]*sowdomain.Sow, error) {
	return s.repository.List(ctx, filter)
}
