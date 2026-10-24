package pigletdeath

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	maxNoteLength = 500
	maxQuantity   = 10
	maxWeight     = 20.0
)

// PigletDeathCause represents the reason of a piglet death.
// Its values match the piglet_death_cause database enum.
type PigletDeathCause string

const (
	CauseCrushed            PigletDeathCause = "Aplastado"
	CauseWeakness           PigletDeathCause = "Debilidad"
	CauseDiarrhea           PigletDeathCause = "Diarrea"
	CauseDeformity          PigletDeathCause = "Deformidad"
	CauseOther              PigletDeathCause = "Otro"
	CauseCannibalism        PigletDeathCause = "Canibalismo"
	CauseOpenLeg            PigletDeathCause = "Pata_abierta"
	CauseBacteria           PigletDeathCause = "Bacteria"
	CauseMedicationReaction PigletDeathCause = "Reaccion_medicamento"
)

// Turn represents the shift when a piglet death was registered.
// Its values match the turn database enum.
type Turn string

const (
	TurnMorning        Turn = "Mañana"
	TurnAfternoon      Turn = "Tarde"
	TurnDawn           Turn = "Madrugada"
	TurnUnassistedDawn Turn = "Madrugada_no_asistida"
)

var (
	ErrInvalidID                = errors.New("El identificador de la muerte de lechones es obligatorio")
	ErrInvalidFarrowing         = errors.New("El parto de la muerte de lechones es obligatorio")
	ErrInvalidSow               = errors.New("La cerda de la muerte de lechones es obligatoria")
	ErrInvalidOperator          = errors.New("El operador de la muerte de lechones es obligatorio")
	ErrInvalidDeathDate         = errors.New("La fecha de la muerte es obligatoria")
	ErrInvalidFarrowDate        = errors.New("La fecha del parto es obligatoria")
	ErrInvalidQuantity          = errors.New("La cantidad de lechones muertos debe ser mayor a cero")
	ErrQuantityTooHigh          = errors.New("La cantidad de lechones muertos no puede ser mayor a 10")
	ErrWeightRequired           = errors.New("El peso es obligatorio")
	ErrInvalidWeight            = errors.New("El peso debe ser mayor a cero")
	ErrWeightTooHigh            = errors.New("El peso no puede ser mayor a 20 kg")
	ErrInvalidCause             = errors.New("La causa de la muerte no es válida")
	ErrInvalidTurn              = errors.New("El turno no es válido")
	ErrInvalidNote              = errors.New("La nota debe tener como máximo 500 caracteres")
	ErrInvalidCreatedBy         = errors.New("El usuario que crea la muerte es obligatorio")
	ErrInvalidUpdatedBy         = errors.New("El usuario que actualiza la muerte es obligatorio")
	ErrDeathDateBeforeFarrowing = errors.New("La fecha de la muerte no puede ser anterior al parto")
	ErrDeathDateInFuture        = errors.New("La fecha de la muerte no puede ser futura")
	ErrDeathDateBeforeWeaging   = errors.New("La fecha de la muerte no puede ser anterior al último destete parcial")
	ErrHasLaterWeaging          = errors.New("No se puede eliminar la muerte porque existe un destete posterior")
	ErrHasLaterPartialWeaging   = errors.New("No se puede eliminar la muerte porque existe un destete parcial posterior")
	ErrInvalidUpdate            = errors.New("Debe actualizar al menos un campo de la muerte")
	ErrImmutableDeathField      = errors.New("La fecha, la cantidad y la cerda de la muerte no se pueden editar")
)

// PigletDeath is the aggregate root that represents the death of piglets in a
// farrowing. It records the quantity, cause, shift and operator of the event.
type PigletDeath struct {
	ID           uuid.UUID
	FarrowingID  uuid.UUID
	SowID        uuid.UUID
	SowCode      string
	OperatorID   uuid.UUID
	OperatorName string
	DeathDate    time.Time
	Quantity     int
	Weight       *float64
	Cause        PigletDeathCause
	Turn         Turn
	Note         *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    uuid.UUID
	UpdatedBy    uuid.UUID
}

