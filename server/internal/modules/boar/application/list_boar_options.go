package application

import (
	"context"

	boardomain "server/internal/modules/boar/domain"
	"server/internal/modules/boar/ports"
)

type ListBoarOptionsService struct {
	repository ports.BoarRepository
}

// NewListBoarOptionsService builds a list-boar-options use case with its repository.
// It returns a service ready to return lightweight boars for selection lists.
func NewListBoarOptionsService(repository ports.BoarRepository) *ListBoarOptionsService {
	return &ListBoarOptionsService{repository: repository}
}

// Execute returns the id and code of the boars matching the active filter.
// A true active restricts to alive boars while nil or false return every state.
func (s *ListBoarOptionsService) Execute(ctx context.Context, active *bool) ([]boardomain.BoarOption, error) {
	return s.repository.ListOptions(ctx, active)
}
