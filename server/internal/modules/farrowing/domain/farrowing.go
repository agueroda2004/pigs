package farrowing

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	maxLocationLength = 100
	maxNoteLength     = 500
	maxTimeLength     = 5
)

var timeLayout = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

var (
	ErrInvalidID              = errors.New("El identificador del parto es obligatorio")
	ErrInvalidService         = errors.New("El servicio del parto es obligatorio")
	ErrInvalidSow             = errors.New("La cerda del parto es obligatoria")
	ErrInvalidFarrowDate      = errors.New("La fecha del parto es obligatoria")
	ErrInvalidLastMount       = errors.New("La fecha de la última monta es obligatoria")
	ErrInvalidStartTime       = errors.New("La hora de inicio no es válida")
	ErrInvalidEndTime         = errors.New("La hora de fin no es válida")
	ErrInvalidLocation        = errors.New("La ubicación debe tener como máximo 100 caracteres")
	ErrInvalidLiveBorn        = errors.New("La cantidad de nacidos vivos no puede ser negativa")
	ErrInvalidStillborn       = errors.New("La cantidad de nacidos muertos no puede ser negativa")
	ErrInvalidMummified       = errors.New("La cantidad de momificados no puede ser negativa")
	ErrInvalidLitterWeight    = errors.New("El peso de la camada no puede ser negativo")
	ErrInvalidStillbornWeight = errors.New("El peso de los nacidos muertos no puede ser negativo")
	ErrInvalidNote            = errors.New("La nota debe tener como máximo 500 caracteres")
	ErrInvalidNurseStartDate  = errors.New("La fecha de inicio de nodriza no es válida")
	ErrInvalidCreatedBy       = errors.New("El usuario que crea el parto es obligatorio")
	ErrInvalidUpdatedBy       = errors.New("El usuario que actualiza el parto es obligatorio")
	ErrFarrowDateInFuture     = errors.New("La fecha del parto no puede ser futura")
	ErrFarrowDateBeforeMount  = errors.New("La fecha del parto debe ser posterior a la última monta")
	ErrDuplicateOperator      = errors.New("No se puede repetir un operador en el parto")
	ErrDuplicateMedication    = errors.New("No se puede repetir un medicamento en el parto")
	ErrInvalidPigletQuantity  = errors.New("La cantidad de lechones debe ser mayor a cero")
	ErrInsufficientPiglets    = errors.New("La cantidad supera los lechones actuales")
)

// Farrowing is the aggregate root that represents a sow farrowing (parto).
// It records the reproductive result and groups the operators and medications
// involved in the event.
type Farrowing struct {
	ID              uuid.UUID
	ServiceID       uuid.UUID
	SowID           uuid.UUID
	FarrowDate      time.Time
	StartTime       *string
	EndTime         *string
	Location        *string
	LiveBorn        int
	Stillborn       int
	Mummified       int
	CurrentPiglets  int
	LitterWeight    *float64
	StillbornWeight *float64
	IsManipulated   bool
	IsNurse         bool
	NurseStartDate  *time.Time
	Note            *string
	Operators       []*FarrowingOperator
	Medications     []*FarrowingMedication
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedBy       uuid.UUID
	UpdatedBy       uuid.UUID
}

// NewFarrowingParams holds the fields required to build a new farrowing.
// Optional values are pointers and may be nil; the counts default to zero and
// the manipulation flag defaults to false.
type NewFarrowingParams struct {
	ID              uuid.UUID
	ServiceID       uuid.UUID
	SowID           uuid.UUID
	FarrowDate      time.Time
	StartTime       *string
	EndTime         *string
	Location        *string
	LiveBorn        int
	Stillborn       int
	Mummified       int
	LitterWeight    *float64
	StillbornWeight *float64
	IsManipulated   bool
	Note            *string
	CreatedBy       uuid.UUID
}

