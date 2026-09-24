package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	operatordomain "server/internal/modules/operator/domain"
	"server/internal/modules/operator/ports"
)

var ErrOperatorNameAlreadyExists = ports.ErrOperatorNameAlreadyUsed

type CreateOperatorCommand struct {
	Name      string
	CreatedBy uuid.UUID
}

type CreateOperatorService struct {
	repository ports.OperatorRepository
	clock      func() time.Time
}

// NewCreateOperatorService builds a create-operator use case with its repository and clock.
// It returns a service ready to execute CreateOperatorCommand values.
func NewCreateOperatorService(repository ports.OperatorRepository, clock func() time.Time) *CreateOperatorService {
	return &CreateOperatorService{repository: repository, clock: clock}
}

// Execute creates an operator after ensuring its name is free.
// It returns ErrOperatorNameAlreadyExists when the name is already taken.
func (s *CreateOperatorService) Execute(ctx context.Context, command CreateOperatorCommand) (*operatordomain.Operator, error) {
	name := strings.TrimSpace(command.Name)
	exists, err := s.repository.ExistsByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrOperatorNameAlreadyExists
	}

	newOperator, err := operatordomain.NewOperator(uuid.New(), name, command.CreatedBy, s.clock())
	if err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, newOperator); err != nil {
		return nil, err
	}
	return newOperator, nil
}
