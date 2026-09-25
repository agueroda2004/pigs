package service

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxMountNoteLength = 500

// MountType represents how a mount was performed.
// Its values match the mount_type database enum.
type MountType string

const (
	MountTypeNatural    MountType = "Natural"
	MountTypeArtificial MountType = "Artificial"
)

var (
	ErrInvalidMountID        = errors.New("El identificador de la monta es obligatorio")
	ErrInvalidMountService   = errors.New("El servicio de la monta es obligatorio")
	ErrInvalidMountBoar      = errors.New("El verraco de la monta es obligatorio")
	ErrInvalidMountOperator  = errors.New("El operador de la monta es obligatorio")
	ErrInvalidMountNumber    = errors.New("El número de monta es obligatorio")
	ErrInvalidMountDate      = errors.New("La fecha de la monta es obligatoria")
	ErrInvalidMountType      = errors.New("El tipo de monta no es válido")
	ErrInvalidMountNote      = errors.New("La nota de la monta debe tener como máximo 500 caracteres")
	ErrInvalidMountCreatedBy = errors.New("El usuario que crea la monta es obligatorio")
)

// Mount is the aggregate that represents a single mount (monta) of a service.
// It references the boar and the operator that executed it and keeps its order.
type Mount struct {
	ID          uuid.UUID
	ServiceID   uuid.UUID
	BoarID      uuid.UUID
	OperatorID  uuid.UUID
	MountNumber int
	MountDate   time.Time
	Type        MountType
	Note        *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
}

// NewMountParams holds the fields required to build a new mount.
// Optional values are pointers and may be nil; the type defaults to
// MountTypeArtificial when it is empty.
type NewMountParams struct {
	ID         uuid.UUID
	BoarID     uuid.UUID
	OperatorID uuid.UUID
	MountDate  time.Time
	Type       MountType
	Note       *string
}

// NewMount builds a mount for a service after validating its fields.
// It defaults an empty type to MountTypeArtificial and sets the audit fields to
// the given creator and both timestamps to now.
func NewMount(
	params NewMountParams,
	serviceID uuid.UUID,
	mountNumber int,
	createdBy uuid.UUID,
	now time.Time,
) (*Mount, error) {
	if params.ID == uuid.Nil {
		return nil, ErrInvalidMountID
	}
	if serviceID == uuid.Nil {
		return nil, ErrInvalidMountService
	}
	if params.BoarID == uuid.Nil {
		return nil, ErrInvalidMountBoar
	}
	if params.OperatorID == uuid.Nil {
		return nil, ErrInvalidMountOperator
	}
	if mountNumber <= 0 {
		return nil, ErrInvalidMountNumber
	}
	if params.MountDate.IsZero() {
		return nil, ErrInvalidMountDate
	}

	mountType := params.Type
	if mountType == "" {
		mountType = MountTypeArtificial
	}
	if !isValidMountType(mountType) {
		return nil, ErrInvalidMountType
	}

	note, err := validateMountNote(params.Note)
	if err != nil {
		return nil, err
	}

	if createdBy == uuid.Nil {
		return nil, ErrInvalidMountCreatedBy
	}

	return &Mount{
		ID:          params.ID,
		ServiceID:   serviceID,
		BoarID:      params.BoarID,
		OperatorID:  params.OperatorID,
		MountNumber: mountNumber,
		MountDate:   params.MountDate,
		Type:        mountType,
		Note:        note,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}, nil
}

// validateMountNote trims the optional note and checks its length.
// It returns nil when note is nil or empty and ErrInvalidMountNote when too long.
func validateMountNote(note *string) (*string, error) {
	if note == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*note)
	if trimmed == "" {
		return nil, nil
	}
	if len([]rune(trimmed)) > maxMountNoteLength {
		return nil, ErrInvalidMountNote
	}
	return &trimmed, nil
}

// isValidMountType checks whether the type belongs to the mount domain.
// Only the two types defined by the mount_type enum are accepted.
func isValidMountType(mountType MountType) bool {
	switch mountType {
	case MountTypeNatural, MountTypeArtificial:
		return true
	default:
		return false
	}
}

// ParseMountType trims and validates a raw mount type value.
// It returns ErrInvalidMountType when the value is empty or unknown.
func ParseMountType(value string) (MountType, error) {
	mountType := MountType(strings.TrimSpace(value))
	if !isValidMountType(mountType) {
		return "", ErrInvalidMountType
	}
	return mountType, nil
}