// NewPigletDeathParams holds the fields required to build a new piglet death.
// The farrowing and sow are resolved by the use case and the note may be nil.
type NewPigletDeathParams struct {
	ID           uuid.UUID
	FarrowingID  uuid.UUID
	SowID        uuid.UUID
	OperatorID   uuid.UUID
	OperatorName string
	DeathDate    time.Time
	Quantity     int
	Weight       *float64
	Cause        PigletDeathCause
	Turn         Turn
	Note         *string
	CreatedBy    uuid.UUID
}

// NewPigletDeath builds a piglet death after validating its fields and dates.
// It requires the death date to be on or after farrowDate and not later than
// today, and rejects a quantity over 10, a missing or non-positive weight and a
// weight over 20 kg. The last partial weaging date may be zero when the farrowing
// is not a nurse; otherwise the death date cannot be earlier than that date.
func NewPigletDeath(params NewPigletDeathParams, farrowDate time.Time, lastPartialWeagingDate time.Time, now time.Time) (*PigletDeath, error) {
	if params.ID == uuid.Nil {
		return nil, ErrInvalidID
	}
	if params.FarrowingID == uuid.Nil {
		return nil, ErrInvalidFarrowing
	}
	if params.SowID == uuid.Nil {
		return nil, ErrInvalidSow
	}
	if params.OperatorID == uuid.Nil {
		return nil, ErrInvalidOperator
	}
	if params.DeathDate.IsZero() {
		return nil, ErrInvalidDeathDate
	}
	if farrowDate.IsZero() {
		return nil, ErrInvalidFarrowDate
	}

	if err := validateDeathDate(params.DeathDate, farrowDate, lastPartialWeagingDate, now); err != nil {
		return nil, err
	}

	if params.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}
	if params.Quantity > maxQuantity {
		return nil, ErrQuantityTooHigh
	}
	if params.Weight == nil {
		return nil, ErrWeightRequired
	}
	if *params.Weight <= 0 {
		return nil, ErrInvalidWeight
	}
	if *params.Weight > maxWeight {
		return nil, ErrWeightTooHigh
	}
	if !isValidCause(params.Cause) {
		return nil, ErrInvalidCause
	}
	if !isValidTurn(params.Turn) {
		return nil, ErrInvalidTurn
	}

	note, err := validateNote(params.Note)
	if err != nil {
		return nil, err
	}

	if params.CreatedBy == uuid.Nil {
		return nil, ErrInvalidCreatedBy
	}

	return &PigletDeath{
		ID:           params.ID,
		FarrowingID:  params.FarrowingID,
		SowID:        params.SowID,
		OperatorID:   params.OperatorID,
		OperatorName: params.OperatorName,
		DeathDate:    params.DeathDate,
		Quantity:     params.Quantity,
		Weight:       params.Weight,
		Cause:        params.Cause,
		Turn:         params.Turn,
		Note:         note,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    params.CreatedBy,
		UpdatedBy:    params.CreatedBy,
	}, nil
}

// validateDeathDate checks the death date bounds against the farrow date and the
// last partial weaging date. It rejects dates after today, dates before the
// farrow date and dates before the last partial weaging when one is provided.
func validateDeathDate(deathDate time.Time, farrowDate time.Time, lastPartialWeagingDate time.Time, now time.Time) error {
	death := truncateToDay(deathDate)
	if death.After(truncateToDay(now)) {
		return ErrDeathDateInFuture
	}
	if death.Before(truncateToDay(farrowDate)) {
		return ErrDeathDateBeforeFarrowing
	}
	if !lastPartialWeagingDate.IsZero() && death.Before(truncateToDay(lastPartialWeagingDate)) {
		return ErrDeathDateBeforeWeaging
	}
	return nil
}

// truncateToDay removes the time portion from a timestamp in UTC.
// It is used to compare the death and farrow dates at day granularity.
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

