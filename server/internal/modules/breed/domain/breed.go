package breed

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxNameLength = 100

var (
	ErrInvalidID        = errors.New("El identificador de la raza es obligatorio")
	ErrInvalidName      = errors.New("El nombre es obligatorio y debe tener como máximo 100 caracteres")
	ErrInvalidUpdate    = errors.New("Debe actualizar al menos un campo de la raza")
	ErrInvalidCreatedBy = errors.New("El usuario que crea la raza es obligatorio")
	ErrInvalidUpdatedBy = errors.New("El usuario que actualiza la raza es obligatorio")
)

type Breed struct {
	ID        uuid.UUID
	Name      string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

// NewBreed builds a breed after validating id, name and createdBy.
// It defaults Active to true and sets CreatedBy/UpdatedBy to the creator.
func NewBreed(id uuid.UUID, name string, createdBy uuid.UUID, now time.Time) (*Breed, error) {
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

	return &Breed{
		ID:        id,
		Name:      name,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}, nil
}

// Update applies only the non-nil name and active fields of a breed.
// It validates each provided value and records updatedBy plus the timestamp.
func (b *Breed) Update(name *string, active *bool, updatedBy uuid.UUID, now time.Time) error {
	if b == nil || b.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if name == nil && active == nil {
		return ErrInvalidUpdate
	}

	validatedName := b.Name
	if name != nil {
		var err error
		validatedName, err = validateName(*name)
		if err != nil {
			return err
		}
	}

	b.Name = validatedName
	if active != nil {
		b.Active = *active
	}
	b.UpdatedAt = now
	b.UpdatedBy = updatedBy
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
