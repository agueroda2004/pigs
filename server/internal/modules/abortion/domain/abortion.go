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
	ErrInvalidCreatedBy        = errors.New("El usuario que crea el aborto es obligatorio")
	ErrInvalidUpdatedBy        = errors.New("El usuario que actualiza el aborto es obligatorio")
	ErrAbortionDateBeforeMount = errors.New("La fecha del aborto debe ser posterior a la última monta")
	ErrAbortionDateInFuture    = errors.New("La fecha del aborto no puede ser futura")
)

// Abortion is the aggregate root that represents a sow abortion (aborto).
// It records the cause, the date and the service that produced the gestation.
type Abortion struct {
	ID           uuid.UUID
	SowID        uuid.UUID
	ServiceID    uuid.UUID
	AbortionDate time.Time
	Cause        Cause
	Note         *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    uuid.UUID
	UpdatedBy    uuid.UUID
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

// NewAbortion builds an abortion after validating its fields and its dates.
// It requires the abortion date to be strictly after lastMountDate and not
// later than today, then sets the audit fields to the creator and now.
func NewAbortion(params NewAbortionParams, lastMountDate time.Time, now time.Time) (*Abortion, error) {
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
	if lastMountDate.IsZero() {
		return nil, ErrInvalidLastMount
	}

	if err := validateAbortionDate(params.AbortionDate, lastMountDate, now); err != nil {
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

// validateAbortionDate checks the abortion date bounds against the last mount.
// It rejects dates after today and dates on or before the last mount date.
func validateAbortionDate(abortionDate time.Time, lastMountDate time.Time, now time.Time) error {
	abortion := truncateToDay(abortionDate)
	if abortion.After(truncateToDay(now)) {
		return ErrAbortionDateInFuture
	}
	if !abortion.After(truncateToDay(lastMountDate)) {
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