// UpdatePigletDeathParams holds the editable fields of a piglet death update.
// A nil pointer means the field is omitted and stays unchanged, except for the
// locked fields, which are only accepted when they match the stored values. The
// operator name is resolved by the use case from the operator identifier.
type UpdatePigletDeathParams struct {
	DeathDate    *time.Time
	Quantity     *int
	SowID        *uuid.UUID
	OperatorID   *uuid.UUID
	OperatorName *string
	Weight       *float64
	Cause        *PigletDeathCause
	Turn         *Turn
	Note         *string
}

// Update applies the provided editable fields to the piglet death.
// It rejects a change to the death date, quantity or sow, requires a positive
// weight of at most 20 kg, validates the cause, turn and note, and records the
// actor plus the timestamp. It returns ErrInvalidUpdate when no field is provided.
func (d *PigletDeath) Update(params UpdatePigletDeathParams, updatedBy uuid.UUID, now time.Time) error {
	if d == nil || d.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if params.DeathDate == nil && params.Quantity == nil && params.SowID == nil &&
		params.OperatorID == nil && params.Weight == nil && params.Cause == nil &&
		params.Turn == nil && params.Note == nil {
		return ErrInvalidUpdate
	}

	if params.DeathDate != nil &&
		!truncateToDay(*params.DeathDate).Equal(truncateToDay(d.DeathDate)) {
		return ErrImmutableDeathField
	}
	if params.Quantity != nil && *params.Quantity != d.Quantity {
		return ErrImmutableDeathField
	}
	if params.SowID != nil && *params.SowID != d.SowID {
		return ErrImmutableDeathField
	}

	if params.Weight == nil {
		return ErrWeightRequired
	}
	if *params.Weight <= 0 {
		return ErrInvalidWeight
	}
	if *params.Weight > maxWeight {
		return ErrWeightTooHigh
	}

	if params.Cause != nil && !isValidCause(*params.Cause) {
		return ErrInvalidCause
	}
	if params.Turn != nil && !isValidTurn(*params.Turn) {
		return ErrInvalidTurn
	}

	note, err := validateNote(params.Note)
	if err != nil {
		return err
	}

	if params.OperatorID != nil {
		d.OperatorID = *params.OperatorID
		if params.OperatorName != nil {
			d.OperatorName = *params.OperatorName
		}
	}
	d.Weight = params.Weight
	if params.Cause != nil {
		d.Cause = *params.Cause
	}
	if params.Turn != nil {
		d.Turn = *params.Turn
	}
	if params.Note != nil {
		d.Note = note
	}

	d.UpdatedAt = now
	d.UpdatedBy = updatedBy
	return nil
}

// isValidCause checks whether the cause belongs to the piglet death domain.
// Only the nine causes defined by the piglet_death_cause enum are accepted.
func isValidCause(cause PigletDeathCause) bool {
	switch cause {
	case CauseCrushed, CauseWeakness, CauseDiarrhea, CauseDeformity, CauseOther,
		CauseCannibalism, CauseOpenLeg, CauseBacteria, CauseMedicationReaction:
		return true
	default:
		return false
	}
}

// ParseCause trims and validates a raw cause value.
// It returns ErrInvalidCause when the value is empty or unknown.
func ParseCause(value string) (PigletDeathCause, error) {
	cause := PigletDeathCause(strings.TrimSpace(value))
	if !isValidCause(cause) {
		return "", ErrInvalidCause
	}
	return cause, nil
}

// isValidTurn checks whether the turn belongs to the piglet death domain.
// Only the four shifts defined by the turn enum are accepted.
func isValidTurn(turn Turn) bool {
	switch turn {
	case TurnMorning, TurnAfternoon, TurnDawn, TurnUnassistedDawn:
		return true
	default:
		return false
	}
}

// ParseTurn trims and validates a raw turn value.
// It returns ErrInvalidTurn when the value is empty or unknown.
func ParseTurn(value string) (Turn, error) {
	turn := Turn(strings.TrimSpace(value))
	if !isValidTurn(turn) {
		return "", ErrInvalidTurn
	}
	return turn, nil
}
