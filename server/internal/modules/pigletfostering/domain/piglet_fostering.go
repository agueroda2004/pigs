package pigletfostering

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	maxNoteLength = 500

	// MaxQuantity is the largest number of piglets a single fostering may move.
	MaxQuantity = 30
)

var (
	ErrInvalidID                        = errors.New("El identificador del traslado es obligatorio")
	ErrInvalidDonor                     = errors.New("El parto donante es obligatorio")
	ErrInvalidReceiver                  = errors.New("El parto receptor es obligatorio")
	ErrInvalidDonorSow                  = errors.New("La cerda donante es obligatoria")
	ErrInvalidReceiverSow               = errors.New("La cerda receptora es obligatoria")
	ErrInvalidMovementDate              = errors.New("La fecha del traslado es obligatoria")
	ErrInvalidFarrowDate                = errors.New("La fecha del parto es obligatoria")
	ErrInvalidQuantity                  = errors.New("La cantidad de lechones debe ser mayor a cero")
	ErrQuantityExceedsMax               = errors.New("La cantidad de lechones no puede superar 30")
	ErrInvalidNote                      = errors.New("La nota debe tener como máximo 500 caracteres")
	ErrInvalidCreatedBy                 = errors.New("El usuario que crea el traslado es obligatorio")
	ErrInvalidUpdatedBy                 = errors.New("El usuario que actualiza el traslado es obligatorio")
	ErrInvalidDeletedBy                 = errors.New("El usuario que elimina el traslado es obligatorio")
	ErrInvalidUpdate                    = errors.New("Debe indicar al menos un campo para actualizar")
	ErrSameFarrowing                    = errors.New("El parto donante y el receptor no pueden ser el mismo")
	ErrMovementDateBeforeFarrowing      = errors.New("La fecha del traslado debe ser posterior a los partos")
	ErrMovementDateBeforePartialWeaging = errors.New("La fecha del traslado debe ser posterior al último destete parcial")
	ErrMovementDateInFuture             = errors.New("La fecha del traslado no puede ser futura")
	ErrHasLaterWeaging                  = errors.New("No se puede eliminar el traslado porque existe un destete posterior")
	ErrHasLaterPartialWeaging           = errors.New("No se puede eliminar el traslado porque existe un destete parcial posterior")
)

