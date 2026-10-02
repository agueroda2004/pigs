package medication

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxNameLength = 50

var (
	ErrInvalidID        = errors.New("El identificador del medicamento es obligatorio")
	ErrInvalidName      = errors.New("El nombre es obligatorio y debe tener como máximo 50 caracteres")
	ErrInvalidUpdate    = errors.New("Debe actualizar al menos un campo del medicamento")
	ErrInvalidCreatedBy = errors.New("El usuario que crea el medicamento es obligatorio")
	ErrInvalidUpdatedBy = errors.New("El usuario que actualiza el medicamento es obligatorio")
)

// Medication is the aggregate root that represents a medication (medicamento).
// It carries the audit fields and only exposes a name and its active flag.
type Medication struct {
	ID        uuid.UUID
	Name      string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

// MedicationOption is a lightweight medication read model for selection lists.
// It only carries the identifier and name of a medication.
type MedicationOption struct {
	ID   uuid.UUID
	Name string
}

// NewMedication builds a medication after validating id, name and createdBy.
// It defaults Active to true and sets CreatedBy/UpdatedBy to the creator.
func NewMedication(id uuid.UUID, name string, createdBy uuid.UUID, now time.Time) (*Medication, error) {
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

	return &Medication{
		ID:        id,
		Name:      name,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}, nil
}

// Update applies only the non-nil name and active fields of a medication.
// It validates each provided value and records updatedBy plus the timestamp.
func (m *Medication) Update(name *string, active *bool, updatedBy uuid.UUID, now time.Time) error {
	if m == nil || m.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if name == nil && active == nil {
		return ErrInvalidUpdate
	}

	validatedName := m.Name
	if name != nil {
		var err error
		validatedName, err = validateName(*name)
		if err != nil {
			return err
		}
	}

	m.Name = validatedName
	if active != nil {
		m.Active = *active
	}
	m.UpdatedAt = now
	m.UpdatedBy = updatedBy
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
