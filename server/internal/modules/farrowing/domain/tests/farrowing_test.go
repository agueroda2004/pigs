package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
)

func validParams() farrowingdomain.NewFarrowingParams {
	return farrowingdomain.NewFarrowingParams{
		ID:         uuid.New(),
		ServiceID:  uuid.New(),
		SowID:      uuid.New(),
		FarrowDate: time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC),
		CreatedBy:  uuid.New(),
	}
}

func TestNewFarrowing(t *testing.T) {
	now := time.Date(2026, time.February, 2, 3, 4, 5, 0, time.UTC)
	lastMount := time.Date(2025, time.October, 10, 0, 0, 0, 0, time.UTC)

	t.Run("creates a valid farrowing with defaults", func(t *testing.T) {
		farrowing, err := farrowingdomain.NewFarrowing(validParams(), nil, nil, lastMount, now)

		if err != nil {
			t.Fatalf("NewFarrowing() error = %v", err)
		}
		if farrowing.LiveBorn != 0 || farrowing.Stillborn != 0 || farrowing.Mummified != 0 {
			t.Fatalf("unexpected counts: %#v", farrowing)
		}
		if farrowing.IsManipulated {
			t.Fatalf("expected IsManipulated false")
		}
		if !farrowing.CreatedAt.Equal(now) || farrowing.CreatedBy != farrowing.UpdatedBy {
			t.Fatalf("unexpected audit fields: %#v", farrowing)
		}
	})

	t.Run("accepts a farrowing that crosses midnight", func(t *testing.T) {
		params := validParams()
		start := "22:00"
		end := "02:00"
		params.StartTime = &start
		params.EndTime = &end

		farrowing, err := farrowingdomain.NewFarrowing(params, nil, nil, lastMount, now)

		if err != nil {
			t.Fatalf("NewFarrowing() error = %v", err)
		}
		if !farrowing.CrossesMidnight() {
			t.Fatalf("expected CrossesMidnight true")
		}
	})

	t.Run("accepts valid times that do not cross midnight", func(t *testing.T) {
		params := validParams()
		start := "08:15"
		end := "12:45"
		params.StartTime = &start
		params.EndTime = &end

		farrowing, err := farrowingdomain.NewFarrowing(params, nil, nil, lastMount, now)

		if err != nil {
			t.Fatalf("NewFarrowing() error = %v", err)
		}
		if farrowing.CrossesMidnight() {
			t.Fatalf("expected CrossesMidnight false")
		}
	})

	t.Run("rejects invalid time format", func(t *testing.T) {
		params := validParams()
		start := "25:00"
		params.StartTime = &start

		_, err := farrowingdomain.NewFarrowing(params, nil, nil, lastMount, now)

		if !errors.Is(err, farrowingdomain.ErrInvalidStartTime) {
			t.Fatalf("error = %v, want ErrInvalidStartTime", err)
		}
	})

	t.Run("rejects nil id", func(t *testing.T) {
		params := validParams()
		params.ID = uuid.Nil

		_, err := farrowingdomain.NewFarrowing(params, nil, nil, lastMount, now)

		if !errors.Is(err, farrowingdomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("rejects nil service", func(t *testing.T) {
		params := validParams()
		params.ServiceID = uuid.Nil

		_, err := farrowingdomain.NewFarrowing(params, nil, nil, lastMount, now)

		if !errors.Is(err, farrowingdomain.ErrInvalidService) {
			t.Fatalf("error = %v, want ErrInvalidService", err)
		}
	})

	t.Run("rejects nil sow", func(t *testing.T) {
		params := validParams()
		params.SowID = uuid.Nil

		_, err := farrowingdomain.NewFarrowing(params, nil, nil, lastMount, now)

		if !errors.Is(err, farrowingdomain.ErrInvalidSow) {
			t.Fatalf("error = %v, want ErrInvalidSow", err)
		}
	})

	t.Run("rejects nil created by", func(t *testing.T) {
		params := validParams()
		params.CreatedBy = uuid.Nil

		_, err := farrowingdomain.NewFarrowing(params, nil, nil, lastMount, now)

		if !errors.Is(err, farrowingdomain.ErrInvalidCreatedBy) {
			t.Fatalf("error = %v, want ErrInvalidCreatedBy", err)
		}
	})

	t.Run("rejects a farrow date in the future", func(t *testing.T) {
		params := validParams()
		params.FarrowDate = now.AddDate(0, 0, 1)

		_, err := farrowingdomain.NewFarrowing(params, nil, nil, lastMount, now)

		if !errors.Is(err, farrowingdomain.ErrFarrowDateInFuture) {
			t.Fatalf("error = %v, want ErrFarrowDateInFuture", err)
		}
	})

	t.Run("rejects a farrow date before the last mount", func(t *testing.T) {
		params := validParams()
		params.FarrowDate = lastMount.AddDate(0, 0, -1)

		_, err := farrowingdomain.NewFarrowing(params, nil, nil, lastMount, now)

		if !errors.Is(err, farrowingdomain.ErrFarrowDateBeforeMount) {
			t.Fatalf("error = %v, want ErrFarrowDateBeforeMount", err)
		}
	})

	t.Run("rejects negative counts", func(t *testing.T) {
		for _, test := range []struct {
			name string
			set  func(*farrowingdomain.NewFarrowingParams)
			want error
		}{
			{"live born", func(p *farrowingdomain.NewFarrowingParams) { p.LiveBorn = -1 }, farrowingdomain.ErrInvalidLiveBorn},
			{"stillborn", func(p *farrowingdomain.NewFarrowingParams) { p.Stillborn = -1 }, farrowingdomain.ErrInvalidStillborn},
			{"mummified", func(p *farrowingdomain.NewFarrowingParams) { p.Mummified = -1 }, farrowingdomain.ErrInvalidMummified},
		} {
			t.Run(test.name, func(t *testing.T) {
				params := validParams()
				test.set(&params)

				_, err := farrowingdomain.NewFarrowing(params, nil, nil, lastMount, now)

				if !errors.Is(err, test.want) {
					t.Fatalf("error = %v, want %v", err, test.want)
				}
			})
		}
	})

	t.Run("rejects negative weights", func(t *testing.T) {
		negative := -1.5
		params := validParams()
		params.LitterWeight = &negative

		_, err := farrowingdomain.NewFarrowing(params, nil, nil, lastMount, now)

		if !errors.Is(err, farrowingdomain.ErrInvalidLitterWeight) {
			t.Fatalf("error = %v, want ErrInvalidLitterWeight", err)
		}

		params = validParams()
		params.StillbornWeight = &negative

		_, err = farrowingdomain.NewFarrowing(params, nil, nil, lastMount, now)

		if !errors.Is(err, farrowingdomain.ErrInvalidStillbornWeight) {
			t.Fatalf("error = %v, want ErrInvalidStillbornWeight", err)
		}
	})

	t.Run("builds operators and medications", func(t *testing.T) {
		operator := farrowingdomain.NewFarrowingOperatorParams{ID: uuid.New(), OperatorID: uuid.New()}
		medication := farrowingdomain.NewFarrowingMedicationParams{
			ID: uuid.New(), MedicationID: uuid.New(), Dose: 2, AppliedBy: uuid.New(),
		}

		farrowing, err := farrowingdomain.NewFarrowing(validParams(), []farrowingdomain.NewFarrowingOperatorParams{operator}, []farrowingdomain.NewFarrowingMedicationParams{medication}, lastMount, now)

		if err != nil {
			t.Fatalf("NewFarrowing() error = %v", err)
		}
		if len(farrowing.Operators) != 1 || len(farrowing.Medications) != 1 {
			t.Fatalf("unexpected children: %#v", farrowing)
		}
		if farrowing.Operators[0].FarrowingID != farrowing.ID || farrowing.Medications[0].FarrowingID != farrowing.ID {
			t.Fatalf("children not linked to the farrowing: %#v", farrowing)
		}
	})

	t.Run("rejects duplicate operators", func(t *testing.T) {
		operatorID := uuid.New()
		operators := []farrowingdomain.NewFarrowingOperatorParams{
			{ID: uuid.New(), OperatorID: operatorID},
			{ID: uuid.New(), OperatorID: operatorID},
		}

		_, err := farrowingdomain.NewFarrowing(validParams(), operators, nil, lastMount, now)

		if !errors.Is(err, farrowingdomain.ErrDuplicateOperator) {
			t.Fatalf("error = %v, want ErrDuplicateOperator", err)
		}
	})

	t.Run("rejects duplicate medications", func(t *testing.T) {
		medicationID := uuid.New()
		medications := []farrowingdomain.NewFarrowingMedicationParams{
			{ID: uuid.New(), MedicationID: medicationID, Dose: 1, AppliedBy: uuid.New()},
			{ID: uuid.New(), MedicationID: medicationID, Dose: 2, AppliedBy: uuid.New()},
		}

		_, err := farrowingdomain.NewFarrowing(validParams(), nil, medications, lastMount, now)

		if !errors.Is(err, farrowingdomain.ErrDuplicateMedication) {
			t.Fatalf("error = %v, want ErrDuplicateMedication", err)
		}
	})
}

func TestNewFarrowingOperator(t *testing.T) {
	now := time.Date(2026, time.February, 2, 3, 4, 5, 0, time.UTC)
	farrowingID := uuid.New()

	t.Run("creates a valid link", func(t *testing.T) {
		operatorID := uuid.New()
		operator, err := farrowingdomain.NewFarrowingOperator(
			farrowingdomain.NewFarrowingOperatorParams{ID: uuid.New(), OperatorID: operatorID},
			farrowingID,
			now,
		)

		if err != nil {
			t.Fatalf("NewFarrowingOperator() error = %v", err)
		}
		if operator.OperatorID != operatorID || operator.FarrowingID != farrowingID || !operator.CreatedAt.Equal(now) {
			t.Fatalf("unexpected operator link: %#v", operator)
		}
	})

	t.Run("rejects nil operator", func(t *testing.T) {
		_, err := farrowingdomain.NewFarrowingOperator(
			farrowingdomain.NewFarrowingOperatorParams{ID: uuid.New(), OperatorID: uuid.Nil},
			farrowingID,
			now,
		)

		if !errors.Is(err, farrowingdomain.ErrInvalidFarrowingOperatorOperator) {
			t.Fatalf("error = %v, want ErrInvalidFarrowingOperatorOperator", err)
		}
	})
}

func TestNewFarrowingMedication(t *testing.T) {
	now := time.Date(2026, time.February, 2, 3, 4, 5, 0, time.UTC)
	farrowingID := uuid.New()

	t.Run("creates a valid link", func(t *testing.T) {
		medicationID := uuid.New()
		appliedBy := uuid.New()
		medication, err := farrowingdomain.NewFarrowingMedication(
			farrowingdomain.NewFarrowingMedicationParams{
				ID: uuid.New(), MedicationID: medicationID, Dose: 3, AppliedBy: appliedBy,
			},
			farrowingID,
			now,
		)

		if err != nil {
			t.Fatalf("NewFarrowingMedication() error = %v", err)
		}
		if medication.MedicationID != medicationID || medication.Dose != 3 || medication.AppliedBy != appliedBy {
			t.Fatalf("unexpected medication link: %#v", medication)
		}
	})

	t.Run("accepts a decimal dose", func(t *testing.T) {
		medication, err := farrowingdomain.NewFarrowingMedication(
			farrowingdomain.NewFarrowingMedicationParams{
				ID: uuid.New(), MedicationID: uuid.New(), Dose: 0.5, AppliedBy: uuid.New(),
			},
			farrowingID,
			now,
		)

		if err != nil {
			t.Fatalf("NewFarrowingMedication() error = %v", err)
		}
		if medication.Dose != 0.5 {
			t.Fatalf("dose = %v, want 0.5", medication.Dose)
		}
	})

	t.Run("rejects a non-positive dose", func(t *testing.T) {
		_, err := farrowingdomain.NewFarrowingMedication(
			farrowingdomain.NewFarrowingMedicationParams{
				ID: uuid.New(), MedicationID: uuid.New(), Dose: 0, AppliedBy: uuid.New(),
			},
			farrowingID,
			now,
		)

		if !errors.Is(err, farrowingdomain.ErrInvalidFarrowingMedicationDose) {
			t.Fatalf("error = %v, want ErrInvalidFarrowingMedicationDose", err)
		}
	})
}
