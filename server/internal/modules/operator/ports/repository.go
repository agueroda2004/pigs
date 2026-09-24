package ports

import (
	"context"

	"github.com/google/uuid"

	operatordomain "server/internal/modules/operator/domain"
)

type OperatorRepository interface {
	Create(ctx context.Context, operator *operatordomain.Operator) error
	GetByID(ctx context.Context, id uuid.UUID) (*operatordomain.Operator, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	List(ctx context.Context) ([]*operatordomain.Operator, error)
	Update(ctx context.Context, operator *operatordomain.Operator) error
}