// PigletFostering is the aggregate root that represents a transfer of piglets
// between two farrowings. It records the donor and receiver farrowings, the
// moved quantity and the movement date, plus the codes of both sows.
type PigletFostering struct {
	ID                  uuid.UUID
	DonorFarrowingID    uuid.UUID
	ReceiverFarrowingID uuid.UUID
	DonorSowID          uuid.UUID
	DonorSowCode        string
	ReceiverSowID       uuid.UUID
	ReceiverSowCode     string
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

// FosteringDateBounds holds the farrow and partial weaging dates used to validate
// a movement date. A zero partial weaging date means that side is not a nurse.
type FosteringDateBounds struct {
	DonorFarrowDate            time.Time
	ReceiverFarrowDate         time.Time
	DonorPartialWeagingDate    time.Time
	ReceiverPartialWeagingDate time.Time
}

// FosteringDeleteBounds holds the latest partial weaging and weaging dates of both
// farrowings. A zero date means that the farrowing has no record of that kind.
type FosteringDeleteBounds struct {
	DonorPartialWeagingDate    time.Time
	DonorWeagingDate           time.Time
	ReceiverPartialWeagingDate time.Time
	ReceiverWeagingDate        time.Time
}

// NewPigletFostering builds a fostering after validating its fields and dates.
// It rejects the same farrowing on both sides and requires the movement date to
// be strictly after both farrow dates, after each nurse partial weaging and not
// later than today.
func NewPigletFostering(
	params NewPigletFosteringParams,
	bounds FosteringDateBounds,
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
	if bounds.DonorFarrowDate.IsZero() || bounds.ReceiverFarrowDate.IsZero() {
		return nil, ErrInvalidFarrowDate
	}

	if err := validateMovementDate(
		params.MovementDate,
		bounds.DonorFarrowDate,
		bounds.ReceiverFarrowDate,
		bounds.DonorPartialWeagingDate,
		bounds.ReceiverPartialWeagingDate,
		now,
	); err != nil {
		return nil, err
	}

	if err := validateQuantity(params.Quantity); err != nil {
		return nil, err
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

// UpdatePigletFosteringParams holds the editable fields of a fostering update.
// A nil pointer means the field is omitted and stays unchanged; the donor and
// receiver farrowings are fixed and can never be edited.
type UpdatePigletFosteringParams struct {
	MovementDate *time.Time
	Quantity     *int
	Note         *string
}

// Update applies the provided editable fields to the fostering.
// It validates the quantity and note always, and revalidates the movement date
// only when it changes, recording the actor plus the timestamp.
func (f *PigletFostering) Update(
	params UpdatePigletFosteringParams,
	bounds FosteringDateBounds,
	updatedBy uuid.UUID,
	now time.Time,
) error {
	if f == nil || f.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if params.MovementDate == nil && params.Quantity == nil && params.Note == nil {
		return ErrInvalidUpdate
	}

	if params.MovementDate != nil &&
		!truncateToDay(*params.MovementDate).Equal(truncateToDay(f.MovementDate)) {
		if bounds.DonorFarrowDate.IsZero() || bounds.ReceiverFarrowDate.IsZero() {
			return ErrInvalidFarrowDate
		}
		if err := validateMovementDate(
			*params.MovementDate,
			bounds.DonorFarrowDate,
			bounds.ReceiverFarrowDate,
			bounds.DonorPartialWeagingDate,
			bounds.ReceiverPartialWeagingDate,
			now,
		); err != nil {
			return err
		}
		f.MovementDate = *params.MovementDate
	}

	if params.Quantity != nil {
		if err := validateQuantity(*params.Quantity); err != nil {
			return err
		}
		f.Quantity = *params.Quantity
	}

	if params.Note != nil {
		note, err := validateNote(params.Note)
		if err != nil {
			return err
		}
		f.Note = note
	}

	f.UpdatedAt = now
	f.UpdatedBy = updatedBy
	return nil
}

// ValidateDelete checks that a fostering can be removed by the given actor.
// It rejects a missing actor and blocks the deletion when either farrowing has a
// full or partial weaging registered strictly after the movement date.
func (f *PigletFostering) ValidateDelete(bounds FosteringDeleteBounds, deletedBy uuid.UUID) error {
	if f == nil || f.ID == uuid.Nil {
		return ErrInvalidID
	}
	if deletedBy == uuid.Nil {
		return ErrInvalidDeletedBy
	}

	if isDayAfter(bounds.DonorWeagingDate, f.MovementDate) ||
		isDayAfter(bounds.ReceiverWeagingDate, f.MovementDate) {
		return ErrHasLaterWeaging
	}
	if isDayAfter(bounds.DonorPartialWeagingDate, f.MovementDate) ||
		isDayAfter(bounds.ReceiverPartialWeagingDate, f.MovementDate) {
		return ErrHasLaterPartialWeaging
	}
	return nil
}

// isDayAfter reports whether value is set and falls after the reference date.
// It compares at day granularity and treats a zero value as no date at all.
func isDayAfter(value time.Time, reference time.Time) bool {
	if value.IsZero() {
		return false
	}
	return truncateToDay(value).After(truncateToDay(reference))
}

// validateQuantity checks that the moved quantity is positive and at most MaxQuantity.
// It returns ErrInvalidQuantity or ErrQuantityExceedsMax when the bound is broken.
func validateQuantity(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	if quantity > MaxQuantity {
		return ErrQuantityExceedsMax
	}
	return nil
}

// validateMovementDate checks the movement date bounds against both farrowings and
// their nurse partial weagings. It rejects dates after today, dates on or before
// either farrow date and dates on or before a nurse partial weaging.
func validateMovementDate(
	movementDate time.Time,
	donorFarrowDate time.Time,
	receiverFarrowDate time.Time,
	donorPartialWeagingDate time.Time,
	receiverPartialWeagingDate time.Time,
	now time.Time,
) error {
	movement := truncateToDay(movementDate)
	if movement.After(truncateToDay(now)) {
		return ErrMovementDateInFuture
	}
	donor := truncateToDay(donorFarrowDate)
	receiver := truncateToDay(receiverFarrowDate)
	if !movement.After(donor) || !movement.After(receiver) {
		return ErrMovementDateBeforeFarrowing
	}
	if !donorPartialWeagingDate.IsZero() && !movement.After(truncateToDay(donorPartialWeagingDate)) {
		return ErrMovementDateBeforePartialWeaging
	}
	if !receiverPartialWeagingDate.IsZero() && !movement.After(truncateToDay(receiverPartialWeagingDate)) {
		return ErrMovementDateBeforePartialWeaging
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
