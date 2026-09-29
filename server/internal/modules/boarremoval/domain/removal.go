package boarremoval

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxNoteLength = 500

const stateAlive = "Vivo"

// Type represents the kind of removal applied to a boar.
// Its values match the removal_type database enum shared with the sow module.
type Type string

const (
	TypeDeath     Type = "Muerte"
	TypeDiscard   Type = "Desecho"
	TypeSacrifice Type = "Sacrificio"
)

// Reason represents the reason that motivated a boar removal.
// Its values match the removal_reason database enum shared with the sow module.
type Reason string

const (
	ReasonAgeParity           Reason = "Edad_Paridad"
	ReasonReproductiveFailure Reason = "Fallo_Reproductivo"
	ReasonLowProductivity     Reason = "Baja_Productividad"
	ReasonLocomotorProblem    Reason = "Problema_Locomotor"
	ReasonDisease             Reason = "Enfermedad"
	ReasonSuddenDeath         Reason = "Muerte_Subita"
	ReasonOther               Reason = "Otro"
)

var (
	ErrInvalidID              = errors.New("El identificador de la baja es obligatorio")
	ErrInvalidBoar            = errors.New("El verraco de la baja es obligatorio")
	ErrInvalidRemovalDate     = errors.New("La fecha de la baja es obligatoria")
	ErrRemovalDateInFuture    = errors.New("La fecha de la baja no puede ser futura")
	ErrInvalidType            = errors.New("El tipo de baja no es válido")
	ErrInvalidReason          = errors.New("El motivo de baja no es válido")
	ErrInvalidNote            = errors.New("La nota debe tener como máximo 500 caracteres")
	ErrInvalidUpdate          = errors.New("Debe actualizar al menos un campo de la baja")
	ErrInvalidCreatedBy       = errors.New("El usuario que crea la baja es obligatorio")
	ErrInvalidUpdatedBy       = errors.New("El usuario que actualiza la baja es obligatorio")
	ErrBoarNotRemovable       = errors.New("El verraco no está en un estado válido para la baja")
	ErrRemovalDateBeforeEntry = errors.New("La fecha de la baja no puede ser anterior a la fecha de ingreso del verraco")
	ErrRemovalDateBeforeMount = errors.New("La fecha de la baja no puede ser anterior a la última monta del verraco")
)

// BoarRemoval is the aggregate root that represents a boar removal (baja).
// It records the type, the reason, the date and the previous boar state so the
// removal can be reverted when the record is deleted.
type BoarRemoval struct {
	ID          uuid.UUID
	BoarID      uuid.UUID
	RemovalDate time.Time
	Type        Type
	Reason      Reason
	Note        *string
	LastState   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
}

// NewBoarRemovalParams holds the fields required to build a new removal.
// LastState is the boar state before the removal and the optional note may be
// nil; the audit fields are derived from the creator inside the constructor.
type NewBoarRemovalParams struct {
	ID          uuid.UUID
	BoarID      uuid.UUID
	RemovalDate time.Time
	Type        Type
	Reason      Reason
	Note        *string
	LastState   string
	CreatedBy   uuid.UUID
}

// UpdateBoarRemovalParams holds the mutable fields of a removal for an update.
// A nil pointer means the field is omitted and stays unchanged; boar_id, last
// state and the audit fields are not mutable by design.
type UpdateBoarRemovalParams struct {
	RemovalDate *time.Time
	Type        *Type
	Reason      *Reason
	Note        *string
}

// Reference groups the dates used to validate a boar removal.
// A zero date means the reference does not exist and its check is skipped.
type Reference struct {
	EntryDate     time.Time
	LastMountDate time.Time
}

// NewBoarRemoval builds a removal after validating its fields and dates.
// It requires an alive boar, a non-future removal date not before the boar entry
// date nor its last mount date, a valid type, reason, note and creator.
func NewBoarRemoval(params NewBoarRemovalParams, reference Reference, now time.Time) (*BoarRemoval, error) {
	if params.ID == uuid.Nil {
		return nil, ErrInvalidID
	}
	if params.BoarID == uuid.Nil {
		return nil, ErrInvalidBoar
	}
	if params.RemovalDate.IsZero() {
		return nil, ErrInvalidRemovalDate
	}
	if !isRemovableState(params.LastState) {
		return nil, ErrBoarNotRemovable
	}
	if err := validateRemovalDate(params.RemovalDate, reference, now); err != nil {
		return nil, err
	}
	if !isValidType(params.Type) {
		return nil, ErrInvalidType
	}
	if !isValidReason(params.Reason) {
		return nil, ErrInvalidReason
	}

	note, err := validateNote(params.Note)
	if err != nil {
		return nil, err
	}

	if params.CreatedBy == uuid.Nil {
		return nil, ErrInvalidCreatedBy
	}

	return &BoarRemoval{
		ID:          params.ID,
		BoarID:      params.BoarID,
		RemovalDate: params.RemovalDate,
		Type:        params.Type,
		Reason:      params.Reason,
		Note:        note,
		LastState:   params.LastState,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   params.CreatedBy,
		UpdatedBy:   params.CreatedBy,
	}, nil
}

