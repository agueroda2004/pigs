package application

import (
	"context"

	boardomain "server/internal/modules/boar/domain"
	"server/internal/modules/boar/ports"
)

type ListBoarsService struct {
	repository ports.BoarRepository
}

// NewListBoarsService builds a list-boars use case with its repository.
// It returns a service ready to return boars matching the given filter.
func NewListBoarsService(repository ports.BoarRepository) *ListBoarsService {
	return &ListBoarsService{repository: repository}
}

// Execute returns the boars matching the filter, ordered by code.
// A zero-value filter returns every registered boar.
func (s *ListBoarsService) Execute(ctx context.Context, filter ports.BoarFilter) ([]*boardomain.Boar, error) {
	return s.repository.List(ctx, filter)
}
