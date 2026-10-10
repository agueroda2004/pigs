package application

import (
	"context"

	operatordomain "server/internal/modules/operator/domain"
	"server/internal/modules/operator/ports"
)

type ListOperatorDropdownService struct {
	repository ports.OperatorRepository
}

// NewListOperatorDropdownService builds a list-operator-dropdown use case with its repository.
// It returns a service ready to return lightweight operators for selection lists.
func NewListOperatorDropdownService(repository ports.OperatorRepository) *ListOperatorDropdownService {
	return &ListOperatorDropdownService{repository: repository}
}

// Execute returns the id, name and active flag of the operators matching the active flag.
// A nil active applies no filter and returns every operator, true returns only the active
// ones and false only the inactive ones.
func (s *ListOperatorDropdownService) Execute(ctx context.Context, active *bool) ([]operatordomain.OperatorDropdown, error) {
	return s.repository.ListDropdown(ctx, active)
}
