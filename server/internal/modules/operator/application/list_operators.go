package application

import (
	"context"

	operatordomain "server/internal/modules/operator/domain"
	"server/internal/modules/operator/ports"
)

type ListOperatorsService struct {
	repository ports.OperatorRepository
}

// NewListOperatorsService builds a list-operators use case with its repository.
// It returns a service ready to return every registered operator.
func NewListOperatorsService(repository ports.OperatorRepository) *ListOperatorsService {
	return &ListOperatorsService{repository: repository}
}

// Execute returns every registered operator without pagination or filtering.
// It uses the lightweight read model, so the audit fields are not returned.
func (s *ListOperatorsService) Execute(ctx context.Context) ([]operatordomain.OperatorListItem, error) {
	return s.repository.List(ctx)
}
