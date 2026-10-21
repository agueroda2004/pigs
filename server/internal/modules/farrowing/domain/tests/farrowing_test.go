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
		if farrowing.CurrentPiglets != 0 {
			t.Fatalf("current piglets = %d, want 0", farrowing.CurrentPiglets)
		}
		if farrowing.IsManipulated {
			t.Fatalf("expected IsManipulated false")
		}
		if !farrowing.CreatedAt.Equal(now) || farrowing.CreatedBy != farrowing.UpdatedBy {
			t.Fatalf("unexpected audit fields: %#v", farrowing)
		}
	})

	t.Run("initializes current piglets from live born", func(t *testing.T) {
		params := validParams()
		params.LiveBorn = 10

		farrowing, err := farrowingdomain.NewFarrowing(params, nil, nil, lastMount, now)

		if err != nil {
			t.Fatalf("NewFarrowing() error = %v", err)
		}
		if farrowing.CurrentPiglets != 10 {
			t.Fatalf("current piglets = %d, want 10", farrowing.CurrentPiglets)
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

	t.Run("accepts counts up to 30", func(t *testing.T) {
		params := validParams()
		params.LiveBorn = 30
		params.Stillborn = 30
		params.Mummified = 30

		farrowing, err := farrowingdomain.NewFarrowing(params, nil, nil, lastMount, now)

		if err != nil {
			t.Fatalf("NewFarrowing() error = %v", err)
		}
		if farrowing.LiveBorn != 30 || farrowing.Stillborn != 30 || farrowing.Mummified != 30 {
			t.Fatalf("unexpected counts: %#v", farrowing)
		}
	})

	t.Run("rejects counts greater than 30", func(t *testing.T) {
		for _, test := range []struct {
			name string
			set  func(*farrowingdomain.NewFarrowingParams)
			want error
		}{
			{"live born", func(p *farrowingdomain.NewFarrowingParams) { p.LiveBorn = 31 }, farrowingdomain.ErrLiveBornTooHigh},
			{"stillborn", func(p *farrowingdomain.NewFarrowingParams) { p.Stillborn = 31 }, farrowingdomain.ErrStillbornTooHigh},
			{"mummified", func(p *farrowingdomain.NewFarrowingParams) { p.Mummified = 31 }, farrowingdomain.ErrMummifiedTooHigh},
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

func TestFarrowingReduceCurrentPiglets(t *testing.T) {
	now := time.Date(2026, time.February, 2, 3, 4, 5, 0, time.UTC)
	actor := uuid.New()

	t.Run("reduces the current piglets balance", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New(), CurrentPiglets: 10}

		if err := farrowing.ReduceCurrentPiglets(4, actor, now); err != nil {
			t.Fatalf("ReduceCurrentPiglets() error = %v", err)
		}
		if farrowing.CurrentPiglets != 6 {
			t.Fatalf("current piglets = %d, want 6", farrowing.CurrentPiglets)
		}
		if !farrowing.UpdatedAt.Equal(now) || farrowing.UpdatedBy != actor {
			t.Fatalf("unexpected audit fields: %#v", farrowing)
		}
	})

	t.Run("rejects a non-positive quantity", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New(), CurrentPiglets: 10}

		if err := farrowing.ReduceCurrentPiglets(0, actor, now); !errors.Is(err, farrowingdomain.ErrInvalidPigletQuantity) {
			t.Fatalf("error = %v, want ErrInvalidPigletQuantity", err)
		}
	})

	t.Run("rejects a quantity greater than the balance", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New(), CurrentPiglets: 3}

		if err := farrowing.ReduceCurrentPiglets(4, actor, now); !errors.Is(err, farrowingdomain.ErrInsufficientPiglets) {
			t.Fatalf("error = %v, want ErrInsufficientPiglets", err)
		}
	})

	t.Run("rejects a nil updated by", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New(), CurrentPiglets: 3}

		if err := farrowing.ReduceCurrentPiglets(1, uuid.Nil, now); !errors.Is(err, farrowingdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})
}

