package partialweaging

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxNoteLength = 500

// Type represents the kind of a partial weaging.
// Its values match the partial_weaging_type database enum.
type Type string

const (
	TypeNormal         Type = "Normal"
	TypeNodriza        Type = "Nodriza"
	TypeBajaViabilidad Type = "Baja_Viabilidad"
)

var (
	ErrInvalidID                  = errors.New("El identificador del destete parcial es obligatorio")
	ErrInvalidFarrowing           = errors.New("El parto del destete parcial es obligatorio")
	ErrInvalidSow                 = errors.New("La cerda del destete parcial es obligatoria")
	ErrInvalidWeagingDate         = errors.New("La fecha del destete parcial es obligatoria")
	ErrInvalidFarrowDate          = errors.New("La fecha del parto es obligatoria")
	ErrInvalidQuantity            = errors.New("La cantidad de lechones debe ser mayor a cero")
	ErrInvalidTotalWeight         = errors.New("El peso total debe ser mayor a cero")
	ErrInvalidType                = errors.New("El tipo de destete parcial no es válido")
	ErrInvalidNote                = errors.New("La nota debe tener como máximo 500 caracteres")
	ErrInvalidCreatedBy           = errors.New("El usuario que crea el destete parcial es obligatorio")
	ErrInvalidUpdatedBy           = errors.New("El usuario que actualiza el destete parcial es obligatorio")
	ErrWeagingDateBeforeFarrowing = errors.New("La fecha del destete parcial debe ser posterior al parto")
	ErrWeagingDateBeforeEvents    = errors.New("La fecha del destete parcial debe ser posterior a las muertes y traslados")
	ErrWeagingDateInFuture        = errors.New("La fecha del destete parcial no puede ser futura")
)

// PartialWeaging is the aggregate root that represents a partial weaning of a
// farrowing. It records the weaned quantity, optional weight and its type.
type PartialWeaging struct {
	ID          uuid.UUID
	FarrowingID uuid.UUID
	SowID       uuid.UUID
	WeagingDate time.Time
	Quantity    int
	TotalWeight *float64
	Type        Type
	Note        *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
}

// NewPartialWeagingParams holds the fields required to build a new partial weaging.
// The farrowing and sow are resolved by the use case and both note and weight may be nil.
type NewPartialWeagingParams struct {
	ID          uuid.UUID
	FarrowingID uuid.UUID
	SowID       uuid.UUID
	WeagingDate time.Time
	Quantity    int
	TotalWeight *float64
	Type        Type
	Note        *string
	CreatedBy   uuid.UUID
}

// NewPartialWeaging builds a partial weaging after validating its fields and dates.
// It requires the weaging date to be strictly after the farrow date and after the
// latest death or fostering event, and never later than today. The last event date
// may be zero when the farrowing has no related events.
func NewPartialWeaging(
	params NewPartialWeagingParams,
	farrowDate time.Time,
	lastEventDate time.Time,
	now time.Time,
) (*PartialWeaging, error) {
	if params.ID == uuid.Nil {
		return nil, ErrInvalidID
	}
	if params.FarrowingID == uuid.Nil {
		return nil, ErrInvalidFarrowing
	}
	if params.SowID == uuid.Nil {
		return nil, ErrInvalidSow
	}
	if params.WeagingDate.IsZero() {
		return nil, ErrInvalidWeagingDate
	}
	if farrowDate.IsZero() {
		return nil, ErrInvalidFarrowDate
	}

	if err := validateWeagingDate(params.WeagingDate, farrowDate, lastEventDate, now); err != nil {
		return nil, err
	}

	if params.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}
	if params.TotalWeight != nil && *params.TotalWeight <= 0 {
		return nil, ErrInvalidTotalWeight
	}
	if !isValidType(params.Type) {
		return nil, ErrInvalidType
	}

	note, err := validateNote(params.Note)
	if err != nil {
		return nil, err
	}

	if params.CreatedBy == uuid.Nil {
		return nil, ErrInvalidCreatedBy
	}

	return &PartialWeaging{
		ID:          params.ID,
		FarrowingID: params.FarrowingID,
		SowID:       params.SowID,
		WeagingDate: params.WeagingDate,
		Quantity:    params.Quantity,
		TotalWeight: params.TotalWeight,
		Type:        params.Type,
		Note:        note,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   params.CreatedBy,
		UpdatedBy:   params.CreatedBy,
	}, nil
}

// validateWeagingDate checks the weaging date bounds against the farrowing and events.
// It rejects dates after today, dates on or before the farrow date and dates on or
// before the latest death or fostering event when one is provided.
func validateWeagingDate(weagingDate time.Time, farrowDate time.Time, lastEventDate time.Time, now time.Time) error {
	weaging := truncateToDay(weagingDate)
	if weaging.After(truncateToDay(now)) {
		return ErrWeagingDateInFuture
	}
	if !weaging.After(truncateToDay(farrowDate)) {
		return ErrWeagingDateBeforeFarrowing
	}
	if !lastEventDate.IsZero() && !weaging.After(truncateToDay(lastEventDate)) {
		return ErrWeagingDateBeforeEvents
	}
	return nil
}

// truncateToDay removes the time portion from a timestamp in UTC.
// It is used to compare the weaging, farrow and event dates at day granularity.
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

// isValidType checks whether the type belongs to the partial weaging domain.
// Only the three types defined by the partial_weaging_type enum are accepted.
func isValidType(value Type) bool {
	switch value {
	case TypeNormal, TypeNodriza, TypeBajaViabilidad:
		return true
	default:
		return false
	}
}

// ParseType trims and validates a raw type value.
// It returns ErrInvalidType when the value is empty or unknown.
func ParseType(value string) (Type, error) {
	parsed := Type(strings.TrimSpace(value))
	if !isValidType(parsed) {
		return "", ErrInvalidType
	}
	return parsed, nil
}
