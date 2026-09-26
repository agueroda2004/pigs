package application

import (
	"context"

	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sow/ports"
)

type ListSowOptionsService struct {
	repository ports.SowRepository
}

// NewListSowOptionsService builds a list-sow-options use case with its repository.
// It returns a service ready to return lightweight sows for selection lists.
func NewListSowOptionsService(repository ports.SowRepository) *ListSowOptionsService {
	return &ListSowOptionsService{repository: repository}
}

// Execute returns the id and code of the sows matching the active filter.
// A true active restricts to serviceable states while nil or false return every
// state, active or inactive.
func (s *ListSowOptionsService) Execute(ctx context.Context, active *bool) ([]sowdomain.SowOption, error) {
	return s.repository.ListOptions(ctx, active)
}
