package sowremoval

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxNoteLength = 500

const (
	stateAlive    = "Viva"
	stateWeaned   = "Destetada"
	stateAborted  = "Abortada"
	statePregnant = "Gestando"
)

// Type represents the kind of removal applied to a sow.
// Its values match the removal_type database enum.
type Type string

const (
	TypeDeath     Type = "Muerte"
	TypeDiscard   Type = "Desecho"
	TypeSacrifice Type = "Sacrificio"
)

// Reason represents the reason that motivated a sow removal.
// Its values match the removal_reason database enum.
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
	ErrInvalidID                 = errors.New("El identificador de la baja es obligatorio")
	ErrInvalidSow                = errors.New("La cerda de la baja es obligatoria")
	ErrInvalidRemovalDate        = errors.New("La fecha de la baja es obligatoria")
	ErrRemovalDateInFuture       = errors.New("La fecha de la baja no puede ser futura")
	ErrInvalidType               = errors.New("El tipo de baja no es válido")
	ErrInvalidReason             = errors.New("El motivo de baja no es válido")
	ErrInvalidNote               = errors.New("La nota debe tener como máximo 500 caracteres")
	ErrInvalidCreatedBy          = errors.New("El usuario que crea la baja es obligatorio")
	ErrInvalidUpdatedBy          = errors.New("El usuario que actualiza la baja es obligatorio")
	ErrSowNotRemovable           = errors.New("La cerda no está en un estado válido para la baja")
	ErrRemovalDateBeforeWeaning  = errors.New("La fecha de la baja debe ser posterior al destete")
	ErrRemovalDateBeforeService  = errors.New("La fecha de la baja debe ser posterior al último servicio")
	ErrRemovalDateBeforeAbortion = errors.New("La fecha de la baja debe ser posterior al último aborto")
)

// SowRemoval is the aggregate root that represents a sow removal (baja).
// It records the type, the reason, the date and the previous sow state so the
// removal can be reverted when the record is deleted.
type SowRemoval struct {
	ID          uuid.UUID
	SowID       uuid.UUID
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

// NewSowRemovalParams holds the fields required to build a new removal.
// LastState is the sow state before the removal and the optional note may be
// nil; the audit fields are derived from the creator inside the constructor.
type NewSowRemovalParams struct {
	ID          uuid.UUID
	SowID       uuid.UUID
	RemovalDate time.Time
	Type        Type
	Reason      Reason
	Note        *string
	LastState   string
	CreatedBy   uuid.UUID
}

// Reference groups the dates used to validate a removal by sow state.
// A zero date means the reference does not exist and its check is skipped.
type Reference struct {
	LastMountDate    time.Time
	LastAbortionDate time.Time
	WeaningDate      time.Time
}

// NewSowRemoval builds a removal after validating its fields and dates.
// It requires a removable last state, a non-future removal date and, depending
// on the state, a date after the last mount or the last abortion.
func NewSowRemoval(params NewSowRemovalParams, reference Reference, now time.Time) (*SowRemoval, error) {
	if params.ID == uuid.Nil {
		return nil, ErrInvalidID
	}
	if params.SowID == uuid.Nil {
		return nil, ErrInvalidSow
	}
	if params.RemovalDate.IsZero() {
		return nil, ErrInvalidRemovalDate
	}
	if !isRemovableState(params.LastState) {
		return nil, ErrSowNotRemovable
	}
	if err := validateRemovalDate(params.RemovalDate, params.LastState, reference, now); err != nil {
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

	return &SowRemoval{
		ID:          params.ID,
		SowID:       params.SowID,
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

// validateRemovalDate checks the removal date against the current day and the
// reference dates required by the sow state that is being removed.
func validateRemovalDate(removalDate time.Time, lastState string, reference Reference, now time.Time) error {
	if truncateToDay(removalDate).After(truncateToDay(now)) {
		return ErrRemovalDateInFuture
	}

	switch lastState {
	case stateAlive:
		return nil
	case stateWeaned:
		// TODO: validate the removal date against the weaning date once the
		// weaning module exists; today the destete date is not available.
		if !reference.WeaningDate.IsZero() && !isAfter(removalDate, reference.WeaningDate) {
			return ErrRemovalDateBeforeWeaning
		}
	case stateAborted:
		if !reference.LastAbortionDate.IsZero() && !isAfter(removalDate, reference.LastAbortionDate) {
			return ErrRemovalDateBeforeAbortion
		}
	case statePregnant:
		if !reference.LastMountDate.IsZero() && !isAfter(removalDate, reference.LastMountDate) {
			return ErrRemovalDateBeforeService
		}
		if !reference.LastAbortionDate.IsZero() && !isAfter(removalDate, reference.LastAbortionDate) {
			return ErrRemovalDateBeforeAbortion
		}
	}
	return nil
}

// truncateToDay removes the time portion from a timestamp in UTC.
// It is used to compare removal and reference dates at day granularity.
func truncateToDay(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

// isAfter reports whether a date is strictly later than a reference date.
// Both values are compared at day granularity to ignore the time portion.
func isAfter(date time.Time, reference time.Time) bool {
	return truncateToDay(date).After(truncateToDay(reference))
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

// isRemovableState reports whether a sow state allows registering a removal.
// Only alive, weaned, aborted and gestating sows may be removed.
func isRemovableState(lastState string) bool {
	switch lastState {
	case stateAlive, stateWeaned, stateAborted, statePregnant:
		return true
	default:
		return false
	}
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

// SowState maps a removal type to the sow state the sow adopts after removal.
// It returns ErrInvalidType when the type is unknown.
func (t Type) SowState() (string, error) {
	switch t {
	case TypeDeath:
		return "Muerta", nil
	case TypeDiscard:
		return "Desecho", nil
	case TypeSacrifice:
		return "Sacrificada", nil
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
