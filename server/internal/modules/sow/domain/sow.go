package sow

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	maxCodeLength     = 50
	maxLocationLength = 100
	maxNoteLength     = 500
)

// State represents the life state of a sow.
// Its values match the Sow_state database enum.
type State string

const (
	StateAlive      State = "Viva"
	StateDead       State = "Muerta"
	StateDiscarded  State = "Desecho"
	StateSacrificed State = "Sacrificada"
	StateAborted    State = "Abortada"
	StatePregnant   State = "Gestando"
	StateLactating  State = "Lactando"
	StateWeaned     State = "Destetada"
)

// Origin represents where a sow comes from.
// Its values match the boar_origin database enum shared with the boar module.
type Origin string

const (
	OriginOwn      Origin = "Propio"
	OriginExternal Origin = "Externo"
)

var (
	ErrInvalidID        = errors.New("El identificador de la cerda es obligatorio")
	ErrInvalidCode      = errors.New("El código es obligatorio y debe tener como máximo 50 caracteres")
	ErrInvalidLocation  = errors.New("La ubicación debe tener como máximo 100 caracteres")
	ErrInvalidEntryDate = errors.New("La fecha de ingreso es obligatoria")
	ErrInvalidBirthDate = errors.New("La fecha de nacimiento no puede ser posterior a la fecha de ingreso")
	ErrInvalidNote      = errors.New("La nota debe tener como máximo 500 caracteres")
	ErrInvalidState     = errors.New("El estado de la cerda no es válido")
	ErrInvalidOrigin    = errors.New("El origen de la cerda no es válido")
	ErrInvalidBreed     = errors.New("La raza es obligatoria")
	ErrInvalidParity    = errors.New("La paridad no puede ser negativa")
	ErrInvalidUpdate    = errors.New("Debe actualizar al menos un campo de la cerda")
	ErrInvalidCreatedBy = errors.New("El usuario que crea la cerda es obligatorio")
	ErrInvalidUpdatedBy = errors.New("El usuario que actualiza la cerda es obligatorio")
)

