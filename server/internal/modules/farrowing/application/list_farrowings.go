package application

import (
	"context"

	farrowingdomain "server/internal/modules/farrowing/domain"
	"server/internal/modules/farrowing/ports"
)

type ListFarrowingsService struct {
	repository ports.FarrowingRepository
}

// NewListFarrowingsService builds a list-farrowings use case with its repository.
// It returns a service ready to return farrowings matching the given filter.
func NewListFarrowingsService(repository ports.FarrowingRepository) *ListFarrowingsService {
	return &ListFarrowingsService{repository: repository}
}

// Execute returns the farrowings matching the filter.
// A zero-value filter returns every registered farrowing.
func (s *ListFarrowingsService) Execute(ctx context.Context, filter ports.FarrowingFilter) ([]*farrowingdomain.Farrowing, error) {
	return s.repository.List(ctx, filter)
}
