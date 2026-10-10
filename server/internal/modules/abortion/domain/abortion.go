package abortion

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxNoteLength = 500

// Cause represents the reason of a sow abortion.
// Its values match the abortion_cause database enum.
type Cause string

const (
	CauseUnknown     Cause = "Desconocido"
	CauseInfectious  Cause = "Infeccioso"
	CauseTrauma      Cause = "Traumatismo"
	CauseManagement  Cause = "Manejo"
	CauseNutritional Cause = "Nutricional"
	CauseOther       Cause = "Otro"
)

var (
	ErrInvalidID               = errors.New("El identificador del aborto es obligatorio")
	ErrInvalidSow              = errors.New("La cerda del aborto es obligatoria")
	ErrInvalidService          = errors.New("El servicio del aborto es obligatorio")
	ErrInvalidAbortionDate     = errors.New("La fecha del aborto es obligatoria")
	ErrInvalidLastMount        = errors.New("La fecha de la última monta es obligatoria")
	ErrInvalidCause            = errors.New("La causa del aborto no es válida")
	ErrInvalidNote             = errors.New("La nota debe tener como máximo 500 caracteres")
	ErrInvalidUpdate           = errors.New("Debe actualizar al menos un campo del aborto")
	ErrInvalidCreatedBy        = errors.New("El usuario que crea el aborto es obligatorio")
	ErrInvalidUpdatedBy        = errors.New("El usuario que actualiza el aborto es obligatorio")
	ErrAbortionDateBeforeMount = errors.New("La fecha del aborto debe ser posterior a la última monta")
	ErrAbortionDateBeforeEntry = errors.New("La fecha del aborto no puede ser anterior al ingreso de la cerda")
	ErrAbortionDateInFuture    = errors.New("La fecha del aborto no puede ser futura")
)

// Reference groups the bounds an abortion date must respect. A zero time means
// the bound does not exist and its check is skipped.
type Reference struct {
	EntryDate     time.Time
	LastMountDate time.Time
}

// Abortion is the aggregate root that represents a sow abortion (aborto).
// It records the cause, the date and the service that produced the gestation.
type Abortion struct {
	ID           uuid.UUID
	SowID        uuid.UUID
	ServiceID    uuid.UUID
	AbortionDate time.Time
	Cause        Cause
	Note         *string
	// SowCode is the code of the referenced sow. It is a read-model field filled
	// only by the list query and never persisted by the write paths.
	SowCode   string
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

// NewAbortionParams holds the fields required to build a new abortion.
// The service is the sow's last service and the optional note may be nil; the
// audit fields are derived from the creator inside the constructor.
type NewAbortionParams struct {
	ID           uuid.UUID
	SowID        uuid.UUID
	ServiceID    uuid.UUID
	AbortionDate time.Time
	Cause        Cause
	Note         *string
	CreatedBy    uuid.UUID
}

// UpdateAbortionParams holds the editable fields of an abortion update.
// A nil pointer leaves the field unchanged; the sow and the service are never
// editable and are intentionally absent.
type UpdateAbortionParams struct {
	AbortionDate *time.Time
	Cause        *Cause
	Note         *string
}

// NewAbortion builds an abortion after validating its fields and its dates.
// It requires the abortion date to be strictly after the last mount, not before
// the sow entry date and not later than today, then sets the audit fields to the
// creator and now.
func NewAbortion(params NewAbortionParams, reference Reference, now time.Time) (*Abortion, error) {
	if params.ID == uuid.Nil {
		return nil, ErrInvalidID
	}
	if params.SowID == uuid.Nil {
		return nil, ErrInvalidSow
	}
	if params.ServiceID == uuid.Nil {
		return nil, ErrInvalidService
	}
	if params.AbortionDate.IsZero() {
		return nil, ErrInvalidAbortionDate
	}
	if reference.LastMountDate.IsZero() {
		return nil, ErrInvalidLastMount
	}

	if err := validateAbortionDate(params.AbortionDate, reference, now); err != nil {
		return nil, err
	}

	if !isValidCause(params.Cause) {
		return nil, ErrInvalidCause
	}

	note, err := validateNote(params.Note)
	if err != nil {
		return nil, err
	}

	if params.CreatedBy == uuid.Nil {
		return nil, ErrInvalidCreatedBy
	}

	return &Abortion{
		ID:           params.ID,
		SowID:        params.SowID,
		ServiceID:    params.ServiceID,
		AbortionDate: params.AbortionDate,
		Cause:        params.Cause,
		Note:         note,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    params.CreatedBy,
		UpdatedBy:    params.CreatedBy,
	}, nil
}

// Update applies the editable fields to the abortion after validating them.
// It never changes the sow or the service; when the date changes it validates it
// against the reference bounds and returns ErrInvalidUpdate when no field is given.
func (a *Abortion) Update(params UpdateAbortionParams, reference Reference, updatedBy uuid.UUID, now time.Time) error {
	if a == nil || a.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if params.AbortionDate == nil && params.Cause == nil && params.Note == nil {
		return ErrInvalidUpdate
	}

	if params.AbortionDate != nil {
		if params.AbortionDate.IsZero() {
			return ErrInvalidAbortionDate
		}
		if err := validateAbortionDate(*params.AbortionDate, reference, now); err != nil {
			return err
		}
		a.AbortionDate = *params.AbortionDate
	}

	if params.Cause != nil {
		if !isValidCause(*params.Cause) {
			return ErrInvalidCause
		}
		a.Cause = *params.Cause
	}

	if params.Note != nil {
		note, err := validateNote(params.Note)
		if err != nil {
			return err
		}
		a.Note = note
	}

	a.UpdatedAt = now
	a.UpdatedBy = updatedBy
	return nil
}

// validateAbortionDate checks the abortion date bounds against the reference.
// It rejects dates after today, dates before the entry date and dates on or
// before the last mount date. Zero reference bounds are skipped.
func validateAbortionDate(abortionDate time.Time, reference Reference, now time.Time) error {
	abortion := truncateToDay(abortionDate)
	if abortion.After(truncateToDay(now)) {
		return ErrAbortionDateInFuture
	}
	if !reference.EntryDate.IsZero() && abortion.Before(truncateToDay(reference.EntryDate)) {
		return ErrAbortionDateBeforeEntry
	}
	if !reference.LastMountDate.IsZero() && !abortion.After(truncateToDay(reference.LastMountDate)) {
		return ErrAbortionDateBeforeMount
	}
	return nil
}

// truncateToDay removes the time portion from a timestamp in UTC.
// It is used to compare the abortion and mount dates at day granularity.
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

// isValidCause checks whether the cause belongs to the abortion domain.
// Only the six causes defined by the abortion_cause enum are accepted.
func isValidCause(cause Cause) bool {
	switch cause {
	case CauseUnknown, CauseInfectious, CauseTrauma,
		CauseManagement, CauseNutritional, CauseOther:
		return true
	default:
		return false
	}
}

// ParseCause trims and validates a raw cause value.
// It returns ErrInvalidCause when the value is empty or unknown.
func ParseCause(value string) (Cause, error) {
	cause := Cause(strings.TrimSpace(value))
	if !isValidCause(cause) {
		return "", ErrInvalidCause
	}
	return cause, nil
}
