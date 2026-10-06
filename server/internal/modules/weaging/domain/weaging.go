package weaging

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	maxNoteLength        = 500
	maxDestinationLength = 100
)

var (
	ErrInvalidID                  = errors.New("El identificador del destete es obligatorio")
	ErrInvalidFarrowing           = errors.New("El parto del destete es obligatorio")
	ErrInvalidSow                 = errors.New("La cerda del destete es obligatoria")
	ErrInvalidWeagingDate         = errors.New("La fecha del destete es obligatoria")
	ErrInvalidFarrowDate          = errors.New("La fecha del parto es obligatoria")
	ErrInvalidQuantity            = errors.New("La cantidad de lechones debe ser mayor a cero")
	ErrInvalidTotalWeight         = errors.New("El peso total debe ser mayor a cero")
	ErrInvalidDestination         = errors.New("El destino debe tener como máximo 100 caracteres")
	ErrInvalidNote                = errors.New("La nota debe tener como máximo 500 caracteres")
	ErrInvalidCreatedBy           = errors.New("El usuario que crea el destete es obligatorio")
	ErrInvalidUpdatedBy           = errors.New("El usuario que actualiza el destete es obligatorio")
	ErrWeagingDateBeforeFarrowing = errors.New("La fecha del destete debe ser posterior al parto")
	ErrWeagingDateBeforeEvents    = errors.New("La fecha del destete debe ser posterior a las muertes, traslados y destetes parciales")
	ErrWeagingDateInFuture        = errors.New("La fecha del destete no puede ser futura")
)

// Weaging is the aggregate root that represents the full weaning of a farrowing.
// It records the weaned quantity, optional weight and destination.
type Weaging struct {
	ID          uuid.UUID
	FarrowingID uuid.UUID
	SowID       uuid.UUID
	WeagingDate time.Time
	Quantity    int
	TotalWeight *float64
	Destination *string
	Note        *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
}

// NewWeagingParams holds the fields required to build a new weaging.
// The farrowing and sow are resolved by the use case and weight, destination and note may be nil.
type NewWeagingParams struct {
	ID          uuid.UUID
	FarrowingID uuid.UUID
	SowID       uuid.UUID
	WeagingDate time.Time
	Quantity    int
	TotalWeight *float64
	Destination *string
	Note        *string
	CreatedBy   uuid.UUID
}

// NewWeaging builds a weaging after validating its fields and dates.
// It requires the weaging date to be strictly after the farrow date and after the
// latest death, fostering or partial weaging event, and never later than today.
// The last event date may be zero when the farrowing has no related events.
func NewWeaging(
	params NewWeagingParams,
	farrowDate time.Time,
	lastEventDate time.Time,
	now time.Time,
) (*Weaging, error) {
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

	destination, err := validateDestination(params.Destination)
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

	return &Weaging{
		ID:          params.ID,
		FarrowingID: params.FarrowingID,
		SowID:       params.SowID,
		WeagingDate: params.WeagingDate,
		Quantity:    params.Quantity,
		TotalWeight: params.TotalWeight,
		Destination: destination,
		Note:        note,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   params.CreatedBy,
		UpdatedBy:   params.CreatedBy,
	}, nil
}

// validateWeagingDate checks the weaging date bounds against the farrowing and events.
// It rejects dates after today, dates on or before the farrow date and dates on or
// before the latest death, fostering or partial weaging event when one is provided.
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

// validateDestination trims the optional destination and checks its length.
// It returns nil when destination is nil or empty (clear to NULL) and
// ErrInvalidDestination when it is too long.
func validateDestination(destination *string) (*string, error) {
	if destination == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*destination)
	if trimmed == "" {
		return nil, nil
	}
	if len([]rune(trimmed)) > maxDestinationLength {
		return nil, ErrInvalidDestination
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
