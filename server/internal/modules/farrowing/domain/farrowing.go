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
	maxPigletCount    = 30
	maxWeight         = 30.0
	maxMedicationDose = 30.0
)

var timeLayout = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

var (
	ErrInvalidID         = errors.New("El identificador del parto es obligatorio")
	ErrInvalidService    = errors.New("El servicio del parto es obligatorio")
	ErrInvalidSow        = errors.New("La cerda del parto es obligatoria")
	ErrInvalidFarrowDate = errors.New("La fecha del parto es obligatoria")
	ErrInvalidLastMount  = errors.New("La fecha de la última monta es obligatoria")
	ErrInvalidStartTime  = errors.New("La hora de inicio no es válida")
	ErrInvalidEndTime    = errors.New("La hora de fin no es válida")
	ErrInvalidLocation   = errors.New("La ubicación debe tener como máximo 100 caracteres")
	ErrInvalidLiveBorn   = errors.New("La cantidad de nacidos vivos no puede ser negativa")
	ErrInvalidStillborn  = errors.New("La cantidad de nacidos muertos no puede ser negativa")
	ErrInvalidMummified  = errors.New("La cantidad de momificados no puede ser negativa")
	// ErrLiveBornTooHigh is returned when the live born count is greater than the limit.
	ErrLiveBornTooHigh = errors.New("La cantidad de nacidos vivos no puede ser mayor a 30")
	// ErrStillbornTooHigh is returned when the stillborn count is greater than the limit.
	ErrStillbornTooHigh = errors.New("La cantidad de nacidos muertos no puede ser mayor a 30")
	// ErrMummifiedTooHigh is returned when the mummified count is greater than the limit.
	ErrMummifiedTooHigh = errors.New("La cantidad de momificados no puede ser mayor a 30")
	// ErrLiveBornRequired is returned when the create request omits the live born count.
	ErrLiveBornRequired = errors.New("La cantidad de nacidos vivos es obligatoria")
	// ErrStillbornRequired is returned when the create request omits the stillborn count.
	ErrStillbornRequired = errors.New("La cantidad de nacidos muertos es obligatoria")
	// ErrMummifiedRequired is returned when the create request omits the mummified count.
	ErrMummifiedRequired = errors.New("La cantidad de momificados es obligatoria")
	// ErrLitterWeightRequired is returned when the create request omits the litter weight.
	ErrLitterWeightRequired = errors.New("El peso de la camada es obligatorio")
	// ErrStillbornWeightRequired is returned when the create request omits the stillborn weight.
	ErrStillbornWeightRequired = errors.New("El peso de los nacidos muertos es obligatorio")
	// ErrLitterWeightTooHigh is returned when the litter weight is greater than the limit.
	ErrLitterWeightTooHigh = errors.New("El peso de la camada no puede ser mayor a 30 kg")
	// ErrStillbornWeightTooHigh is returned when the stillborn weight is greater than the limit.
	ErrStillbornWeightTooHigh = errors.New("El peso de los nacidos muertos no puede ser mayor a 30 kg")
	// ErrMedicationDoseTooHigh is returned when a medication dose is greater than the limit.
	ErrMedicationDoseTooHigh   = errors.New("La dosis del medicamento no puede ser mayor a 30")
	ErrInvalidLitterWeight     = errors.New("El peso de la camada no puede ser negativo")
	ErrInvalidStillbornWeight  = errors.New("El peso de los nacidos muertos no puede ser negativo")
	ErrInvalidNote             = errors.New("La nota debe tener como máximo 500 caracteres")
	ErrInvalidNurseStartDate   = errors.New("La fecha de inicio de nodriza no es válida")
	ErrInvalidCreatedBy        = errors.New("El usuario que crea el parto es obligatorio")
	ErrInvalidUpdatedBy        = errors.New("El usuario que actualiza el parto es obligatorio")
	ErrFarrowDateInFuture      = errors.New("La fecha del parto no puede ser futura")
	ErrFarrowDateBeforeMount   = errors.New("La fecha del parto debe ser posterior a la última monta")
	ErrDuplicateOperator       = errors.New("No se puede repetir un operador en el parto")
	ErrDuplicateMedication     = errors.New("No se puede repetir un medicamento en el parto")
	ErrInvalidPigletQuantity   = errors.New("La cantidad de lechones debe ser mayor a cero")
	ErrInsufficientPiglets     = errors.New("La cantidad supera los lechones actuales")
	ErrWeagingQuantityMismatch = errors.New("La cantidad del destete debe coincidir con los lechones actuales")
	// ErrInvalidUpdate is returned when an update request carries no editable field.
	ErrInvalidUpdate = errors.New("Debe actualizar al menos un campo del parto")
	// ErrFarrowDateAfterRelatedRecords is returned when the new farrow date falls
	// after the earliest record (death, fostering or weaging) linked to the farrowing.
	ErrFarrowDateAfterRelatedRecords = errors.New("La fecha del parto no puede ser posterior a un registro de lechones")
	// ErrLiveBornLocked is returned when the live born count is changed on a
	// farrowing that already has related records, because its current piglets
	// balance has since been adjusted and could no longer be recomputed.
	ErrLiveBornLocked = errors.New("No se puede cambiar la cantidad de nacidos vivos porque el parto tiene registros enlazados")
)

