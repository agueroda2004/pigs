package operator

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxNameLength = 100

var (
	ErrInvalidID        = errors.New("El identificador del operador es obligatorio")
	ErrInvalidName      = errors.New("El nombre es obligatorio y debe tener como máximo 100 caracteres")
	ErrInvalidUpdate    = errors.New("Debe actualizar al menos un campo del operador")
	ErrInvalidCreatedBy = errors.New("El usuario que crea el operador es obligatorio")
	ErrInvalidUpdatedBy = errors.New("El usuario que actualiza el operador es obligatorio")
)

// Operator is the aggregate root that represents an operator (operador) in the farm.
// It carries the audit fields and only exposes a name and its active flag.
type Operator struct {
	ID        uuid.UUID
	Name      string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

// NewOperator builds an operator after validating id, name and createdBy.
// It defaults Active to true and sets CreatedBy/UpdatedBy to the creator.
func NewOperator(id uuid.UUID, name string, createdBy uuid.UUID, now time.Time) (*Operator, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidID
	}

	name, err := validateName(name)
	if err != nil {
		return nil, err
	}

	if createdBy == uuid.Nil {
		return nil, ErrInvalidCreatedBy
	}

	return &Operator{
		ID:        id,
		Name:      name,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}, nil
}

// Update applies only the non-nil name and active fields of an operator.
// It validates each provided value and records updatedBy plus the timestamp.
func (o *Operator) Update(name *string, active *bool, updatedBy uuid.UUID, now time.Time) error {
	if o == nil || o.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if name == nil && active == nil {
		return ErrInvalidUpdate
	}

	validatedName := o.Name
	if name != nil {
		var err error
		validatedName, err = validateName(*name)
		if err != nil {
			return err
		}
	}

	o.Name = validatedName
	if active != nil {
		o.Active = *active
	}
	o.UpdatedAt = now
	o.UpdatedBy = updatedBy
	return nil
}

// validateName trims the name and requires it to be non-empty.
// It returns ErrInvalidName when the name exceeds the maximum length.
func validateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > maxNameLength {
		return "", ErrInvalidName
	}
	return name, nil
}
