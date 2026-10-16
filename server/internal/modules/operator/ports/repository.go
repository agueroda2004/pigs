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
	// List returns the lightweight operator read models used by the listing.
	// It omits the audit fields and orders the operators by name.
	List(ctx context.Context) ([]operatordomain.OperatorListItem, error)
	// ListDropdown returns lightweight operator read models filtered by the
	// active flag. A nil active applies no filter and returns every operator.
	ListDropdown(ctx context.Context, active *bool) ([]operatordomain.OperatorDropdown, error)
	Update(ctx context.Context, operator *operatordomain.Operator) error
	// Delete removes an operator by id. It reports ErrOperatorNotFound when no
	// row matches and ErrOperatorInUse when the operator still has linked records.
	Delete(ctx context.Context, id uuid.UUID) error
}
