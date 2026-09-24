package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	operatordomain "server/internal/modules/operator/domain"
	"server/internal/modules/operator/ports"
)

type UpdateOperatorCommand struct {
	Name      *string
	Active    *bool
	UpdatedBy uuid.UUID
}

type UpdateOperatorService struct {
	repository ports.OperatorRepository
	clock      func() time.Time
}

// NewUpdateOperatorService builds an update-operator use case with its repository and clock.
// It returns a service ready to execute UpdateOperatorCommand values.
func NewUpdateOperatorService(repository ports.OperatorRepository, clock func() time.Time) *UpdateOperatorService {
	return &UpdateOperatorService{repository: repository, clock: clock}
}

// Execute applies the provided fields to an existing operator by its identifier.
// It loads the operator, validates the change and persists the updated entity.
func (s *UpdateOperatorService) Execute(
	ctx context.Context,
	operatorID uuid.UUID,
	command UpdateOperatorCommand,
) (*operatordomain.Operator, error) {
	currentOperator, err := s.repository.GetByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	if err := currentOperator.Update(command.Name, command.Active, command.UpdatedBy, s.clock()); err != nil {
		return nil, err
	}

	if err := s.repository.Update(ctx, currentOperator); err != nil {
		return nil, err
	}
	return currentOperator, nil
}
