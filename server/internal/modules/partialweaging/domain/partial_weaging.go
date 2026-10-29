package partialweaging

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	maxNoteLength  = 500
	maxQuantity    = 30
	maxTotalWeight = 200.0
)

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
	ErrQuantityExceedsMax         = errors.New("La cantidad de lechones no puede ser mayor a 30")
	ErrQuantityMismatch           = errors.New("La cantidad de lechones debe ser igual a los lechones actuales del parto")
	ErrInvalidTotalWeight         = errors.New("El peso total debe ser mayor a cero")
	ErrTotalWeightExceedsMax      = errors.New("El peso total no puede ser mayor a 200 kg")
	ErrInvalidType                = errors.New("El tipo de destete parcial no es válido")
	ErrInvalidNote                = errors.New("La nota debe tener como máximo 500 caracteres")
	ErrInvalidCreatedBy           = errors.New("El usuario que crea el destete parcial es obligatorio")
	ErrInvalidUpdatedBy           = errors.New("El usuario que actualiza el destete parcial es obligatorio")
	ErrWeagingDateBeforeFarrowing = errors.New("La fecha del destete parcial debe ser posterior al parto")
	ErrWeagingDateBeforeEvents    = errors.New("La fecha del destete parcial no puede ser anterior a las muertes y traslados")
	ErrWeagingDateBeforePartial   = errors.New("La fecha del destete parcial no puede ser anterior al último destete parcial")
	ErrWeagingDateInFuture        = errors.New("La fecha del destete parcial no puede ser futura")
	ErrHasFutureEvents            = errors.New("No se puede eliminar el destete parcial porque existen montas o bajas posteriores")
	ErrHasLaterNurseEvents        = errors.New("No se puede eliminar el destete parcial porque existen eventos posteriores al destete parcial")
	ErrInvalidUpdate              = errors.New("Debe indicar al menos un campo para actualizar")
)

// PartialWeaging is the aggregate root that represents a partial weaning of a
// farrowing. It records the weaned quantity, optional weight and its type.
type PartialWeaging struct {
	ID          uuid.UUID
	FarrowingID uuid.UUID
	SowID       uuid.UUID
	SowCode     string
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
	SowCode     string
	WeagingDate time.Time
	Quantity    int
	TotalWeight *float64
	Type        Type
	Note        *string
	CreatedBy   uuid.UUID
}

// PartialWeagingBounds holds the farrowing references used to validate a new
// partial weaging. A zero event date means the farrowing has no death or
// fostering, and a zero partial date means it has no previous partial weaging.
type PartialWeagingBounds struct {
	FarrowDate             time.Time
	LastEventDate          time.Time
	LastPartialWeagingDate time.Time
	CurrentPiglets         int
}

// UpdatePartialWeagingParams holds the editable fields of a partial weaging.
// A nil pointer means the field is omitted and stays unchanged, an empty note
// clears it and ClearTotalWeight removes the optional weight. Quantity, type,
// sow, farrowing and date are intentionally absent because they are immutable.
type UpdatePartialWeagingParams struct {
	TotalWeight      *float64
	ClearTotalWeight bool
	Note             *string
}

// NewPartialWeaging builds a partial weaging after validating its fields and dates.
// It requires the weaging date to be strictly after the farrow date, not earlier than
// the latest death or fostering event nor the last partial weaging, and never later
// than today. The quantity must equal the farrowing's current piglets and be at most 30.
func NewPartialWeaging(
	params NewPartialWeagingParams,
	bounds PartialWeagingBounds,
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
	if bounds.FarrowDate.IsZero() {
		return nil, ErrInvalidFarrowDate
	}

	if err := validateWeagingDate(params.WeagingDate, bounds, now); err != nil {
		return nil, err
	}

	if err := validateQuantity(params.Quantity, bounds.CurrentPiglets); err != nil {
		return nil, err
	}
	if params.TotalWeight != nil {
		if *params.TotalWeight <= 0 {
			return nil, ErrInvalidTotalWeight
		}
		if *params.TotalWeight > maxTotalWeight {
			return nil, ErrTotalWeightExceedsMax
		}
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
		SowCode:     params.SowCode,
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

// Update applies the editable fields to the partial weaging after validating them.
// It never changes the sow, farrowing, date, quantity or type, and returns
// ErrInvalidUpdate when no field is provided.
func (p *PartialWeaging) Update(
	params UpdatePartialWeagingParams,
	updatedBy uuid.UUID,
	now time.Time,
) error {
	if p == nil || p.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if params.TotalWeight == nil && !params.ClearTotalWeight && params.Note == nil {
		return ErrInvalidUpdate
	}
	if params.TotalWeight != nil && params.ClearTotalWeight {
		return ErrInvalidUpdate
	}

	if params.TotalWeight != nil {
		if *params.TotalWeight <= 0 {
			return ErrInvalidTotalWeight
		}
		if *params.TotalWeight > maxTotalWeight {
			return ErrTotalWeightExceedsMax
		}
		p.TotalWeight = params.TotalWeight
	} else if params.ClearTotalWeight {
		p.TotalWeight = nil
	}

	if params.Note != nil {
		note, err := validateNote(params.Note)
		if err != nil {
			return err
		}
		p.Note = note
	}

	p.UpdatedAt = now
	p.UpdatedBy = updatedBy
	return nil
}

// validateWeagingDate checks the weaging date bounds against the farrowing, its events
// and its previous partial weagings. It rejects future dates, dates on or before the
// farrow date, dates earlier than the latest death or fostering event and dates earlier
// than the last partial weaging. Events on the same day are allowed.
func validateWeagingDate(weagingDate time.Time, bounds PartialWeagingBounds, now time.Time) error {
	weaging := truncateToDay(weagingDate)
	if weaging.After(truncateToDay(now)) {
		return ErrWeagingDateInFuture
	}
	if !weaging.After(truncateToDay(bounds.FarrowDate)) {
		return ErrWeagingDateBeforeFarrowing
	}
	if !bounds.LastEventDate.IsZero() && weaging.Before(truncateToDay(bounds.LastEventDate)) {
		return ErrWeagingDateBeforeEvents
	}
	if !bounds.LastPartialWeagingDate.IsZero() && weaging.Before(truncateToDay(bounds.LastPartialWeagingDate)) {
		return ErrWeagingDateBeforePartial
	}
	return nil
}

// validateQuantity checks the weaning quantity against the allowed range and balance.
// It rejects non-positive values, values above 30 and values that do not match the
// farrowing's current piglets.
func validateQuantity(quantity int, currentPiglets int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	if quantity > maxQuantity {
		return ErrQuantityExceedsMax
	}
	if quantity != currentPiglets {
		return ErrQuantityMismatch
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
