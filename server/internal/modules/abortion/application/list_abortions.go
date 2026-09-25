package application

import (
	"context"

	abortiondomain "server/internal/modules/abortion/domain"
	"server/internal/modules/abortion/ports"
)

type ListAbortionsService struct {
	repository ports.AbortionRepository
}

// NewListAbortionsService builds a list-abortions use case with its repository.
// It returns a service ready to return abortions matching the given filter.
func NewListAbortionsService(repository ports.AbortionRepository) *ListAbortionsService {
	return &ListAbortionsService{repository: repository}
}

// Execute returns the abortions matching the filter.
// A zero-value filter returns every registered abortion.
func (s *ListAbortionsService) Execute(ctx context.Context, filter ports.AbortionFilter) ([]*abortiondomain.Abortion, error) {
	return s.repository.List(ctx, filter)
}