func TestFarrowingAddCurrentPiglets(t *testing.T) {
	now := time.Date(2026, time.February, 2, 3, 4, 5, 0, time.UTC)
	actor := uuid.New()

	t.Run("increases the current piglets balance", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New(), CurrentPiglets: 10}

		if err := farrowing.AddCurrentPiglets(4, actor, now); err != nil {
			t.Fatalf("AddCurrentPiglets() error = %v", err)
		}
		if farrowing.CurrentPiglets != 14 {
			t.Fatalf("current piglets = %d, want 14", farrowing.CurrentPiglets)
		}
		if !farrowing.UpdatedAt.Equal(now) || farrowing.UpdatedBy != actor {
			t.Fatalf("unexpected audit fields: %#v", farrowing)
		}
	})

	t.Run("rejects a non-positive quantity", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New(), CurrentPiglets: 10}

		if err := farrowing.AddCurrentPiglets(0, actor, now); !errors.Is(err, farrowingdomain.ErrInvalidPigletQuantity) {
			t.Fatalf("error = %v, want ErrInvalidPigletQuantity", err)
		}
	})

	t.Run("rejects a nil updated by", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New(), CurrentPiglets: 10}

		if err := farrowing.AddCurrentPiglets(1, uuid.Nil, now); !errors.Is(err, farrowingdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})
}

func TestFarrowingWeanAll(t *testing.T) {
	now := time.Date(2026, time.February, 2, 3, 4, 5, 0, time.UTC)
	actor := uuid.New()

	t.Run("zeroes the current piglets balance", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New(), CurrentPiglets: 8}

		if err := farrowing.WeanAll(8, actor, now); err != nil {
			t.Fatalf("WeanAll() error = %v", err)
		}
		if farrowing.CurrentPiglets != 0 {
			t.Fatalf("current piglets = %d, want 0", farrowing.CurrentPiglets)
		}
		if !farrowing.UpdatedAt.Equal(now) || farrowing.UpdatedBy != actor {
			t.Fatalf("unexpected audit fields: %#v", farrowing)
		}
	})

	t.Run("rejects a non-positive quantity", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New(), CurrentPiglets: 8}

		if err := farrowing.WeanAll(0, actor, now); !errors.Is(err, farrowingdomain.ErrInvalidPigletQuantity) {
			t.Fatalf("error = %v, want ErrInvalidPigletQuantity", err)
		}
	})

	t.Run("rejects a quantity that does not match the balance", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New(), CurrentPiglets: 8}

		if err := farrowing.WeanAll(7, actor, now); !errors.Is(err, farrowingdomain.ErrWeagingQuantityMismatch) {
			t.Fatalf("error = %v, want ErrWeagingQuantityMismatch", err)
		}
		if farrowing.CurrentPiglets != 8 {
			t.Fatalf("current piglets = %d, want 8", farrowing.CurrentPiglets)
		}
	})

	t.Run("rejects a nil updated by", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New(), CurrentPiglets: 8}

		if err := farrowing.WeanAll(8, uuid.Nil, now); !errors.Is(err, farrowingdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})
}

