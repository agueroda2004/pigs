package application

import (
	"context"

	partialweagingdomain "server/internal/modules/partialweaging/domain"
	"server/internal/modules/partialweaging/ports"
)

type ListPartialWeagingsService struct {
	repository ports.PartialWeagingRepository
}

// NewListPartialWeagingsService builds a list-partial-weagings use case with its repository.
// It returns a service ready to return partial weagings matching the given filter.
func NewListPartialWeagingsService(repository ports.PartialWeagingRepository) *ListPartialWeagingsService {
	return &ListPartialWeagingsService{repository: repository}
}

// Execute returns the partial weagings matching the filter.
// A zero-value filter returns every registered partial weaging.
func (s *ListPartialWeagingsService) Execute(ctx context.Context, filter ports.PartialWeagingFilter) ([]*partialweagingdomain.PartialWeaging, error) {
	return s.repository.List(ctx, filter)
}
