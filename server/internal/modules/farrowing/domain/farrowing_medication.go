package farrowing

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidFarrowingMedicationID         = errors.New("El identificador del medicamento del parto es obligatorio")
	ErrInvalidFarrowingMedicationFarrowing  = errors.New("El parto del medicamento es obligatorio")
	ErrInvalidFarrowingMedicationMedication = errors.New("El medicamento del parto es obligatorio")
	ErrInvalidFarrowingMedicationDose       = errors.New("La dosis del medicamento debe ser mayor que cero")
	ErrInvalidFarrowingMedicationAppliedBy  = errors.New("El operador que aplica el medicamento es obligatorio")
)

// FarrowingMedication links a farrowing with a medication applied during it.
// It records the dose and the operator that applied the medication.
type FarrowingMedication struct {
	ID           uuid.UUID
	FarrowingID  uuid.UUID
	MedicationID uuid.UUID
	Dose         float64
	AppliedBy    uuid.UUID
	CreatedAt    time.Time
}

// NewFarrowingMedicationParams holds the fields required to build a join row.
// The farrowing id and the timestamp are supplied by the caller.
type NewFarrowingMedicationParams struct {
	ID           uuid.UUID
	MedicationID uuid.UUID
	Dose         float64
	AppliedBy    uuid.UUID
}

// NewFarrowingMedication builds a farrowing-medication link after validating its fields.
// It requires a positive dose and non-nil ids for the row, the farrowing, the
// medication and the applying operator.
func NewFarrowingMedication(params NewFarrowingMedicationParams, farrowingID uuid.UUID, now time.Time) (*FarrowingMedication, error) {
	if params.ID == uuid.Nil {
		return nil, ErrInvalidFarrowingMedicationID
	}
	if farrowingID == uuid.Nil {
		return nil, ErrInvalidFarrowingMedicationFarrowing
	}
	if params.MedicationID == uuid.Nil {
		return nil, ErrInvalidFarrowingMedicationMedication
	}
	if params.Dose <= 0 {
		return nil, ErrInvalidFarrowingMedicationDose
	}
	if params.AppliedBy == uuid.Nil {
		return nil, ErrInvalidFarrowingMedicationAppliedBy
	}

	return &FarrowingMedication{
		ID:           params.ID,
		FarrowingID:  farrowingID,
		MedicationID: params.MedicationID,
		Dose:         params.Dose,
		AppliedBy:    params.AppliedBy,
		CreatedAt:    now,
	}, nil
}
