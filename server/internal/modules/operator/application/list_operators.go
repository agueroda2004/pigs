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
// It is intended for the small catalog of operators managed by the farm.
func (s *ListOperatorsService) Execute(ctx context.Context) ([]*operatordomain.Operator, error) {
	return s.repository.List(ctx)
}
