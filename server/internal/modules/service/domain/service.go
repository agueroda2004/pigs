package service

import (
	"errors"
	"sort"
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

// Sow states accepted as the previous state of a service. They mirror the
// serviceable states of the sow domain and the sow_state database enum.
const (
	sowStateAlive    = "Viva"
	sowStateWeaned   = "Destetada"
	sowStateAborted  = "Abortada"
	sowStatePregnant = "Gestando"
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
	ErrInvalidLastState       = errors.New("El estado previo de la cerda no es válido")
	// ErrServiceNotDeletable is returned when a service cannot be deleted because
	// its state is not Confirmado, meaning it already has related events.
	ErrServiceNotDeletable = errors.New("Solo se pueden eliminar servicios en estado Confirmado")
	// ErrServiceNotEditable is returned when a service cannot be edited because
	// its state is not Confirmado, meaning it already has related events.
	ErrServiceNotEditable = errors.New("Solo se pueden editar servicios en estado Confirmado")
	// ErrMountBeforeEntryDate is returned when a mount date falls before the sow
	// entry date.
	ErrMountBeforeEntryDate = errors.New("La fecha de monta no puede ser anterior al ingreso de la cerda")
	// ErrMountBeforePreviousService is returned when a mount date falls on or
	// before the previous service's latest mount date.
	ErrMountBeforePreviousService = errors.New("La fecha de monta debe ser posterior al servicio anterior")
	// ErrMountBeforeAbortion is returned when a mount date falls on or before the
	// sow's last abortion date.
	ErrMountBeforeAbortion = errors.New("La fecha de monta debe ser posterior al último aborto")
	// ErrMountNotFound is returned when an update or delete references a mount
	// that does not belong to the service.
	ErrMountNotFound = errors.New("La monta no pertenece al servicio")
)

// Reference groups the dates a service's mounts must respect. A zero time means
// the reference does not exist and its check is skipped.
type Reference struct {
	EntryDate         time.Time
	PreviousMountDate time.Time
	LastAbortionDate  time.Time
}

// Service is the aggregate root that represents a sow service (servicio).
// It groups between one and three mounts and computes the expected farrowing date.
type Service struct {
	ID                    uuid.UUID
	SowID                 uuid.UUID
	ExpectedFarrowingDate *time.Time
	Note                  *string
	State                 State
	Location              *string
	// LastState is the sow state before the service, kept so it can be restored
	// when the service is deleted.
	LastState string
	// SowCode is the code of the referenced sow. It is a read-model field filled
	// only by the list query and never persisted by the write paths.
	SowCode   string
	Mounts    []*Mount
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

// NewServiceParams holds the fields required to build a new service.
// Optional values are pointers and may be nil; the state and the expected
// farrowing date are derived by the domain, not provided by the caller.
type NewServiceParams struct {
	ID        uuid.UUID
	SowID     uuid.UUID
	Note      *string
	Location  *string
	LastState string
	CreatedBy uuid.UUID
}

// UpdateMountParams holds the fields required to update a single mount.
// The identifier selects the mount; every other field replaces its current value.
type UpdateMountParams struct {
	ID         uuid.UUID
	BoarID     uuid.UUID
	OperatorID uuid.UUID
	MountDate  time.Time
	Type       MountType
	Note       *string
}

// UpdateServiceParams holds the editable fields of a service update.
// A nil location or note is left unchanged; the mounts are edited through three
// lists: those to create, those to update by identifier and those to delete.
type UpdateServiceParams struct {
	Location       *string
	Note           *string
	CreateMounts   []NewMountParams
	UpdateMounts   []UpdateMountParams
	DeleteMountIDs []uuid.UUID
}

// NewService builds a service with its mounts after validating every field.
// It requires between one and three mounts ordered by date with a maximum gap
// of 24 hours, rejects future dates and dates that break the reference bounds,
// defaults the state to StateConfirmed and computes the expected farrowing date
// as the last mount plus 114 days.
func NewService(params NewServiceParams, mounts []NewMountParams, reference Reference, now time.Time) (*Service, error) {
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

	if !isValidLastState(params.LastState) {
		return nil, ErrInvalidLastState
	}

	if len(mounts) < minMounts || len(mounts) > maxMounts {
		return nil, ErrInvalidMounts
	}

	mountDates := make([]time.Time, 0, len(mounts))
	for _, mount := range mounts {
		mountDates = append(mountDates, mount.MountDate)
	}
	if err := validateMountSchedule(mountDates, now); err != nil {
		return nil, err
	}
	if err := validateMountReference(mountDates, reference); err != nil {
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
		LastState:             params.LastState,
		Mounts:                builtMounts,
		CreatedAt:             now,
		UpdatedAt:             now,
		CreatedBy:             params.CreatedBy,
		UpdatedBy:             params.CreatedBy,
	}, nil
}

// isValidLastState reports whether a sow state may be stored as the previous
// state of a service. Only the serviceable states are accepted.
func isValidLastState(state string) bool {
	switch state {
	case sowStateAlive, sowStateWeaned, sowStateAborted, sowStatePregnant:
		return true
	default:
		return false
	}
}

// validateMountSchedule validates the date order, gap and future bounds.
// It returns ErrMountDateInFuture, ErrMountDatesNotAscending or
// ErrMountDateGapTooLarge when the mount dates do not follow the schedule rules.
func validateMountSchedule(dates []time.Time, now time.Time) error {
	today := truncateToDay(now)
	for index, date := range dates {
		if truncateToDay(date).After(today) {
			return ErrMountDateInFuture
		}
		if index == 0 {
			continue
		}
		previous := dates[index-1]
		if date.Before(previous) {
			return ErrMountDatesNotAscending
		}
		if date.Sub(previous) > maxMountGap {
			return ErrMountDateGapTooLarge
		}
	}
	return nil
}

// validateMountReference checks every mount date against the reference bounds.
// It requires a date not before the entry date and strictly after both the
// previous service's latest mount date and the last abortion date.
func validateMountReference(dates []time.Time, reference Reference) error {
	for _, date := range dates {
		day := truncateToDay(date)
		if !reference.EntryDate.IsZero() && day.Before(truncateToDay(reference.EntryDate)) {
			return ErrMountBeforeEntryDate
		}
		if !reference.PreviousMountDate.IsZero() && !day.After(truncateToDay(reference.PreviousMountDate)) {
			return ErrMountBeforePreviousService
		}
		if !reference.LastAbortionDate.IsZero() && !day.After(truncateToDay(reference.LastAbortionDate)) {
			return ErrMountBeforeAbortion
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

// EnsureDeletable checks whether the service may be deleted.
// Only a service in StateConfirmed can be deleted; any other state means the
// service already has related events and returns ErrServiceNotDeletable.
func (s *Service) EnsureDeletable() error {
	if s == nil || s.ID == uuid.Nil {
		return ErrInvalidID
	}
	if s.State != StateConfirmed {
		return ErrServiceNotDeletable
	}
	return nil
}

// EnsureEditable checks whether the service may be edited.
// Only a service in StateConfirmed can be edited; any other state means the
// service already has related events and returns ErrServiceNotEditable.
func (s *Service) EnsureEditable() error {
	if s == nil || s.ID == uuid.Nil {
		return ErrInvalidID
	}
	if s.State != StateConfirmed {
		return ErrServiceNotEditable
	}
	return nil
}

// Update applies the editable fields and the mount operations to the service.
// It requires StateConfirmed, validates the resulting mounts against the
// schedule and the reference dates, renumbers them by date and recomputes the
// expected farrowing date. It returns ErrMountNotFound for unknown identifiers.
func (s *Service) Update(params UpdateServiceParams, reference Reference, updatedBy uuid.UUID, now time.Time) error {
	if s == nil || s.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if err := s.EnsureEditable(); err != nil {
		return err
	}

	byID := make(map[uuid.UUID]*Mount, len(s.Mounts))
	for _, mount := range s.Mounts {
		byID[mount.ID] = mount
	}

	deleted := make(map[uuid.UUID]struct{}, len(params.DeleteMountIDs))
	for _, id := range params.DeleteMountIDs {
		if _, ok := byID[id]; !ok {
			return ErrMountNotFound
		}
		deleted[id] = struct{}{}
	}

	for _, update := range params.UpdateMounts {
		mount, ok := byID[update.ID]
		if !ok {
			return ErrMountNotFound
		}
		if err := mount.Update(update, updatedBy, now); err != nil {
			return err
		}
	}

	final := make([]*Mount, 0, len(s.Mounts)+len(params.CreateMounts))
	for _, mount := range s.Mounts {
		if _, ok := deleted[mount.ID]; ok {
			continue
		}
		final = append(final, mount)
	}
	for _, create := range params.CreateMounts {
		mount, err := NewMount(create, s.ID, len(final)+1, updatedBy, now)
		if err != nil {
			return err
		}
		final = append(final, mount)
	}

	if len(final) < minMounts || len(final) > maxMounts {
		return ErrInvalidMounts
	}

	sort.SliceStable(final, func(i, j int) bool {
		return final[i].MountDate.Before(final[j].MountDate)
	})

	dates := make([]time.Time, 0, len(final))
	for _, mount := range final {
		dates = append(dates, mount.MountDate)
	}
	if err := validateMountSchedule(dates, now); err != nil {
		return err
	}
	if err := validateMountReference(dates, reference); err != nil {
		return err
	}

	for index, mount := range final {
		mount.MountNumber = index + 1
	}
	expectedFarrowingDate := final[len(final)-1].MountDate.AddDate(0, 0, gestationDays)

	if params.Location != nil {
		location, err := validateLocation(params.Location)
		if err != nil {
			return err
		}
		s.Location = location
	}
	if params.Note != nil {
		note, err := validateNote(params.Note)
		if err != nil {
			return err
		}
		s.Note = note
	}

	s.Mounts = final
	s.ExpectedFarrowingDate = &expectedFarrowingDate
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
