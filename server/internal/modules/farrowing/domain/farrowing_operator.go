package farrowing

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidFarrowingOperatorID        = errors.New("El identificador del operador del parto es obligatorio")
	ErrInvalidFarrowingOperatorFarrowing = errors.New("El parto del operador es obligatorio")
	ErrInvalidFarrowingOperatorOperator  = errors.New("El operador del parto es obligatorio")
)

// FarrowingOperator links a farrowing with an operator that took part in it.
// It is a lightweight join row that only records the audit timestamp.
type FarrowingOperator struct {
	ID          uuid.UUID
	FarrowingID uuid.UUID
	OperatorID  uuid.UUID
	CreatedAt   time.Time
}

// NewFarrowingOperatorParams holds the fields required to build a join row.
// The farrowing id and the timestamp are supplied by the caller.
type NewFarrowingOperatorParams struct {
	ID         uuid.UUID
	OperatorID uuid.UUID
}

// NewFarrowingOperator builds a farrowing-operator link after validating its ids.
// It requires a non-nil id, a non-nil farrowing id and a non-nil operator id.
func NewFarrowingOperator(params NewFarrowingOperatorParams, farrowingID uuid.UUID, now time.Time) (*FarrowingOperator, error) {
	if params.ID == uuid.Nil {
		return nil, ErrInvalidFarrowingOperatorID
	}
	if farrowingID == uuid.Nil {
		return nil, ErrInvalidFarrowingOperatorFarrowing
	}
	if params.OperatorID == uuid.Nil {
		return nil, ErrInvalidFarrowingOperatorOperator
	}

	return &FarrowingOperator{
		ID:          params.ID,
		FarrowingID: farrowingID,
		OperatorID:  params.OperatorID,
		CreatedAt:   now,
	}, nil
}
