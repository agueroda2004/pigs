package application

import (
	"context"

	"github.com/google/uuid"

	"server/internal/modules/operator/ports"
)

type DeleteOperatorService struct {
	repository ports.OperatorRepository
}

// NewDeleteOperatorService builds a delete-operator use case with its repository.
// It returns a service ready to execute operator deletions.
func NewDeleteOperatorService(repository ports.OperatorRepository) *DeleteOperatorService {
	return &DeleteOperatorService{repository: repository}
}

// Execute deletes an operator by its identifier.
// It returns ErrOperatorNotFound when the operator does not exist and
// ErrOperatorInUse when the operator still has records linked to it.
func (s *DeleteOperatorService) Execute(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}
