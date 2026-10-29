package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
)

func validParams() farrowingdomain.NewFarrowingParams {
	litterWeight := 15.5
	stillbornWeight := 2.5
	return farrowingdomain.NewFarrowingParams{
		ID:              uuid.New(),
		ServiceID:       uuid.New(),
		SowID:           uuid.New(),
		FarrowDate:      time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC),
		LitterWeight:    &litterWeight,
		StillbornWeight: &stillbornWeight,
		CreatedBy:       uuid.New(),
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

	t.Run("requires both weights", func(t *testing.T) {
		for _, test := range []struct {
			name string
			set  func(*farrowingdomain.NewFarrowingParams)
			want error
		}{
			{"litter weight", func(p *farrowingdomain.NewFarrowingParams) { p.LitterWeight = nil }, farrowingdomain.ErrLitterWeightRequired},
			{"stillborn weight", func(p *farrowingdomain.NewFarrowingParams) { p.StillbornWeight = nil }, farrowingdomain.ErrStillbornWeightRequired},
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

	t.Run("accepts weights up to 30", func(t *testing.T) {
		params := validParams()
		limit := 30.0
		params.LitterWeight = &limit
		params.StillbornWeight = &limit

		farrowing, err := farrowingdomain.NewFarrowing(params, nil, nil, lastMount, now)

		if err != nil {
			t.Fatalf("NewFarrowing() error = %v", err)
		}
		if farrowing.LitterWeight == nil || *farrowing.LitterWeight != 30 {
			t.Fatalf("unexpected litter weight: %#v", farrowing.LitterWeight)
		}
		if farrowing.StillbornWeight == nil || *farrowing.StillbornWeight != 30 {
			t.Fatalf("unexpected stillborn weight: %#v", farrowing.StillbornWeight)
		}
	})

	t.Run("rejects weights greater than 30", func(t *testing.T) {
		for _, test := range []struct {
			name string
			set  func(*farrowingdomain.NewFarrowingParams)
			want error
		}{
			{"litter weight", func(p *farrowingdomain.NewFarrowingParams) {
				value := 30.5
				p.LitterWeight = &value
			}, farrowingdomain.ErrLitterWeightTooHigh},
			{"stillborn weight", func(p *farrowingdomain.NewFarrowingParams) {
				value := 30.5
				p.StillbornWeight = &value
			}, farrowingdomain.ErrStillbornWeightTooHigh},
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

	t.Run("accepts a medication dose of 30", func(t *testing.T) {
		medication := farrowingdomain.NewFarrowingMedicationParams{
			ID: uuid.New(), MedicationID: uuid.New(), Dose: 30, AppliedBy: uuid.New(),
		}

		farrowing, err := farrowingdomain.NewFarrowing(validParams(), nil, []farrowingdomain.NewFarrowingMedicationParams{medication}, lastMount, now)

		if err != nil {
			t.Fatalf("NewFarrowing() error = %v", err)
		}
		if len(farrowing.Medications) != 1 || farrowing.Medications[0].Dose != 30 {
			t.Fatalf("unexpected medications: %#v", farrowing.Medications)
		}
	})

	t.Run("rejects a medication dose greater than 30", func(t *testing.T) {
		medication := farrowingdomain.NewFarrowingMedicationParams{
			ID: uuid.New(), MedicationID: uuid.New(), Dose: 30.5, AppliedBy: uuid.New(),
		}

		_, err := farrowingdomain.NewFarrowing(validParams(), nil, []farrowingdomain.NewFarrowingMedicationParams{medication}, lastMount, now)

		if !errors.Is(err, farrowingdomain.ErrMedicationDoseTooHigh) {
			t.Fatalf("error = %v, want ErrMedicationDoseTooHigh", err)
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

func TestFarrowingRestoreNurseState(t *testing.T) {
	now := time.Date(2026, time.February, 2, 3, 4, 5, 0, time.UTC)
	actor := uuid.New()
	startDate := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)

	t.Run("clears the nurse state with a zero date", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{
			ID:             uuid.New(),
			IsNurse:        true,
			NurseStartDate: &startDate,
		}

		if err := farrowing.RestoreNurseState(time.Time{}, actor, now); err != nil {
			t.Fatalf("RestoreNurseState() error = %v", err)
		}
		if farrowing.IsNurse || farrowing.NurseStartDate != nil {
			t.Fatalf("nurse state was not cleared: %#v", farrowing)
		}
		if !farrowing.UpdatedAt.Equal(now) || farrowing.UpdatedBy != actor {
			t.Fatalf("unexpected audit fields: %#v", farrowing)
		}
	})

	t.Run("keeps the nurse state from an earlier date", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New()}
		earlier := time.Date(2026, time.January, 28, 16, 15, 0, 0, time.UTC)

		if err := farrowing.RestoreNurseState(earlier, actor, now); err != nil {
			t.Fatalf("RestoreNurseState() error = %v", err)
		}
		if !farrowing.IsNurse || farrowing.NurseStartDate == nil {
			t.Fatalf("nurse fields not set: %#v", farrowing)
		}
		if !farrowing.NurseStartDate.Equal(time.Date(2026, time.January, 28, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("nurse start date = %v, want truncated day", farrowing.NurseStartDate)
		}
	})

	t.Run("rejects a nil actor", func(t *testing.T) {
		farrowing := &farrowingdomain.Farrowing{ID: uuid.New()}

		if err := farrowing.RestoreNurseState(time.Time{}, uuid.Nil, now); !errors.Is(err, farrowingdomain.ErrInvalidUpdatedBy) {
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

func newTestFarrowing() *farrowingdomain.Farrowing {
	farrowDate := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	litterWeight := 15.5
	stillbornWeight := 1.5
	return &farrowingdomain.Farrowing{
		ID:              uuid.New(),
		FarrowDate:      farrowDate,
		LiveBorn:        10,
		Stillborn:       1,
		CurrentPiglets:  10,
		LitterWeight:    &litterWeight,
		StillbornWeight: &stillbornWeight,
		Operators:       []*farrowingdomain.FarrowingOperator{},
		Medications:     []*farrowingdomain.FarrowingMedication{},
	}
}

func TestFarrowingUpdate(t *testing.T) {
	now := time.Date(2026, time.February, 2, 3, 4, 5, 0, time.UTC)
	actor := uuid.New()

	t.Run("updates the provided scalar fields", func(t *testing.T) {
		farrowing := newTestFarrowing()
		location := "Corral 5"
		note := "Actualizado"
		manipulated := true

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{
			Location:      &location,
			Note:          &note,
			IsManipulated: &manipulated,
		}, actor, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if farrowing.Location == nil || *farrowing.Location != "Corral 5" {
			t.Fatalf("unexpected fields: %#v", farrowing)
		}
		if farrowing.Note == nil || *farrowing.Note != "Actualizado" || !farrowing.IsManipulated {
			t.Fatalf("unexpected fields: %#v", farrowing)
		}
		if !farrowing.UpdatedAt.Equal(now) || farrowing.UpdatedBy != actor {
			t.Fatalf("unexpected audit fields: %#v", farrowing)
		}
	})

	t.Run("clears nullable string fields with an empty string", func(t *testing.T) {
		farrowing := newTestFarrowing()
		farrowing.StartTime = ptr("22:00")
		empty := ""

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{
			StartTime: &empty,
		}, actor, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if farrowing.StartTime != nil {
			t.Fatalf("expected the start time cleared: %#v", farrowing)
		}
	})

	t.Run("keeps the weights when they are omitted", func(t *testing.T) {
		farrowing := newTestFarrowing()
		location := "Corral 5"

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{
			Location: &location,
		}, actor, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if farrowing.LitterWeight == nil || *farrowing.LitterWeight != 15.5 {
			t.Fatalf("unexpected litter weight: %#v", farrowing.LitterWeight)
		}
		if farrowing.StillbornWeight == nil || *farrowing.StillbornWeight != 1.5 {
			t.Fatalf("unexpected stillborn weight: %#v", farrowing.StillbornWeight)
		}
	})

	t.Run("updates the weights", func(t *testing.T) {
		farrowing := newTestFarrowing()
		litter := 30.0
		stillborn := 0.0

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{
			LitterWeight:    &litter,
			StillbornWeight: &stillborn,
		}, actor, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if farrowing.LitterWeight == nil || *farrowing.LitterWeight != 30 {
			t.Fatalf("unexpected litter weight: %#v", farrowing.LitterWeight)
		}
		if farrowing.StillbornWeight == nil || *farrowing.StillbornWeight != 0 {
			t.Fatalf("unexpected stillborn weight: %#v", farrowing.StillbornWeight)
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		farrowing := newTestFarrowing()

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{}, actor, now)

		if !errors.Is(err, farrowingdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects nil id and nil actor", func(t *testing.T) {
		location := "Corral"
		if err := (&farrowingdomain.Farrowing{}).Update(
			farrowingdomain.UpdateFarrowingParams{Location: &location}, actor, now,
		); !errors.Is(err, farrowingdomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}

		farrowing := newTestFarrowing()
		if err := farrowing.Update(
			farrowingdomain.UpdateFarrowingParams{Location: &location}, uuid.Nil, now,
		); !errors.Is(err, farrowingdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})

	t.Run("rejects changes to the immovable fields", func(t *testing.T) {
		for _, test := range []struct {
			name string
			run  func(*farrowingdomain.Farrowing) error
		}{
			{"farrow date", func(f *farrowingdomain.Farrowing) error {
				value := f.FarrowDate.AddDate(0, 0, 1)
				return f.Update(farrowingdomain.UpdateFarrowingParams{FarrowDate: &value}, actor, now)
			}},
			{"live born", func(f *farrowingdomain.Farrowing) error {
				value := 11
				return f.Update(farrowingdomain.UpdateFarrowingParams{LiveBorn: &value}, actor, now)
			}},
			{"stillborn", func(f *farrowingdomain.Farrowing) error {
				value := 2
				return f.Update(farrowingdomain.UpdateFarrowingParams{Stillborn: &value}, actor, now)
			}},
			{"mummified", func(f *farrowingdomain.Farrowing) error {
				value := 1
				return f.Update(farrowingdomain.UpdateFarrowingParams{Mummified: &value}, actor, now)
			}},
		} {
			t.Run(test.name, func(t *testing.T) {
				farrowing := newTestFarrowing()

				err := test.run(farrowing)

				if !errors.Is(err, farrowingdomain.ErrImmutableFarrowField) {
					t.Fatalf("error = %v, want ErrImmutableFarrowField", err)
				}
			})
		}
	})

	t.Run("accepts the current date and counts as a no-op", func(t *testing.T) {
		farrowing := newTestFarrowing()
		date := farrowing.FarrowDate
		liveBorn := 10
		stillborn := 1
		mummified := 0

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{
			FarrowDate: &date,
			LiveBorn:   &liveBorn,
			Stillborn:  &stillborn,
			Mummified:  &mummified,
		}, actor, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if farrowing.CurrentPiglets != 10 {
			t.Fatalf("current piglets = %d, want 10", farrowing.CurrentPiglets)
		}
	})

	t.Run("rejects invalid weights", func(t *testing.T) {
		for _, test := range []struct {
			name string
			run  func(*farrowingdomain.Farrowing) error
			want error
		}{
			{"negative litter weight", func(f *farrowingdomain.Farrowing) error {
				value := -1.5
				return f.Update(farrowingdomain.UpdateFarrowingParams{LitterWeight: &value}, actor, now)
			}, farrowingdomain.ErrInvalidLitterWeight},
			{"negative stillborn weight", func(f *farrowingdomain.Farrowing) error {
				value := -1.5
				return f.Update(farrowingdomain.UpdateFarrowingParams{StillbornWeight: &value}, actor, now)
			}, farrowingdomain.ErrInvalidStillbornWeight},
			{"litter weight over 30", func(f *farrowingdomain.Farrowing) error {
				value := 30.5
				return f.Update(farrowingdomain.UpdateFarrowingParams{LitterWeight: &value}, actor, now)
			}, farrowingdomain.ErrLitterWeightTooHigh},
			{"stillborn weight over 30", func(f *farrowingdomain.Farrowing) error {
				value := 30.5
				return f.Update(farrowingdomain.UpdateFarrowingParams{StillbornWeight: &value}, actor, now)
			}, farrowingdomain.ErrStillbornWeightTooHigh},
		} {
			t.Run(test.name, func(t *testing.T) {
				farrowing := newTestFarrowing()
				if err := test.run(farrowing); !errors.Is(err, test.want) {
					t.Fatalf("error = %v, want %v", err, test.want)
				}
			})
		}
	})

	t.Run("rejects a medication dose greater than 30", func(t *testing.T) {
		farrowing := newTestFarrowing()

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{
			UpdateMedications: true,
			Medications: []farrowingdomain.NewFarrowingMedicationParams{
				{ID: uuid.New(), MedicationID: uuid.New(), Dose: 31, AppliedBy: uuid.New()},
			},
		}, actor, now)

		if !errors.Is(err, farrowingdomain.ErrMedicationDoseTooHigh) {
			t.Fatalf("error = %v, want ErrMedicationDoseTooHigh", err)
		}
	})

	t.Run("rejects an invalid start time", func(t *testing.T) {
		farrowing := newTestFarrowing()
		bad := "25:00"

		if err := farrowing.Update(
			farrowingdomain.UpdateFarrowingParams{StartTime: &bad}, actor, now,
		); !errors.Is(err, farrowingdomain.ErrInvalidStartTime) {
			t.Fatalf("error = %v, want ErrInvalidStartTime", err)
		}
	})

	t.Run("replaces the operator and medication links", func(t *testing.T) {
		farrowing := newTestFarrowing()

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{
			UpdateOperators: true,
			Operators: []farrowingdomain.NewFarrowingOperatorParams{
				{ID: uuid.New(), OperatorID: uuid.New()},
			},
			UpdateMedications: true,
			Medications: []farrowingdomain.NewFarrowingMedicationParams{
				{ID: uuid.New(), MedicationID: uuid.New(), Dose: 1, AppliedBy: uuid.New()},
			},
		}, actor, now)

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
		farrowing := newTestFarrowing()
		farrowing.Operators = []*farrowingdomain.FarrowingOperator{{ID: uuid.New(), FarrowingID: farrowing.ID, OperatorID: uuid.New()}}

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{
			UpdateOperators: true,
			Operators:       []farrowingdomain.NewFarrowingOperatorParams{},
		}, actor, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if len(farrowing.Operators) != 0 {
			t.Fatalf("expected operators cleared: %#v", farrowing.Operators)
		}
	})

	t.Run("rejects duplicate operators on replacement", func(t *testing.T) {
		farrowing := newTestFarrowing()
		operatorID := uuid.New()

		err := farrowing.Update(farrowingdomain.UpdateFarrowingParams{
			UpdateOperators: true,
			Operators: []farrowingdomain.NewFarrowingOperatorParams{
				{ID: uuid.New(), OperatorID: operatorID},
				{ID: uuid.New(), OperatorID: operatorID},
			},
		}, actor, now)

		if !errors.Is(err, farrowingdomain.ErrDuplicateOperator) {
			t.Fatalf("error = %v, want ErrDuplicateOperator", err)
		}
	})
}

func ptr(value string) *string {
	return &value
}
