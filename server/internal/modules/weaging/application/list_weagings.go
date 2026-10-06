package application

import (
	"context"

	weagingdomain "server/internal/modules/weaging/domain"
	"server/internal/modules/weaging/ports"
)

type ListWeagingsService struct {
	repository ports.WeagingRepository
}

// NewListWeagingsService builds a list-weagings use case with its repository.
// It returns a service ready to return weagings matching the given filter.
func NewListWeagingsService(repository ports.WeagingRepository) *ListWeagingsService {
	return &ListWeagingsService{repository: repository}
}

// Execute returns the weagings matching the filter.
// A zero-value filter returns every registered weaging.
func (s *ListWeagingsService) Execute(ctx context.Context, filter ports.WeagingFilter) ([]*weagingdomain.Weaging, error) {
	return s.repository.List(ctx, filter)
}
