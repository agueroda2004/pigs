package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	partialweagingdomain "server/internal/modules/partialweaging/domain"
)

var (
	farrowDate    = time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC)
	lastEventDate = time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC)
	nowReference  = time.Date(2026, time.April, 30, 12, 0, 0, 0, time.UTC)
)

// validBounds returns the farrowing references of a farrowing with three current piglets.
func validBounds() partialweagingdomain.PartialWeagingBounds {
	return partialweagingdomain.PartialWeagingBounds{
		FarrowDate:     farrowDate,
		LastEventDate:  lastEventDate,
		CurrentPiglets: 3,
	}
}

func validParams() partialweagingdomain.NewPartialWeagingParams {
	weight := 42.5
	note := "Destete parcial de la camada"
	return partialweagingdomain.NewPartialWeagingParams{
		ID:          uuid.New(),
		FarrowingID: uuid.New(),
		SowID:       uuid.New(),
		SowCode:     "C-001",
		WeagingDate: time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC),
		Quantity:    3,
		TotalWeight: &weight,
		Type:        partialweagingdomain.TypeNormal,
		Note:        &note,
		CreatedBy:   uuid.New(),
	}
}

func TestNewPartialWeaging(t *testing.T) {
	t.Run("builds a valid partial weaging", func(t *testing.T) {
		params := validParams()

		weaging, err := partialweagingdomain.NewPartialWeaging(params, validBounds(), nowReference)

		if err != nil {
			t.Fatalf("NewPartialWeaging() error = %v", err)
		}
		if weaging.FarrowingID != params.FarrowingID || weaging.SowID != params.SowID {
			t.Fatalf("unexpected references: %#v", weaging)
		}
		if weaging.SowCode != params.SowCode {
			t.Fatalf("sow code = %q, want %q", weaging.SowCode, params.SowCode)
		}
		if weaging.Quantity != 3 || weaging.TotalWeight == nil || *weaging.TotalWeight != 42.5 {
			t.Fatalf("unexpected fields: %#v", weaging)
		}
		if weaging.Type != partialweagingdomain.TypeNormal || weaging.Note == nil || *weaging.Note != "Destete parcial de la camada" {
			t.Fatalf("unexpected type or note: %#v", weaging)
		}
		if !weaging.CreatedAt.Equal(nowReference) || weaging.CreatedBy != weaging.UpdatedBy {
			t.Fatalf("unexpected audit fields: %#v", weaging)
		}
	})

	t.Run("clears an empty note", func(t *testing.T) {
		params := validParams()
		empty := "   "
		params.Note = &empty

		weaging, err := partialweagingdomain.NewPartialWeaging(params, validBounds(), nowReference)

		if err != nil {
			t.Fatalf("NewPartialWeaging() error = %v", err)
		}
		if weaging.Note != nil {
			t.Fatalf("note = %v, want nil", weaging.Note)
		}
	})

	t.Run("rejects invalid required fields", func(t *testing.T) {
		for _, test := range []struct {
			name string
			set  func(*partialweagingdomain.NewPartialWeagingParams)
			want error
		}{
			{"id", func(p *partialweagingdomain.NewPartialWeagingParams) { p.ID = uuid.Nil }, partialweagingdomain.ErrInvalidID},
			{"farrowing", func(p *partialweagingdomain.NewPartialWeagingParams) { p.FarrowingID = uuid.Nil }, partialweagingdomain.ErrInvalidFarrowing},
			{"sow", func(p *partialweagingdomain.NewPartialWeagingParams) { p.SowID = uuid.Nil }, partialweagingdomain.ErrInvalidSow},
			{"weaging date", func(p *partialweagingdomain.NewPartialWeagingParams) { p.WeagingDate = time.Time{} }, partialweagingdomain.ErrInvalidWeagingDate},
			{"created by", func(p *partialweagingdomain.NewPartialWeagingParams) { p.CreatedBy = uuid.Nil }, partialweagingdomain.ErrInvalidCreatedBy},
		} {
			t.Run(test.name, func(t *testing.T) {
				params := validParams()
				test.set(&params)

				_, err := partialweagingdomain.NewPartialWeaging(params, validBounds(), nowReference)

				if !errors.Is(err, test.want) {
					t.Fatalf("error = %v, want %v", err, test.want)
				}
			})
		}
	})

	t.Run("rejects a nil farrow date", func(t *testing.T) {
		params := validParams()
		bounds := validBounds()
		bounds.FarrowDate = time.Time{}

		_, err := partialweagingdomain.NewPartialWeaging(params, bounds, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrInvalidFarrowDate) {
			t.Fatalf("error = %v, want ErrInvalidFarrowDate", err)
		}
	})

	t.Run("rejects a non-positive quantity", func(t *testing.T) {
		params := validParams()
		params.Quantity = 0

		_, err := partialweagingdomain.NewPartialWeaging(params, validBounds(), nowReference)

		if !errors.Is(err, partialweagingdomain.ErrInvalidQuantity) {
			t.Fatalf("error = %v, want ErrInvalidQuantity", err)
		}
	})

	t.Run("rejects a quantity above 30", func(t *testing.T) {
		params := validParams()
		params.Quantity = 31
		bounds := validBounds()
		bounds.CurrentPiglets = 31

		_, err := partialweagingdomain.NewPartialWeaging(params, bounds, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrQuantityExceedsMax) {
			t.Fatalf("error = %v, want ErrQuantityExceedsMax", err)
		}
	})

	t.Run("rejects a quantity that differs from the current piglets", func(t *testing.T) {
		params := validParams()
		params.Quantity = 2

		_, err := partialweagingdomain.NewPartialWeaging(params, validBounds(), nowReference)

		if !errors.Is(err, partialweagingdomain.ErrQuantityMismatch) {
			t.Fatalf("error = %v, want ErrQuantityMismatch", err)
		}
	})

	t.Run("rejects a non-positive weight", func(t *testing.T) {
		params := validParams()
		weight := 0.0
		params.TotalWeight = &weight

		_, err := partialweagingdomain.NewPartialWeaging(params, validBounds(), nowReference)

		if !errors.Is(err, partialweagingdomain.ErrInvalidTotalWeight) {
			t.Fatalf("error = %v, want ErrInvalidTotalWeight", err)
		}
	})

	t.Run("rejects a weight above 200", func(t *testing.T) {
		params := validParams()
		weight := 200.5
		params.TotalWeight = &weight

		_, err := partialweagingdomain.NewPartialWeaging(params, validBounds(), nowReference)

		if !errors.Is(err, partialweagingdomain.ErrTotalWeightExceedsMax) {
			t.Fatalf("error = %v, want ErrTotalWeightExceedsMax", err)
		}
	})

	t.Run("accepts a weight of 200", func(t *testing.T) {
		params := validParams()
		weight := 200.0
		params.TotalWeight = &weight

		if _, err := partialweagingdomain.NewPartialWeaging(params, validBounds(), nowReference); err != nil {
			t.Fatalf("NewPartialWeaging() error = %v", err)
		}
	})

	t.Run("rejects an invalid type", func(t *testing.T) {
		params := validParams()
		params.Type = "Desconocido"

		_, err := partialweagingdomain.NewPartialWeaging(params, validBounds(), nowReference)

		if !errors.Is(err, partialweagingdomain.ErrInvalidType) {
			t.Fatalf("error = %v, want ErrInvalidType", err)
		}
	})

	t.Run("rejects a note that is too long", func(t *testing.T) {
		params := validParams()
		long := string(make([]rune, 501))
		params.Note = &long

		_, err := partialweagingdomain.NewPartialWeaging(params, validBounds(), nowReference)

		if !errors.Is(err, partialweagingdomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
	})

	t.Run("rejects a weaging date on or before the farrow date", func(t *testing.T) {
		params := validParams()
		params.WeagingDate = farrowDate

		_, err := partialweagingdomain.NewPartialWeaging(params, validBounds(), nowReference)

		if !errors.Is(err, partialweagingdomain.ErrWeagingDateBeforeFarrowing) {
			t.Fatalf("error = %v, want ErrWeagingDateBeforeFarrowing", err)
		}
	})

	t.Run("accepts a weaging date on the last event day", func(t *testing.T) {
		params := validParams()
		params.WeagingDate = lastEventDate

		_, err := partialweagingdomain.NewPartialWeaging(params, validBounds(), nowReference)

		if err != nil {
			t.Fatalf("NewPartialWeaging() error = %v", err)
		}
	})

	t.Run("rejects a weaging date before the last event", func(t *testing.T) {
		params := validParams()
		bounds := validBounds()
		bounds.LastEventDate = params.WeagingDate.AddDate(0, 0, 1)

		_, err := partialweagingdomain.NewPartialWeaging(params, bounds, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrWeagingDateBeforeEvents) {
			t.Fatalf("error = %v, want ErrWeagingDateBeforeEvents", err)
		}
	})

	t.Run("skips the event check when there is no event", func(t *testing.T) {
		params := validParams()
		bounds := validBounds()
		bounds.LastEventDate = time.Time{}

		if _, err := partialweagingdomain.NewPartialWeaging(params, bounds, nowReference); err != nil {
			t.Fatalf("NewPartialWeaging() error = %v", err)
		}
	})

	t.Run("accepts a weaging date after the last partial weaging", func(t *testing.T) {
		params := validParams()
		bounds := validBounds()
		bounds.LastPartialWeagingDate = time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC)

		if _, err := partialweagingdomain.NewPartialWeaging(params, bounds, nowReference); err != nil {
			t.Fatalf("NewPartialWeaging() error = %v", err)
		}
	})

	t.Run("accepts a weaging date on the last partial weaging day", func(t *testing.T) {
		params := validParams()
		bounds := validBounds()
		bounds.LastPartialWeagingDate = params.WeagingDate

		if _, err := partialweagingdomain.NewPartialWeaging(params, bounds, nowReference); err != nil {
			t.Fatalf("NewPartialWeaging() error = %v", err)
		}
	})

	t.Run("rejects a weaging date before the last partial weaging", func(t *testing.T) {
		params := validParams()
		bounds := validBounds()
		bounds.LastPartialWeagingDate = time.Date(2026, time.April, 26, 0, 0, 0, 0, time.UTC)

		_, err := partialweagingdomain.NewPartialWeaging(params, bounds, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrWeagingDateBeforePartial) {
			t.Fatalf("error = %v, want ErrWeagingDateBeforePartial", err)
		}
	})

	t.Run("rejects a weaging date in the future", func(t *testing.T) {
		params := validParams()
		params.WeagingDate = nowReference.AddDate(0, 0, 1)

		_, err := partialweagingdomain.NewPartialWeaging(params, validBounds(), nowReference)

		if !errors.Is(err, partialweagingdomain.ErrWeagingDateInFuture) {
			t.Fatalf("error = %v, want ErrWeagingDateInFuture", err)
		}
	})
}