// Update applies only the non-nil fields of the params to the removal.
// It re-validates the given values and, when the date is provided, the date
// rules against the reference, then records updatedBy plus the timestamp.
func (r *BoarRemoval) Update(params UpdateBoarRemovalParams, reference Reference, updatedBy uuid.UUID, now time.Time) error {
	if r == nil || r.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if params.RemovalDate == nil && params.Type == nil && params.Reason == nil && params.Note == nil {
		return ErrInvalidUpdate
	}

	validatedDate := r.RemovalDate
	if params.RemovalDate != nil {
		if params.RemovalDate.IsZero() {
			return ErrInvalidRemovalDate
		}
		if err := validateRemovalDate(*params.RemovalDate, reference, now); err != nil {
			return err
		}
		validatedDate = *params.RemovalDate
	}

	validatedType := r.Type
	if params.Type != nil {
		if !isValidType(*params.Type) {
			return ErrInvalidType
		}
		validatedType = *params.Type
	}

	validatedReason := r.Reason
	if params.Reason != nil {
		if !isValidReason(*params.Reason) {
			return ErrInvalidReason
		}
		validatedReason = *params.Reason
	}

	validatedNote := r.Note
	if params.Note != nil {
		var err error
		validatedNote, err = validateNote(params.Note)
		if err != nil {
			return err
		}
	}

	r.RemovalDate = validatedDate
	r.Type = validatedType
	r.Reason = validatedReason
	r.Note = validatedNote
	r.UpdatedAt = now
	r.UpdatedBy = updatedBy
	return nil
}

// validateRemovalDate checks the removal date against the current day, the boar
// entry date and its last mount date.
func validateRemovalDate(removalDate time.Time, reference Reference, now time.Time) error {
	if truncateToDay(removalDate).After(truncateToDay(now)) {
		return ErrRemovalDateInFuture
	}

	if !reference.EntryDate.IsZero() && truncateToDay(removalDate).Before(truncateToDay(reference.EntryDate)) {
		return ErrRemovalDateBeforeEntry
	}

	if !reference.LastMountDate.IsZero() && truncateToDay(removalDate).Before(truncateToDay(reference.LastMountDate)) {
		return ErrRemovalDateBeforeMount
	}
	return nil
}

// truncateToDay removes the time portion from a timestamp in UTC.
// It is used to compare removal and reference dates at day granularity.
func truncateToDay(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

// validateNote trims the optional note and checks its length.
// It returns nil when note is nil or empty (clear to NULL) and ErrInvalidNote
// when it is too long.
func validateNote(note *string) (*string, error) {
	if note == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*note)
	if trimmed == "" {
		return nil, nil
	}
	if len([]rune(trimmed)) > maxNoteLength {
		return nil, ErrInvalidNote
	}
	return &trimmed, nil
}

// isRemovableState reports whether a boar state allows registering a removal.
// Only alive boars may be removed.
func isRemovableState(lastState string) bool {
	return lastState == stateAlive
}

// isValidType checks whether the type belongs to the removal domain.
// Only the three types defined by the removal_type enum are accepted.
func isValidType(removalType Type) bool {
	switch removalType {
	case TypeDeath, TypeDiscard, TypeSacrifice:
		return true
	default:
		return false
	}
}

// isValidReason checks whether the reason belongs to the removal domain.
// Only the seven reasons defined by the removal_reason enum are accepted.
func isValidReason(reason Reason) bool {
	switch reason {
	case ReasonAgeParity, ReasonReproductiveFailure, ReasonLowProductivity,
		ReasonLocomotorProblem, ReasonDisease, ReasonSuddenDeath, ReasonOther:
		return true
	default:
		return false
	}
}

// BoarState maps a removal type to the boar state the boar adopts after removal.
// It returns ErrInvalidType when the type is unknown.
func (t Type) BoarState() (string, error) {
	switch t {
	case TypeDeath:
		return "Muerto", nil
	case TypeDiscard:
		return "Desecho", nil
	case TypeSacrifice:
		return "Sacrificado", nil
	default:
		return "", ErrInvalidType
	}
}

// ParseType trims and validates a raw removal type value.
// It returns ErrInvalidType when the value is empty or unknown.
func ParseType(value string) (Type, error) {
	removalType := Type(strings.TrimSpace(value))
	if !isValidType(removalType) {
		return "", ErrInvalidType
	}
	return removalType, nil
}

// ParseReason trims and validates a raw removal reason value.
// It returns ErrInvalidReason when the value is empty or unknown.
func ParseReason(value string) (Reason, error) {
	reason := Reason(strings.TrimSpace(value))
	if !isValidReason(reason) {
		return "", ErrInvalidReason
	}
	return reason, nil
}
