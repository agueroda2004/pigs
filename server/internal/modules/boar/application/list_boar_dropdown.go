package application

import (
	"context"

	boardomain "server/internal/modules/boar/domain"
	"server/internal/modules/boar/ports"
)

type ListBoarDropdownService struct {
	repository ports.BoarRepository
}

// NewListBoarDropdownService builds a list-boar-dropdown use case with its repository.
// It returns a service ready to return boars for selection lists.
func NewListBoarDropdownService(repository ports.BoarRepository) *ListBoarDropdownService {
	return &ListBoarDropdownService{repository: repository}
}

// Execute returns the id, code and active flag of the boars for a selection list.
// A nil active returns every boar, true only the active boars and false only the
// inactive ones; an optional state narrows the result to that state while nil
// returns every state.
func (s *ListBoarDropdownService) Execute(ctx context.Context, active *bool, state *boardomain.State) ([]boardomain.BoarDropdown, error) {
	return s.repository.ListDropdown(ctx, active, state)
}