// Sow is the aggregate root that represents a sow (cerda) in the farm.
// It carries the audit fields, its parity and the breed it belongs to.
type Sow struct {
	ID        uuid.UUID
	Code      string
	Location  *string
	Active    bool
	EntryDate time.Time
	BirthDate *time.Time
	Note      *string
	State     State
	Origin    Origin
	Parity    int
	BreedID   uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

// SowOption is a lightweight sow read model for selection lists.
// It only carries the identifier and code of a sow.
type SowOption struct {
	ID   uuid.UUID
	Code string
}

// NewSowParams holds the fields required to build a new sow.
// Optional values are pointers and may be nil; the state is not part of the
// params because every new sow starts as StateAlive.
type NewSowParams struct {
	ID        uuid.UUID
	Code      string
	Location  *string
	EntryDate time.Time
	BirthDate *time.Time
	Note      *string
	Origin    Origin
	Parity    int
	BreedID   uuid.UUID
	CreatedBy uuid.UUID
}

// UpdateSowParams holds the mutable fields of a sow for an update.
// A nil pointer means the field is omitted and stays unchanged; nullable string
// fields are cleared with an empty string and the nullable date uses
// ClearBirthDate. Parity and state are intentionally absent because parity is
// set only at creation and only ChangeState may alter the state.
type UpdateSowParams struct {
	Code           *string
	Location       *string
	Active         *bool
	EntryDate      *time.Time
	BirthDate      *time.Time
	ClearBirthDate bool
	Note           *string
	Origin         *Origin
	BreedID        *uuid.UUID
}

// NewSow builds a sow after validating its fields.
// It defaults Active to true and the state to StateAlive, then sets the audit
// fields to the creator and both timestamps to now.
func NewSow(params NewSowParams, now time.Time) (*Sow, error) {
	if params.ID == uuid.Nil {
		return nil, ErrInvalidID
	}

	code, err := validateCode(params.Code)
	if err != nil {
		return nil, err
	}

	location, err := validateLocation(params.Location)
	if err != nil {
		return nil, err
	}

	if params.EntryDate.IsZero() {
		return nil, ErrInvalidEntryDate
	}

	birthDate, err := validateBirthDate(params.BirthDate, params.EntryDate)
	if err != nil {
		return nil, err
	}

	note, err := validateNote(params.Note)
	if err != nil {
		return nil, err
	}

	if !isValidOrigin(params.Origin) {
		return nil, ErrInvalidOrigin
	}

	if params.Parity < 0 {
		return nil, ErrInvalidParity
	}

	if params.BreedID == uuid.Nil {
		return nil, ErrInvalidBreed
	}

	if params.CreatedBy == uuid.Nil {
		return nil, ErrInvalidCreatedBy
	}

	return &Sow{
		ID:        params.ID,
		Code:      code,
		Location:  location,
		Active:    true,
		EntryDate: params.EntryDate,
		BirthDate: birthDate,
		Note:      note,
		State:     StateAlive,
		Origin:    params.Origin,
		Parity:    params.Parity,
		BreedID:   params.BreedID,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: params.CreatedBy,
		UpdatedBy: params.CreatedBy,
	}, nil
}

// Update applies only the non-nil fields of the given params to the sow.
// It validates each provided value and records updatedBy plus the timestamp,
// returning ErrInvalidUpdate when no field is provided; it never touches state
// or parity.
func (s *Sow) Update(params UpdateSowParams, updatedBy uuid.UUID, now time.Time) error {
	if s == nil || s.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if params.Code == nil && params.Location == nil && params.Active == nil &&
		params.EntryDate == nil && params.BirthDate == nil && !params.ClearBirthDate &&
		params.Note == nil && params.Origin == nil && params.BreedID == nil {
		return ErrInvalidUpdate
	}

	validatedCode := s.Code
	if params.Code != nil {
		var err error
		validatedCode, err = validateCode(*params.Code)
		if err != nil {
			return err
		}
	}

	validatedLocation := s.Location
	if params.Location != nil {
		var err error
		validatedLocation, err = validateLocation(params.Location)
		if err != nil {
			return err
		}
	}

	validatedEntryDate := s.EntryDate
	if params.EntryDate != nil {
		if params.EntryDate.IsZero() {
			return ErrInvalidEntryDate
		}
		validatedEntryDate = *params.EntryDate
	}

	validatedBirthDate := s.BirthDate
	switch {
	case params.BirthDate != nil:
		var err error
		validatedBirthDate, err = validateBirthDate(params.BirthDate, validatedEntryDate)
		if err != nil {
			return err
		}
	case params.ClearBirthDate:
		validatedBirthDate = nil
	case params.EntryDate != nil && validatedBirthDate != nil && validatedBirthDate.After(validatedEntryDate):
		return ErrInvalidBirthDate
	}

	validatedNote := s.Note
	if params.Note != nil {
		var err error
		validatedNote, err = validateNote(params.Note)
		if err != nil {
			return err
		}
	}

	validatedOrigin := s.Origin
	if params.Origin != nil {
		if !isValidOrigin(*params.Origin) {
			return ErrInvalidOrigin
		}
		validatedOrigin = *params.Origin
	}

	validatedBreedID := s.BreedID
	if params.BreedID != nil {
		if *params.BreedID == uuid.Nil {
			return ErrInvalidBreed
		}
		validatedBreedID = *params.BreedID
	}

	s.Code = validatedCode
	s.Location = validatedLocation
	if params.Active != nil {
		s.Active = *params.Active
	}
	s.EntryDate = validatedEntryDate
	s.BirthDate = validatedBirthDate
	s.Note = validatedNote
	s.Origin = validatedOrigin
	s.BreedID = validatedBreedID
	s.UpdatedAt = now
	s.UpdatedBy = updatedBy
	return nil
}

// ChangeState validates the target state and applies it to the sow.
// It is the only way to alter State and is reserved for domain event handlers,
// never for the user-facing update flow; it records updatedBy plus the timestamp.
func (s *Sow) ChangeState(newState State, updatedBy uuid.UUID, now time.Time) error {
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

// validateCode trims the code and requires it to be non-empty.
// It returns ErrInvalidCode when the code exceeds the maximum length.
func validateCode(code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" || len([]rune(code)) > maxCodeLength {
		return "", ErrInvalidCode
	}
	return code, nil
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

// validateBirthDate validates the optional birth date against the entry date.
// It returns nil when birthDate is nil and ErrInvalidBirthDate when it is
// zero or later than the entry date.
func validateBirthDate(birthDate *time.Time, entryDate time.Time) (*time.Time, error) {
	if birthDate == nil {
		return nil, nil
	}
	if birthDate.IsZero() || birthDate.After(entryDate) {
		return nil, ErrInvalidBirthDate
	}
	return birthDate, nil
}

// isValidState checks whether the state belongs to the sow domain.
// Only the eight states defined by the Sow_state enum are accepted.
func isValidState(state State) bool {
	switch state {
	case StateAlive, StateDead, StateDiscarded, StateSacrificed,
		StateAborted, StatePregnant, StateLactating, StateWeaned:
		return true
	default:
		return false
	}
}

// isValidOrigin checks whether the origin belongs to the sow domain.
// Only the two origins defined by the boar_origin enum are accepted.
func isValidOrigin(origin Origin) bool {
	switch origin {
	case OriginOwn, OriginExternal:
		return true
	default:
		return false
	}
}

// ParseOrigin trims and validates a raw origin value.
// It returns ErrInvalidOrigin when the value is empty or unknown.
func ParseOrigin(value string) (Origin, error) {
	origin := Origin(strings.TrimSpace(value))
	if !isValidOrigin(origin) {
		return "", ErrInvalidOrigin
	}
	return origin, nil
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
