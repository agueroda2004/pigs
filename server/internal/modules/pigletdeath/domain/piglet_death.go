package pigletdeath

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxNoteLength = 500

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
	ErrInvalidWeight            = errors.New("El peso debe ser mayor a cero")
	ErrInvalidCause             = errors.New("La causa de la muerte no es válida")
	ErrInvalidTurn              = errors.New("El turno no es válido")
	ErrInvalidNote              = errors.New("La nota debe tener como máximo 500 caracteres")
	ErrInvalidCreatedBy         = errors.New("El usuario que crea la muerte es obligatorio")
	ErrInvalidUpdatedBy         = errors.New("El usuario que actualiza la muerte es obligatorio")
	ErrDeathDateBeforeFarrowing = errors.New("La fecha de la muerte debe ser posterior al parto")
	ErrDeathDateInFuture        = errors.New("La fecha de la muerte no puede ser futura")
)

// PigletDeath is the aggregate root that represents the death of piglets in a
// farrowing. It records the quantity, cause, shift and operator of the event.
type PigletDeath struct {
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
// It requires the death date to be strictly after farrowDate and not later than
// today, then sets the audit fields to the creator and now. The weight is
// optional but must be greater than zero when provided.
func NewPigletDeath(params NewPigletDeathParams, farrowDate time.Time, now time.Time) (*PigletDeath, error) {
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

	if err := validateDeathDate(params.DeathDate, farrowDate, now); err != nil {
		return nil, err
	}

	if params.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}
	if params.Weight != nil && *params.Weight <= 0 {
		return nil, ErrInvalidWeight
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

// validateDeathDate checks the death date bounds against the farrow date.
// It rejects dates after today and dates on or before the farrow date.
func validateDeathDate(deathDate time.Time, farrowDate time.Time, now time.Time) error {
	death := truncateToDay(deathDate)
	if death.After(truncateToDay(now)) {
		return ErrDeathDateInFuture
	}
	if !death.After(truncateToDay(farrowDate)) {
		return ErrDeathDateBeforeFarrowing
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
