package application

import (
	"context"

	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sow/ports"
)

type ListSowDropdownService struct {
	repository ports.SowRepository
}

// NewListSowDropdownService builds a list-sow-dropdown use case with its repository.
// It returns a service ready to return lightweight sows for selection lists.
func NewListSowDropdownService(repository ports.SowRepository) *ListSowDropdownService {
	return &ListSowDropdownService{repository: repository}
}

// Execute returns the id and code of the sows matching the active flag and state
// list. A nil active applies no active filter and empty states apply no state
// filter, so an empty request returns every sow.
func (s *ListSowDropdownService) Execute(ctx context.Context, active *bool, states []sowdomain.State) ([]sowdomain.SowDropdown, error) {
	return s.repository.ListDropdown(ctx, active, states)
}
