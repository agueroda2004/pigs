package pigletfostering

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxNoteLength = 500

var (
	ErrInvalidID                   = errors.New("El identificador del traslado es obligatorio")
	ErrInvalidDonor                = errors.New("El parto donante es obligatorio")
	ErrInvalidReceiver             = errors.New("El parto receptor es obligatorio")
	ErrInvalidDonorSow             = errors.New("La cerda donante es obligatoria")
	ErrInvalidReceiverSow          = errors.New("La cerda receptora es obligatoria")
	ErrInvalidMovementDate         = errors.New("La fecha del traslado es obligatoria")
	ErrInvalidFarrowDate           = errors.New("La fecha del parto es obligatoria")
	ErrInvalidQuantity             = errors.New("La cantidad de lechones debe ser mayor a cero")
	ErrInvalidNote                 = errors.New("La nota debe tener como máximo 500 caracteres")
	ErrInvalidCreatedBy            = errors.New("El usuario que crea el traslado es obligatorio")
	ErrInvalidUpdatedBy            = errors.New("El usuario que actualiza el traslado es obligatorio")
	ErrSameFarrowing               = errors.New("El parto donante y el receptor no pueden ser el mismo")
	ErrMovementDateBeforeFarrowing = errors.New("La fecha del traslado debe ser posterior a los partos")
	ErrMovementDateInFuture        = errors.New("La fecha del traslado no puede ser futura")
)

// PigletFostering is the aggregate root that represents a transfer of piglets
// between two farrowings. It records the donor and receiver farrowings, the
// moved quantity and the movement date.
type PigletFostering struct {
	ID                  uuid.UUID
	DonorFarrowingID    uuid.UUID
	ReceiverFarrowingID uuid.UUID
	DonorSowID          uuid.UUID
	ReceiverSowID       uuid.UUID
	MovementDate        time.Time
	Quantity            int
	Note                *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CreatedBy           uuid.UUID
	UpdatedBy           uuid.UUID
}

// NewPigletFosteringParams holds the fields required to build a new fostering.
// The farrowings and sows are resolved by the use case and the note may be nil.
type NewPigletFosteringParams struct {
	ID                  uuid.UUID
	DonorFarrowingID    uuid.UUID
	ReceiverFarrowingID uuid.UUID
	DonorSowID          uuid.UUID
	ReceiverSowID       uuid.UUID
	MovementDate        time.Time
	Quantity            int
	Note                *string
	CreatedBy           uuid.UUID
}

// NewPigletFostering builds a fostering after validating its fields and dates.
// It rejects the same farrowing on both sides and requires the movement date to
// be strictly after both farrow dates and not later than today.
func NewPigletFostering(
	params NewPigletFosteringParams,
	donorFarrowDate time.Time,
	receiverFarrowDate time.Time,
	now time.Time,
) (*PigletFostering, error) {
	if params.ID == uuid.Nil {
		return nil, ErrInvalidID
	}
	if params.DonorFarrowingID == uuid.Nil {
		return nil, ErrInvalidDonor
	}
	if params.ReceiverFarrowingID == uuid.Nil {
		return nil, ErrInvalidReceiver
	}
	if params.DonorFarrowingID == params.ReceiverFarrowingID {
		return nil, ErrSameFarrowing
	}
	if params.DonorSowID == uuid.Nil {
		return nil, ErrInvalidDonorSow
	}
	if params.ReceiverSowID == uuid.Nil {
		return nil, ErrInvalidReceiverSow
	}
	if params.MovementDate.IsZero() {
		return nil, ErrInvalidMovementDate
	}
	if donorFarrowDate.IsZero() || receiverFarrowDate.IsZero() {
		return nil, ErrInvalidFarrowDate
	}

	if err := validateMovementDate(params.MovementDate, donorFarrowDate, receiverFarrowDate, now); err != nil {
		return nil, err
	}

	if params.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	note, err := validateNote(params.Note)
	if err != nil {
		return nil, err
	}

	if params.CreatedBy == uuid.Nil {
		return nil, ErrInvalidCreatedBy
	}

	return &PigletFostering{
		ID:                  params.ID,
		DonorFarrowingID:    params.DonorFarrowingID,
		ReceiverFarrowingID: params.ReceiverFarrowingID,
		DonorSowID:          params.DonorSowID,
		ReceiverSowID:       params.ReceiverSowID,
		MovementDate:        params.MovementDate,
		Quantity:            params.Quantity,
		Note:                note,
		CreatedAt:           now,
		UpdatedAt:           now,
		CreatedBy:           params.CreatedBy,
		UpdatedBy:           params.CreatedBy,
	}, nil
}

// validateMovementDate checks the movement date bounds against both farrowings.
// It rejects dates after today and dates on or before either farrow date.
func validateMovementDate(movementDate time.Time, donorFarrowDate time.Time, receiverFarrowDate time.Time, now time.Time) error {
	movement := truncateToDay(movementDate)
	if movement.After(truncateToDay(now)) {
		return ErrMovementDateInFuture
	}
	donor := truncateToDay(donorFarrowDate)
	receiver := truncateToDay(receiverFarrowDate)
	if !movement.After(donor) || !movement.After(receiver) {
		return ErrMovementDateBeforeFarrowing
	}
	return nil
}

// truncateToDay removes the time portion from a timestamp in UTC.
// It is used to compare the movement and farrow dates at day granularity.
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