// Farrowing is the aggregate root that represents a sow farrowing (parto).
// It records the reproductive result and groups the operators and medications
// involved in the event.
type Farrowing struct {
	ID              uuid.UUID
	ServiceID       uuid.UUID
	SowID           uuid.UUID
	SowCode         string
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
// Optional values are pointers and may be nil; both weights are required, the
// counts default to zero and the manipulation flag defaults to false.
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
// overnight farrowings are allowed), requires both weights, rejects any piglet
// count, weight or medication dose over 30 and rejects duplicate operators or
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
	if params.LiveBorn > maxPigletCount {
		return nil, ErrLiveBornTooHigh
	}
	if params.Stillborn < 0 {
		return nil, ErrInvalidStillborn
	}
	if params.Stillborn > maxPigletCount {
		return nil, ErrStillbornTooHigh
	}
	if params.Mummified < 0 {
		return nil, ErrInvalidMummified
	}
	if params.Mummified > maxPigletCount {
		return nil, ErrMummifiedTooHigh
	}
	if params.LitterWeight == nil {
		return nil, ErrLitterWeightRequired
	}
	if params.StillbornWeight == nil {
		return nil, ErrStillbornWeightRequired
	}
	if *params.LitterWeight < 0 {
		return nil, ErrInvalidLitterWeight
	}
	if *params.LitterWeight > maxWeight {
		return nil, ErrLitterWeightTooHigh
	}
	if *params.StillbornWeight < 0 {
		return nil, ErrInvalidStillbornWeight
	}
	if *params.StillbornWeight > maxWeight {
		return nil, ErrStillbornWeightTooHigh
	}

	if params.CreatedBy == uuid.Nil {
		return nil, ErrInvalidCreatedBy
	}

	if err := validateMedicationDoses(medications); err != nil {
		return nil, err
	}

	builtOperators, err := buildOperators(operators, params.ID, now)
	if err != nil {
		return nil, err
	}

	builtMedications, err := buildMedications(medications, params.ID, now)
	if err != nil {
		return nil, err
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

// UpdateFarrowingParams holds the editable fields of a farrowing update.
// A nil pointer means the field is omitted and stays unchanged. Nullable string
// fields are cleared to NULL by providing an empty string, while the nullable
// weights use their Clear flag because they cannot carry an empty value. The
// operator and medication lists are only replaced when their Update flag is set;
// an empty provided list clears every link. Sow, service, state and the nurse
// fields are intentionally absent because they are not editable here.
type UpdateFarrowingParams struct {
	FarrowDate           *time.Time
	StartTime            *string
	EndTime              *string
	Location             *string
	LiveBorn             *int
	Stillborn            *int
	Mummified            *int
	LitterWeight         *float64
	ClearLitterWeight    bool
	StillbornWeight      *float64
	ClearStillbornWeight bool
	IsManipulated        *bool
	Note                 *string
	Operators            []NewFarrowingOperatorParams
	UpdateOperators      bool
	Medications          []NewFarrowingMedicationParams
	UpdateMedications    bool
}

// UpdateFarrowingReference carries the validation context of an update.
// LastMountDate bounds the earliest valid farrow date and EarliestRelatedDate is
// the earliest linked death, fostering or weaging date; it also signals that the
// farrowing has related records when it is not nil.
type UpdateFarrowingReference struct {
	LastMountDate       time.Time
	EarliestRelatedDate *time.Time
}

// Update applies only the provided fields to the farrowing.
// It validates the farrow date against the last mount, the future and the
// earliest related record, locks the live born count when related records exist,
// recomputes the current piglets balance when the live born count changes without
// related records, and replaces the operator and medication links when requested.
// It records updatedBy plus the timestamp and returns ErrInvalidUpdate when no
// field is provided.
func (f *Farrowing) Update(params UpdateFarrowingParams, reference UpdateFarrowingReference, updatedBy uuid.UUID, now time.Time) error {
	if f == nil || f.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if params.FarrowDate == nil && params.StartTime == nil && params.EndTime == nil &&
		params.Location == nil && params.LiveBorn == nil && params.Stillborn == nil &&
		params.Mummified == nil && params.LitterWeight == nil && !params.ClearLitterWeight &&
		params.StillbornWeight == nil && !params.ClearStillbornWeight &&
		params.IsManipulated == nil && params.Note == nil &&
		!params.UpdateOperators && !params.UpdateMedications {
		return ErrInvalidUpdate
	}

	if params.FarrowDate != nil {
		if params.FarrowDate.IsZero() {
			return ErrInvalidFarrowDate
		}
		if err := validateFarrowDate(*params.FarrowDate, reference.LastMountDate, now); err != nil {
			return err
		}
		if reference.EarliestRelatedDate != nil &&
			truncateToDay(*params.FarrowDate).After(truncateToDay(*reference.EarliestRelatedDate)) {
			return ErrFarrowDateAfterRelatedRecords
		}
	}

	if params.LiveBorn != nil {
		if *params.LiveBorn < 0 {
			return ErrInvalidLiveBorn
		}
		if reference.EarliestRelatedDate != nil && *params.LiveBorn != f.LiveBorn {
			return ErrLiveBornLocked
		}
	}
	if params.Stillborn != nil && *params.Stillborn < 0 {
		return ErrInvalidStillborn
	}
	if params.Mummified != nil && *params.Mummified < 0 {
		return ErrInvalidMummified
	}
	if params.LitterWeight != nil && !params.ClearLitterWeight && *params.LitterWeight < 0 {
		return ErrInvalidLitterWeight
	}
	if params.StillbornWeight != nil && !params.ClearStillbornWeight && *params.StillbornWeight < 0 {
		return ErrInvalidStillbornWeight
	}

	startTime, err := validateTime(params.StartTime, ErrInvalidStartTime)
	if err != nil {
		return err
	}
	endTime, err := validateTime(params.EndTime, ErrInvalidEndTime)
	if err != nil {
		return err
	}
	location, err := validateLocation(params.Location)
	if err != nil {
		return err
	}
	note, err := validateNote(params.Note)
	if err != nil {
		return err
	}

	var replacedOperators []*FarrowingOperator
	if params.UpdateOperators {
		if replacedOperators, err = buildOperators(params.Operators, f.ID, now); err != nil {
			return err
		}
	}
	var replacedMedications []*FarrowingMedication
	if params.UpdateMedications {
		if replacedMedications, err = buildMedications(params.Medications, f.ID, now); err != nil {
			return err
		}
	}

	if params.FarrowDate != nil {
		f.FarrowDate = *params.FarrowDate
	}
	if params.StartTime != nil {
		f.StartTime = startTime
	}
	if params.EndTime != nil {
		f.EndTime = endTime
	}
	if params.Location != nil {
		f.Location = location
	}
	if params.LiveBorn != nil {
		f.LiveBorn = *params.LiveBorn
		if reference.EarliestRelatedDate == nil {
			f.CurrentPiglets = *params.LiveBorn
		}
	}
	if params.Stillborn != nil {
		f.Stillborn = *params.Stillborn
	}
	if params.Mummified != nil {
		f.Mummified = *params.Mummified
	}
	switch {
	case params.ClearLitterWeight:
		f.LitterWeight = nil
	case params.LitterWeight != nil:
		f.LitterWeight = params.LitterWeight
	}
	switch {
	case params.ClearStillbornWeight:
		f.StillbornWeight = nil
	case params.StillbornWeight != nil:
		f.StillbornWeight = params.StillbornWeight
	}
	if params.IsManipulated != nil {
		f.IsManipulated = *params.IsManipulated
	}
	if params.Note != nil {
		f.Note = note
	}
	if params.UpdateOperators {
		f.Operators = replacedOperators
	}
	if params.UpdateMedications {
		f.Medications = replacedMedications
	}

	f.UpdatedAt = now
	f.UpdatedBy = updatedBy
	return nil
}

// validateMedicationDoses checks that every medication dose is at most 30.
// It returns ErrMedicationDoseTooHigh for the first dose above the limit.
func validateMedicationDoses(medications []NewFarrowingMedicationParams) error {
	for _, medication := range medications {
		if medication.Dose > maxMedicationDose {
			return ErrMedicationDoseTooHigh
		}
	}
	return nil
}

// buildOperators validates and links the given operator params to a farrowing.
// It rejects duplicate operators and returns the built link rows.
func buildOperators(params []NewFarrowingOperatorParams, farrowingID uuid.UUID, now time.Time) ([]*FarrowingOperator, error) {
	operators := make([]*FarrowingOperator, 0, len(params))
	seen := make(map[uuid.UUID]struct{}, len(params))
	for _, operatorParams := range params {
		operator, err := NewFarrowingOperator(operatorParams, farrowingID, now)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[operator.OperatorID]; exists {
			return nil, ErrDuplicateOperator
		}
		seen[operator.OperatorID] = struct{}{}
		operators = append(operators, operator)
	}
	return operators, nil
}

// buildMedications validates and links the given medication params to a farrowing.
// It rejects duplicate medications and returns the built link rows.
func buildMedications(params []NewFarrowingMedicationParams, farrowingID uuid.UUID, now time.Time) ([]*FarrowingMedication, error) {
	medications := make([]*FarrowingMedication, 0, len(params))
	seen := make(map[uuid.UUID]struct{}, len(params))
	for _, medicationParams := range params {
		medication, err := NewFarrowingMedication(medicationParams, farrowingID, now)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[medication.MedicationID]; exists {
			return nil, ErrDuplicateMedication
		}
		seen[medication.MedicationID] = struct{}{}
		medications = append(medications, medication)
	}
	return medications, nil
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

// WeanAll completes the weaning of the farrowing by zeroing its current balance.
// It rejects non-positive or mismatched quantities and missing actors, then sets
// the balance to zero and records updatedBy plus the timestamp.
func (f *Farrowing) WeanAll(quantity int, updatedBy uuid.UUID, now time.Time) error {
	if f == nil || f.ID == uuid.Nil {
		return ErrInvalidID
	}
	if updatedBy == uuid.Nil {
		return ErrInvalidUpdatedBy
	}
	if quantity <= 0 {
		return ErrInvalidPigletQuantity
	}
	if quantity != f.CurrentPiglets {
		return ErrWeagingQuantityMismatch
	}

	f.CurrentPiglets = 0
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
