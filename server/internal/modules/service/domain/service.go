package service

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	maxLocationLength = 100
	maxNoteLength     = 500
	minMounts         = 1
	maxMounts         = 3
	gestationDays     = 114
	maxMountGap       = 24 * time.Hour
)

// State represents the breeding state of a service.
// Its values match the service_state database enum.
type State string

const (
	StateConfirmed State = "Confirmado"
	StateFailed    State = "Fallido"
	StateAborted   State = "Aborto"
	StateFinished  State = "Terminado"
)

var (
	ErrInvalidID              = errors.New("El identificador del servicio es obligatorio")
	ErrInvalidSow             = errors.New("La cerda del servicio es obligatoria")
	ErrInvalidLocation        = errors.New("La ubicación debe tener como máximo 100 caracteres")
	ErrInvalidNote            = errors.New("La nota debe tener como máximo 500 caracteres")
	ErrInvalidState           = errors.New("El estado del servicio no es válido")
	ErrInvalidCreatedBy       = errors.New("El usuario que crea el servicio es obligatorio")
	ErrInvalidUpdatedBy       = errors.New("El usuario que actualiza el servicio es obligatorio")
	ErrInvalidMounts          = errors.New("El servicio debe tener entre 1 y 3 montas")
	ErrMountDatesNotAscending = errors.New("Las fechas de monta deben estar en orden ascendente")
	ErrMountDateGapTooLarge   = errors.New("Las montas no pueden tener más de 24 horas de diferencia")
	ErrMountDateInFuture      = errors.New("La fecha de monta no puede ser futura")
)

// Service is the aggregate root that represents a sow service (servicio).
// It groups between one and three mounts and computes the expected farrowing date.
type Service struct {
	ID                    uuid.UUID
	SowID                 uuid.UUID
	ExpectedFarrowingDate *time.Time
	Note                  *string
	State                 State
	Location              *string
	Mounts                []*Mount
	CreatedAt             time.Time
	UpdatedAt             time.Time
	CreatedBy             uuid.UUID
	UpdatedBy             uuid.UUID
}

// NewServiceParams holds the fields required to build a new service.
// Optional values are pointers and may be nil; the state and the expected
// farrowing date are derived by the domain, not provided by the caller.
type NewServiceParams struct {
	ID        uuid.UUID
	SowID     uuid.UUID
	Note      *string
	Location  *string
	CreatedBy uuid.UUID
}

// NewService builds a service with its mounts after validating every field.
// It requires between one and three mounts ordered by date with a maximum gap
// of 24 hours, rejects future dates, defaults the state to StateConfirmed and
// computes the expected farrowing date as the last mount plus 114 days.
func NewService(params NewServiceParams, mounts []NewMountParams, now time.Time) (*Service, error) {
	if params.ID == uuid.Nil {
		return nil, ErrInvalidID
	}
	if params.SowID == uuid.Nil {
		return nil, ErrInvalidSow
	}

	location, err := validateLocation(params.Location)
	if err != nil {
		return nil, err
	}

	note, err := validateNote(params.Note)
	if err != nil {
		return nil, err
	}

	if params.CreatedBy == uuid.Nil {
		return nil, ErrInvalidCreatedBy
	}

	if len(mounts) < minMounts || len(mounts) > maxMounts {
		return nil, ErrInvalidMounts
	}

	if err := validateMountSchedule(mounts, now); err != nil {
		return nil, err
	}

	builtMounts := make([]*Mount, 0, len(mounts))
	for index, mountParams := range mounts {
		mount, err := NewMount(mountParams, params.ID, index+1, params.CreatedBy, now)
		if err != nil {
			return nil, err
		}
		builtMounts = append(builtMounts, mount)
	}

	expectedFarrowingDate := builtMounts[len(builtMounts)-1].MountDate.AddDate(0, 0, gestationDays)

	return &Service{
		ID:                    params.ID,
		SowID:                 params.SowID,
		ExpectedFarrowingDate: &expectedFarrowingDate,
		Note:                  note,
		State:                 StateConfirmed,
		Location:              location,
		Mounts:                builtMounts,
		CreatedAt:             now,
		UpdatedAt:             now,
		CreatedBy:             params.CreatedBy,
		UpdatedBy:             params.CreatedBy,
	}, nil
}

// validateMountSchedule validates the date order, gap and future bounds.
// It returns ErrMountDateInFuture, ErrMountDatesNotAscending or
// ErrMountDateGapTooLarge when the mounts do not follow the schedule rules.
func validateMountSchedule(mounts []NewMountParams, now time.Time) error {
	today := truncateToDay(now)
	for index, mount := range mounts {
		if truncateToDay(mount.MountDate).After(today) {
			return ErrMountDateInFuture
		}
		if index == 0 {
			continue
		}
		previous := mounts[index-1].MountDate
		if mount.MountDate.Before(previous) {
			return ErrMountDatesNotAscending
		}
		if mount.MountDate.Sub(previous) > maxMountGap {
			return ErrMountDateGapTooLarge
		}
	}
	return nil
}

// truncateToDay removes the time portion from a timestamp in UTC.
// It is used to compare mount dates at day granularity.
func truncateToDay(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

// validateLocation trims the optional location and checks its length.
// It returns nil when location is nil or empty (clear to NULL) and
// ErrInvalidLocation when it is too long.
func validateLocation(location *string) (*string, error) {
	if location == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*location)
	if trimmed == "" {
		return nil, nil
	}
	if len([]rune(trimmed)) > maxLocationLength {
		return nil, ErrInvalidLocation
	}
	return &trimmed, nil
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

// ChangeState validates the target state and applies it to the service.
// It is reserved for domain event handlers such as the abortion flow and
// records updatedBy plus the timestamp, never touching the mounts.
func (s *Service) ChangeState(newState State, updatedBy uuid.UUID, now time.Time) error {
	if s == nil || s.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if !isValidState(newState) {
		return ErrInvalidState
	}

	s.State = newState
	s.UpdatedAt = now
	s.UpdatedBy = updatedBy
	return nil
}

// isValidState checks whether the state belongs to the service domain.
// Only the four states defined by the service_state enum are accepted.
func isValidState(state State) bool {
	switch state {
	case StateConfirmed, StateFailed, StateAborted, StateFinished:
		return true
	default:
		return false
	}
}

// ParseState trims and validates a raw state value.
// It returns ErrInvalidState when the value is empty or unknown.
func ParseState(value string) (State, error) {
	state := State(strings.TrimSpace(value))
	if !isValidState(state) {
		return "", ErrInvalidState
	}
	return state, nil
}