// NewFarrowing builds a farrowing with its operators and medications after
// validating every field. It rejects a farrow date in the future or not later
// than the last mount date, validates time formats without comparing them (so
// overnight farrowings are allowed) and rejects duplicate operators or
// medications. It sets the audit fields to the creator and both timestamps to now.
// The current piglets balance starts at the live born count.
func NewFarrowing(
	params NewFarrowingParams,
	operators []NewFarrowingOperatorParams,
	medications []NewFarrowingMedicationParams,
	lastMountDate time.Time,
	now time.Time,
) (*Farrowing, error) {
	if params.ID == uuid.Nil {
		return nil, ErrInvalidID
	}
	if params.ServiceID == uuid.Nil {
		return nil, ErrInvalidService
	}
	if params.SowID == uuid.Nil {
		return nil, ErrInvalidSow
	}
	if params.FarrowDate.IsZero() {
		return nil, ErrInvalidFarrowDate
	}
	if lastMountDate.IsZero() {
		return nil, ErrInvalidLastMount
	}
	if err := validateFarrowDate(params.FarrowDate, lastMountDate, now); err != nil {
		return nil, err
	}

	startTime, err := validateTime(params.StartTime, ErrInvalidStartTime)
	if err != nil {
		return nil, err
	}
	endTime, err := validateTime(params.EndTime, ErrInvalidEndTime)
	if err != nil {
		return nil, err
	}

	location, err := validateLocation(params.Location)
	if err != nil {
		return nil, err
	}

	note, err := validateNote(params.Note)
	if err != nil {
		return nil, err
	}

	if params.LiveBorn < 0 {
		return nil, ErrInvalidLiveBorn
	}
	if params.Stillborn < 0 {
		return nil, ErrInvalidStillborn
	}
	if params.Mummified < 0 {
		return nil, ErrInvalidMummified
	}
	if params.LitterWeight != nil && *params.LitterWeight < 0 {
		return nil, ErrInvalidLitterWeight
	}
	if params.StillbornWeight != nil && *params.StillbornWeight < 0 {
		return nil, ErrInvalidStillbornWeight
	}

	if params.CreatedBy == uuid.Nil {
		return nil, ErrInvalidCreatedBy
	}

	builtOperators := make([]*FarrowingOperator, 0, len(operators))
	seenOperators := make(map[uuid.UUID]struct{}, len(operators))
	for _, operatorParams := range operators {
		operator, err := NewFarrowingOperator(operatorParams, params.ID, now)
		if err != nil {
			return nil, err
		}
		if _, exists := seenOperators[operator.OperatorID]; exists {
			return nil, ErrDuplicateOperator
		}
		seenOperators[operator.OperatorID] = struct{}{}
		builtOperators = append(builtOperators, operator)
	}

	builtMedications := make([]*FarrowingMedication, 0, len(medications))
	seenMedications := make(map[uuid.UUID]struct{}, len(medications))
	for _, medicationParams := range medications {
		medication, err := NewFarrowingMedication(medicationParams, params.ID, now)
		if err != nil {
			return nil, err
		}
		if _, exists := seenMedications[medication.MedicationID]; exists {
			return nil, ErrDuplicateMedication
		}
		seenMedications[medication.MedicationID] = struct{}{}
		builtMedications = append(builtMedications, medication)
	}

	return &Farrowing{
		ID:              params.ID,
		ServiceID:       params.ServiceID,
		SowID:           params.SowID,
		FarrowDate:      params.FarrowDate,
		StartTime:       startTime,
		EndTime:         endTime,
		Location:        location,
		LiveBorn:        params.LiveBorn,
		Stillborn:       params.Stillborn,
		Mummified:       params.Mummified,
		CurrentPiglets:  params.LiveBorn,
		LitterWeight:    params.LitterWeight,
		StillbornWeight: params.StillbornWeight,
		IsManipulated:   params.IsManipulated,
		Note:            note,
		Operators:       builtOperators,
		Medications:     builtMedications,
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedBy:       params.CreatedBy,
		UpdatedBy:       params.CreatedBy,
	}, nil
}

// CrossesMidnight reports whether the farrowing ended after midnight relative
// to its start time. It returns false when either time is missing or invalid
// and never validates the ordering, because overnight farrowings are valid.
func (f *Farrowing) CrossesMidnight() bool {
	if f == nil || f.StartTime == nil || f.EndTime == nil {
		return false
	}
	start, err := time.Parse("15:04", *f.StartTime)
	if err != nil {
		return false
	}
	end, err := time.Parse("15:04", *f.EndTime)
	if err != nil {
		return false
	}
	return end.Before(start)
}

// ReduceCurrentPiglets subtracts the given quantity from the current balance.
// It rejects non-positive quantities and quantities greater than the current
// balance, then records updatedBy plus the timestamp.
func (f *Farrowing) ReduceCurrentPiglets(quantity int, updatedBy uuid.UUID, now time.Time) error {
	if f == nil || f.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if quantity <= 0 {
		return ErrInvalidPigletQuantity
	}
	if quantity > f.CurrentPiglets {
		return ErrInsufficientPiglets
	}

	f.CurrentPiglets -= quantity
	f.UpdatedAt = now
	f.UpdatedBy = updatedBy
	return nil
}

// AddCurrentPiglets adds the given quantity to the current balance.
// It rejects non-positive quantities, then records updatedBy plus the timestamp.
func (f *Farrowing) AddCurrentPiglets(quantity int, updatedBy uuid.UUID, now time.Time) error {
	if f == nil || f.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if quantity <= 0 {
		return ErrInvalidPigletQuantity
	}

	f.CurrentPiglets += quantity
	f.UpdatedAt = now
	f.UpdatedBy = updatedBy
	return nil
}

// MarkAsNurse flags the farrowing as a nurse and records its start date.
// It rejects a zero date or missing actor, then sets the nurse fields plus the
// update metadata so the farrowing can receive donated piglets.
func (f *Farrowing) MarkAsNurse(startDate time.Time, updatedBy uuid.UUID, now time.Time) error {
	if f == nil || f.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if startDate.IsZero() {
		return ErrInvalidNurseStartDate
	}

	normalized := truncateToDay(startDate)
	f.IsNurse = true
	f.NurseStartDate = &normalized
	f.UpdatedAt = now
	f.UpdatedBy = updatedBy
	return nil
}

// validateFarrowDate checks the farrow date bounds against the last mount.
// It rejects dates after today and dates on or before the last mount date.
func validateFarrowDate(farrowDate time.Time, lastMountDate time.Time, now time.Time) error {
	farrow := truncateToDay(farrowDate)
	if farrow.After(truncateToDay(now)) {
		return ErrFarrowDateInFuture
	}
	if !farrow.After(truncateToDay(lastMountDate)) {
		return ErrFarrowDateBeforeMount
	}
	return nil
}

// truncateToDay removes the time portion from a timestamp in UTC.
// It is used to compare the farrow and mount dates at day granularity.
func truncateToDay(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

// validateTime trims an optional HH:MM time and checks its format.
// It returns nil when the value is nil or empty (clear to NULL) and the given
// error when the value does not match the 24h HH:MM pattern.
func validateTime(value *string, invalid error) (*string, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	if len(trimmed) > maxTimeLength || !timeLayout.MatchString(trimmed) {
		return nil, invalid
	}
	return &trimmed, nil
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
