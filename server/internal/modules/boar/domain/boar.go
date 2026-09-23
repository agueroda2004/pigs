package boar

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

// State represents the life state of a boar.
// Its values match the Boar_state database enum.
type State string

const (
	StateAlive      State = "Vivo"
	StateDead       State = "Muerto"
	StateDiscarded  State = "Desecho"
	StateSacrificed State = "Sacrificado"
)

// Origin represents where a boar comes from.
// Its values match the boar_origin database enum.
type Origin string

const (
	OriginOwn      Origin = "Propio"
	OriginExternal Origin = "Externo"
)

var (
	ErrInvalidID        = errors.New("El identificador del verraco es obligatorio")
	ErrInvalidCode      = errors.New("El código es obligatorio y debe tener como máximo 50 caracteres")
	ErrInvalidLocation  = errors.New("La ubicación debe tener como máximo 100 caracteres")
	ErrInvalidEntryDate = errors.New("La fecha de ingreso es obligatoria")
	ErrInvalidBirthDate = errors.New("La fecha de nacimiento no puede ser posterior a la fecha de ingreso")
	ErrInvalidNote      = errors.New("La nota debe tener como máximo 500 caracteres")
	ErrInvalidState     = errors.New("El estado del verraco no es válido")
	ErrInvalidOrigin    = errors.New("El origen del verraco no es válido")
	ErrInvalidBreed     = errors.New("La raza es obligatoria")
	ErrInvalidUpdate    = errors.New("Debe actualizar al menos un campo del verraco")
	ErrInvalidCreatedBy = errors.New("El usuario que crea el verraco es obligatorio")
	ErrInvalidUpdatedBy = errors.New("El usuario que actualiza el verraco es obligatorio")
)

// Boar is the aggregate root that represents a boar (verraco) in the farm.
// It carries the audit fields and references the breed it belongs to.
type Boar struct {
	ID        uuid.UUID
	Code      string
	Location  *string
	Active    bool
	EntryDate time.Time
	BirthDate *time.Time
	Note      *string
	State     State
	Origin    Origin
	BreedID   uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

// NewBoarParams holds the fields required to build a new boar.
// Optional values are represented as pointers and may be nil; the state is not
// part of the params because every new boar starts as StateAlive.
type NewBoarParams struct {
	ID        uuid.UUID
	Code      string
	Location  *string
	EntryDate time.Time
	BirthDate *time.Time
	Note      *string
	Origin    Origin
	BreedID   uuid.UUID
	CreatedBy uuid.UUID
}

// UpdateBoarParams holds the mutable fields of a boar for an update.
// A nil pointer means the field is omitted and stays unchanged. Nullable string
// fields are cleared to NULL by providing an empty string, while the nullable
// date uses ClearBirthDate because it cannot carry an empty value. The state is
// intentionally absent because only ChangeState may alter it.
type UpdateBoarParams struct {
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

// NewBoar builds a boar after validating its fields.
// It defaults Active to true and the state to StateAlive, then sets the audit
// fields to the creator and both timestamps to now.
func NewBoar(params NewBoarParams, now time.Time) (*Boar, error) {
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

	if params.BreedID == uuid.Nil {
		return nil, ErrInvalidBreed
	}

	if params.CreatedBy == uuid.Nil {
		return nil, ErrInvalidCreatedBy
	}

	return &Boar{
		ID:        params.ID,
		Code:      code,
		Location:  location,
		Active:    true,
		EntryDate: params.EntryDate,
		BirthDate: birthDate,
		Note:      note,
		State:     StateAlive,
		Origin:    params.Origin,
		BreedID:   params.BreedID,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: params.CreatedBy,
		UpdatedBy: params.CreatedBy,
	}, nil
}

// Update applies only the non-nil fields of the given params to the boar.
// It validates each provided value and records updatedBy plus the timestamp,
// returning ErrInvalidUpdate when no field is provided; it never touches state.
func (b *Boar) Update(params UpdateBoarParams, updatedBy uuid.UUID, now time.Time) error {
	if b == nil || b.ID == uuid.Nil {
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

	validatedCode := b.Code
	if params.Code != nil {
		var err error
		validatedCode, err = validateCode(*params.Code)
		if err != nil {
			return err
		}
	}

	validatedLocation := b.Location
	if params.Location != nil {
		var err error
		validatedLocation, err = validateLocation(params.Location)
		if err != nil {
			return err
		}
	}

	validatedEntryDate := b.EntryDate
	if params.EntryDate != nil {
		if params.EntryDate.IsZero() {
			return ErrInvalidEntryDate
		}
		validatedEntryDate = *params.EntryDate
	}

	validatedBirthDate := b.BirthDate
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

	validatedNote := b.Note
	if params.Note != nil {
		var err error
		validatedNote, err = validateNote(params.Note)
		if err != nil {
			return err
		}
	}

	validatedOrigin := b.Origin
	if params.Origin != nil {
		if !isValidOrigin(*params.Origin) {
			return ErrInvalidOrigin
		}
		validatedOrigin = *params.Origin
	}

	validatedBreedID := b.BreedID
	if params.BreedID != nil {
		if *params.BreedID == uuid.Nil {
			return ErrInvalidBreed
		}
		validatedBreedID = *params.BreedID
	}

	b.Code = validatedCode
	b.Location = validatedLocation
	if params.Active != nil {
		b.Active = *params.Active
	}
	b.EntryDate = validatedEntryDate
	b.BirthDate = validatedBirthDate
	b.Note = validatedNote
	b.Origin = validatedOrigin
	b.BreedID = validatedBreedID
	b.UpdatedAt = now
	b.UpdatedBy = updatedBy
	return nil
}

// ChangeState validates the target state and applies it to the boar.
// It is the only way to alter State and is reserved for domain event handlers,
// never for the user-facing update flow; it records updatedBy plus the timestamp.
func (b *Boar) ChangeState(newState State, updatedBy uuid.UUID, now time.Time) error {
	if b == nil || b.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if !isValidState(newState) {
		return ErrInvalidState
	}

	b.State = newState
	b.UpdatedAt = now
	b.UpdatedBy = updatedBy
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

// isValidState checks whether the state belongs to the boar domain.
// Only the four states defined by the Boar_state enum are accepted.
func isValidState(state State) bool {
	switch state {
	case StateAlive, StateDead, StateDiscarded, StateSacrificed:
		return true
	default:
		return false
	}
}

// isValidOrigin checks whether the origin belongs to the boar domain.
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