func updateWeaging(weagingType partialweagingdomain.Type) *partialweagingdomain.PartialWeaging {
	weight := 40.0
	note := "Destete parcial"
	actor := uuid.New()
	return &partialweagingdomain.PartialWeaging{
		ID:          uuid.New(),
		FarrowingID: uuid.New(),
		SowID:       uuid.New(),
		WeagingDate: time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC),
		Quantity:    3,
		TotalWeight: &weight,
		Type:        weagingType,
		Note:        &note,
		CreatedAt:   nowReference,
		UpdatedAt:   nowReference,
		CreatedBy:   actor,
		UpdatedBy:   actor,
	}
}

func TestUpdatePartialWeaging(t *testing.T) {
	updatedBy := uuid.New()

	t.Run("updates the editable fields", func(t *testing.T) {
		weaging := updateWeaging(partialweagingdomain.TypeNormal)
		weagingDate := weaging.WeagingDate
		weight := 55.5
		note := "  Nueva nota  "

		err := weaging.Update(partialweagingdomain.UpdatePartialWeagingParams{
			TotalWeight: &weight,
			Note:        &note,
		}, updatedBy, nowReference)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if weaging.TotalWeight == nil || *weaging.TotalWeight != 55.5 {
			t.Fatalf("unexpected fields: %#v", weaging)
		}
		if !weaging.WeagingDate.Equal(weagingDate) {
			t.Fatalf("date should not change: %#v", weaging.WeagingDate)
		}
		if weaging.Note == nil || *weaging.Note != "Nueva nota" {
			t.Fatalf("unexpected note: %#v", weaging.Note)
		}
		if weaging.UpdatedBy != updatedBy || !weaging.UpdatedAt.Equal(nowReference) {
			t.Fatalf("unexpected audit fields: %#v", weaging)
		}
		if weaging.Quantity != 3 || weaging.Type != partialweagingdomain.TypeNormal {
			t.Fatalf("immutable fields changed: %#v", weaging)
		}
	})

	t.Run("clears the weight", func(t *testing.T) {
		weaging := updateWeaging(partialweagingdomain.TypeNormal)

		err := weaging.Update(partialweagingdomain.UpdatePartialWeagingParams{
			ClearTotalWeight: true,
		}, updatedBy, nowReference)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if weaging.TotalWeight != nil {
			t.Fatalf("weight = %v, want nil", weaging.TotalWeight)
		}
	})

	t.Run("clears the note with an empty value", func(t *testing.T) {
		weaging := updateWeaging(partialweagingdomain.TypeNormal)
		empty := "   "

		err := weaging.Update(partialweagingdomain.UpdatePartialWeagingParams{
			Note: &empty,
		}, updatedBy, nowReference)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if weaging.Note != nil {
			t.Fatalf("note = %v, want nil", weaging.Note)
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		weaging := updateWeaging(partialweagingdomain.TypeNormal)

		err := weaging.Update(partialweagingdomain.UpdatePartialWeagingParams{}, updatedBy, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects setting and clearing the weight together", func(t *testing.T) {
		weaging := updateWeaging(partialweagingdomain.TypeNormal)
		weight := 10.0

		err := weaging.Update(partialweagingdomain.UpdatePartialWeagingParams{
			TotalWeight:      &weight,
			ClearTotalWeight: true,
		}, updatedBy, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects a missing actor", func(t *testing.T) {
		weaging := updateWeaging(partialweagingdomain.TypeNormal)

		err := weaging.Update(partialweagingdomain.UpdatePartialWeagingParams{
			ClearTotalWeight: true,
		}, uuid.Nil, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})

	t.Run("rejects a nil partial weaging", func(t *testing.T) {
		var weaging *partialweagingdomain.PartialWeaging

		err := weaging.Update(partialweagingdomain.UpdatePartialWeagingParams{
			ClearTotalWeight: true,
		}, updatedBy, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("validates the weight", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			weight float64
			want   error
		}{
			{"zero", 0, partialweagingdomain.ErrInvalidTotalWeight},
			{"negative", -1, partialweagingdomain.ErrInvalidTotalWeight},
			{"above maximum", 200.5, partialweagingdomain.ErrTotalWeightExceedsMax},
		} {
			t.Run(test.name, func(t *testing.T) {
				weaging := updateWeaging(partialweagingdomain.TypeNormal)
				weight := test.weight

				err := weaging.Update(partialweagingdomain.UpdatePartialWeagingParams{
					TotalWeight: &weight,
				}, updatedBy, nowReference)

				if !errors.Is(err, test.want) {
					t.Fatalf("error = %v, want %v", err, test.want)
				}
			})
		}
	})

	t.Run("accepts a weight of 200", func(t *testing.T) {
		weaging := updateWeaging(partialweagingdomain.TypeNormal)
		weight := 200.0

		err := weaging.Update(partialweagingdomain.UpdatePartialWeagingParams{
			TotalWeight: &weight,
		}, updatedBy, nowReference)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
	})

	t.Run("rejects a note that is too long", func(t *testing.T) {
		weaging := updateWeaging(partialweagingdomain.TypeNormal)
		long := string(make([]rune, 501))

		err := weaging.Update(partialweagingdomain.UpdatePartialWeagingParams{
			Note: &long,
		}, updatedBy, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
	})
}

func TestParseType(t *testing.T) {
	t.Run("parses the known types", func(t *testing.T) {
		for _, value := range []partialweagingdomain.Type{
			partialweagingdomain.TypeNormal,
			partialweagingdomain.TypeNodriza,
			partialweagingdomain.TypeBajaViabilidad,
		} {
			parsed, err := partialweagingdomain.ParseType(string(value))
			if err != nil || parsed != value {
				t.Fatalf("ParseType(%q) = %q, %v", value, parsed, err)
			}
		}
	})

	t.Run("rejects an unknown type", func(t *testing.T) {
		if _, err := partialweagingdomain.ParseType("Nope"); !errors.Is(err, partialweagingdomain.ErrInvalidType) {
			t.Fatalf("error = %v, want ErrInvalidType", err)
		}
	})
}