func TestFarrowingMarkAsNurse(t *testing.T) {
	now := time.Date(2026, time.February, 2, 3, 4, 5, 0, time.UTC)
	actor := uuid.New()
	startDate := time.Date(2026, time.February, 1, 18, 30, 0, 0, time.UTC)

	t.Run("flags the farrowing as a nurse", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New()}

		if err := farrowing.MarkAsNurse(startDate, actor, now); err != nil {
			t.Fatalf("MarkAsNurse() error = %v", err)
		}
		if !farrowing.IsNurse || farrowing.NurseStartDate == nil {
			t.Fatalf("nurse fields not set: %#v", farrowing)
		}
		if !farrowing.NurseStartDate.Equal(time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("nurse start date = %v, want truncated day", farrowing.NurseStartDate)
		}
		if !farrowing.UpdatedAt.Equal(now) || farrowing.UpdatedBy != actor {
			t.Fatalf("unexpected audit fields: %#v", farrowing)
		}
	})

	t.Run("rejects a zero start date", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New()}

		if err := farrowing.MarkAsNurse(time.Time{}, actor, now); !errors.Is(err, farrowingdomain.ErrInvalidNurseStartDate) {
			t.Fatalf("error = %v, want ErrInvalidNurseStartDate", err)
		}
	})

	t.Run("rejects a nil updated by", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New()}

		if err := farrowing.MarkAsNurse(startDate, uuid.Nil, now); !errors.Is(err, farrowingdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
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

func newTestFarrowing() (*farrowingdomain.Farrowing, farrowingdomain.UpdateFarrowingReference) {
	lastMount := time.Date(2025, time.October, 10, 0, 0, 0, 0, time.UTC)
	farrowDate := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	return &farrowingdomain.Farrowing{
		ID:             uuid.New(),
		FarrowDate:     farrowDate,
		LiveBorn:       10,
		Stillborn:      1,
		CurrentPiglets: 10,
		Operators:      []*farrowingdomain.FarrowingOperator{},
		Medications:    []*farrowingdomain.FarrowingMedication{},
	}, farrowingdomain.UpdateFarrowingReference{LastMountDate: lastMount}
}

func TestFarrowingUpdate(t *testing.T) {
	now := time.Date(2026, time.February, 2, 3, 4, 5, 0, time.UTC)
	actor := uuid.New()

	t.Run("updates the provided scalar fields", func(t *testing.T) {
		farrowing, reference := newTestFarrowing()
		stillborn := 2
		location := "Corral 5"
		note := "Actualizado"
		manipulated := true

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{
			Stillborn:     &stillborn,
			Location:      &location,
			Note:          &note,
			IsManipulated: &manipulated,
		}, reference, actor, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if farrowing.Stillborn != 2 || farrowing.Location == nil || *farrowing.Location != "Corral 5" {
			t.Fatalf("unexpected fields: %#v", farrowing)
		}
		if farrowing.Note == nil || *farrowing.Note != "Actualizado" || !farrowing.IsManipulated {
			t.Fatalf("unexpected fields: %#v", farrowing)
		}
		if !farrowing.UpdatedAt.Equal(now) || farrowing.UpdatedBy != actor {
			t.Fatalf("unexpected audit fields: %#v", farrowing)
		}
	})

	t.Run("clears nullable fields with an empty string or a clear flag", func(t *testing.T) {
		farrowing, reference := newTestFarrowing()
		weight := 12.5
		farrowing.LitterWeight = &weight
		farrowing.StartTime = ptr("22:00")
		empty := ""

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{
			StartTime:         &empty,
			ClearLitterWeight: true,
		}, reference, actor, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if farrowing.StartTime != nil || farrowing.LitterWeight != nil {
			t.Fatalf("expected nullable fields cleared: %#v", farrowing)
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		farrowing, reference := newTestFarrowing()

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{}, reference, actor, now)

		if !errors.Is(err, farrowingdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects nil id and nil actor", func(t *testing.T) {
		location := "Corral"
		if err := (&farrowingdomain.Farrowing{}).Update(
			farrowingdomain.UpdateFarrowingParams{Location: &location},
			farrowingdomain.UpdateFarrowingReference{}, actor, now,
		); !errors.Is(err, farrowingdomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}

		farrowing, reference := newTestFarrowing()
		if err := farrowing.Update(
			farrowingdomain.UpdateFarrowingParams{Location: &location},
			reference, uuid.Nil, now,
		); !errors.Is(err, farrowingdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})

	t.Run("validates the new farrow date", func(t *testing.T) {
		future := now.AddDate(0, 0, 1)
		beforeMount := time.Date(2025, time.October, 9, 0, 0, 0, 0, time.UTC)

		farrowing, reference := newTestFarrowing()
		if err := farrowing.Update(
			farrowingdomain.UpdateFarrowingParams{FarrowDate: &future}, reference, actor, now,
		); !errors.Is(err, farrowingdomain.ErrFarrowDateInFuture) {
			t.Fatalf("error = %v, want ErrFarrowDateInFuture", err)
		}

		farrowing, reference = newTestFarrowing()
		if err := farrowing.Update(
			farrowingdomain.UpdateFarrowingParams{FarrowDate: &beforeMount}, reference, actor, now,
		); !errors.Is(err, farrowingdomain.ErrFarrowDateBeforeMount) {
			t.Fatalf("error = %v, want ErrFarrowDateBeforeMount", err)
		}
	})

	t.Run("rejects a date after the earliest related record but allows the same day", func(t *testing.T) {
		lateNow := time.Date(2026, time.February, 10, 3, 4, 5, 0, time.UTC)
		earliest := time.Date(2026, time.February, 5, 0, 0, 0, 0, time.UTC)
		after := time.Date(2026, time.February, 6, 0, 0, 0, 0, time.UTC)

		farrowing, reference := newTestFarrowing()
		reference.EarliestRelatedDate = &earliest
		if err := farrowing.Update(
			farrowingdomain.UpdateFarrowingParams{FarrowDate: &after}, reference, actor, lateNow,
		); !errors.Is(err, farrowingdomain.ErrFarrowDateAfterRelatedRecords) {
			t.Fatalf("error = %v, want ErrFarrowDateAfterRelatedRecords", err)
		}

		sameDay := earliest
		farrowing, reference = newTestFarrowing()
		reference.EarliestRelatedDate = &earliest
		if err := farrowing.Update(
			farrowingdomain.UpdateFarrowingParams{FarrowDate: &sameDay}, reference, actor, lateNow,
		); err != nil {
			t.Fatalf("Update() error = %v, want nil for the same day", err)
		}
	})

	t.Run("locks the live born count when related records exist", func(t *testing.T) {
		earliest := time.Date(2026, time.February, 5, 0, 0, 0, 0, time.UTC)
		liveBorn := 12

		farrowing, reference := newTestFarrowing()
		farrowing.CurrentPiglets = 8
		reference.EarliestRelatedDate = &earliest
		if err := farrowing.Update(
			farrowingdomain.UpdateFarrowingParams{LiveBorn: &liveBorn}, reference, actor, now,
		); !errors.Is(err, farrowingdomain.ErrLiveBornLocked) {
			t.Fatalf("error = %v, want ErrLiveBornLocked", err)
		}

		same := farrowing.LiveBorn
		farrowing, reference = newTestFarrowing()
		farrowing.CurrentPiglets = 8
		reference.EarliestRelatedDate = &earliest
		if err := farrowing.Update(
			farrowingdomain.UpdateFarrowingParams{LiveBorn: &same}, reference, actor, now,
		); err != nil {
			t.Fatalf("Update() error = %v, want nil for the unchanged count", err)
		}
		if farrowing.CurrentPiglets != 8 {
			t.Fatalf("current piglets = %d, want 8", farrowing.CurrentPiglets)
		}
	})

	t.Run("recomputes the current piglets balance without related records", func(t *testing.T) {
		farrowing, reference := newTestFarrowing()
		liveBorn := 14

		if err := farrowing.Update(
			farrowingdomain.UpdateFarrowingParams{LiveBorn: &liveBorn}, reference, actor, now,
		); err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if farrowing.LiveBorn != 14 || farrowing.CurrentPiglets != 14 {
			t.Fatalf("unexpected counts: %#v", farrowing)
		}
	})

	t.Run("rejects negative counts and weights", func(t *testing.T) {
		for _, test := range []struct {
			name string
			run  func(*farrowingdomain.Farrowing) error
			want error
		}{
			{"live born", func(f *farrowingdomain.Farrowing) error {
				value := -1
				return f.Update(farrowingdomain.UpdateFarrowingParams{LiveBorn: &value}, farrowingdomain.UpdateFarrowingReference{}, actor, now)
			}, farrowingdomain.ErrInvalidLiveBorn},
			{"stillborn", func(f *farrowingdomain.Farrowing) error {
				value := -1
				return f.Update(farrowingdomain.UpdateFarrowingParams{Stillborn: &value}, farrowingdomain.UpdateFarrowingReference{}, actor, now)
			}, farrowingdomain.ErrInvalidStillborn},
			{"mummified", func(f *farrowingdomain.Farrowing) error {
				value := -1
				return f.Update(farrowingdomain.UpdateFarrowingParams{Mummified: &value}, farrowingdomain.UpdateFarrowingReference{}, actor, now)
			}, farrowingdomain.ErrInvalidMummified},
			{"litter weight", func(f *farrowingdomain.Farrowing) error {
				value := -1.5
				return f.Update(farrowingdomain.UpdateFarrowingParams{LitterWeight: &value}, farrowingdomain.UpdateFarrowingReference{}, actor, now)
			}, farrowingdomain.ErrInvalidLitterWeight},
		} {
			t.Run(test.name, func(t *testing.T) {
				farrowing, _ := newTestFarrowing()
				if err := test.run(farrowing); !errors.Is(err, test.want) {
					t.Fatalf("error = %v, want %v", err, test.want)
				}
			})
		}
	})

	t.Run("rejects an invalid start time", func(t *testing.T) {
		farrowing, reference := newTestFarrowing()
		bad := "25:00"

		if err := farrowing.Update(
			farrowingdomain.UpdateFarrowingParams{StartTime: &bad}, reference, actor, now,
		); !errors.Is(err, farrowingdomain.ErrInvalidStartTime) {
			t.Fatalf("error = %v, want ErrInvalidStartTime", err)
		}
	})

	t.Run("replaces the operator and medication links", func(t *testing.T) {
		farrowing, reference := newTestFarrowing()

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{
			UpdateOperators: true,
			Operators: []farrowingdomain.NewFarrowingOperatorParams{
				{ID: uuid.New(), OperatorID: uuid.New()},
			},
			UpdateMedications: true,
			Medications: []farrowingdomain.NewFarrowingMedicationParams{
				{ID: uuid.New(), MedicationID: uuid.New(), Dose: 1, AppliedBy: uuid.New()},
			},
		}, reference, actor, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if len(farrowing.Operators) != 1 || farrowing.Operators[0].FarrowingID != farrowing.ID {
			t.Fatalf("unexpected operators: %#v", farrowing.Operators)
		}
		if len(farrowing.Medications) != 1 || farrowing.Medications[0].FarrowingID != farrowing.ID {
			t.Fatalf("unexpected medications: %#v", farrowing.Medications)
		}
	})

	t.Run("clears the links with empty lists", func(t *testing.T) {
		farrowing, reference := newTestFarrowing()
		farrowing.Operators = []*farrowingdomain.FarrowingOperator{{ID: uuid.New(), FarrowingID: farrowing.ID, OperatorID: uuid.New()}}

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{
			UpdateOperators: true,
			Operators:       []farrowingdomain.NewFarrowingOperatorParams{},
		}, reference, actor, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if len(farrowing.Operators) != 0 {
			t.Fatalf("expected operators cleared: %#v", farrowing.Operators)
		}
	})

	t.Run("rejects duplicate operators on replacement", func(t *testing.T) {
		farrowing, reference := newTestFarrowing()
		operatorID := uuid.New()

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{
			UpdateOperators: true,
			Operators: []farrowingdomain.NewFarrowingOperatorParams{
				{ID: uuid.New(), OperatorID: operatorID},
				{ID: uuid.New(), OperatorID: operatorID},
			},
		}, reference, actor, now)

		if !errors.Is(err, farrowingdomain.ErrDuplicateOperator) {
			t.Fatalf("error = %v, want ErrDuplicateOperator", err)
		}
	})
}

func ptr(value string) *string {
	return &value
}
