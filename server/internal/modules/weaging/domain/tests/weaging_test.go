package tests

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	weagingdomain "server/internal/modules/weaging/domain"
)

var (
	farrowDate    = time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC)
	lastEventDate = time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC)
	nowReference  = time.Date(2026, time.April, 30, 12, 0, 0, 0, time.UTC)
)

func validParams() weagingdomain.NewWeagingParams {
	weight := 120.5
	destination := "Nave 2"
	note := "Destete completo de la camada"
	return weagingdomain.NewWeagingParams{
		ID:          uuid.New(),
		FarrowingID: uuid.New(),
		SowID:       uuid.New(),
		WeagingDate: time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC),
		Quantity:    8,
		TotalWeight: &weight,
		Destination: &destination,
		Note:        &note,
		CreatedBy:   uuid.New(),
	}
}

func TestNewWeaging(t *testing.T) {
	t.Run("builds a valid weaging", func(t *testing.T) {
		params := validParams()

		weaging, err := weagingdomain.NewWeaging(params, farrowDate, lastEventDate, nowReference)

		if err != nil {
			t.Fatalf("NewWeaging() error = %v", err)
		}
		if weaging.FarrowingID != params.FarrowingID || weaging.SowID != params.SowID {
			t.Fatalf("unexpected references: %#v", weaging)
		}
		if weaging.Quantity != 8 || weaging.TotalWeight == nil || *weaging.TotalWeight != 120.5 {
			t.Fatalf("unexpected fields: %#v", weaging)
		}
		if weaging.Destination == nil || *weaging.Destination != "Nave 2" {
			t.Fatalf("unexpected destination: %#v", weaging)
		}
		if weaging.Note == nil || *weaging.Note != "Destete completo de la camada" {
			t.Fatalf("unexpected note: %#v", weaging)
		}
		if !weaging.CreatedAt.Equal(nowReference) || weaging.CreatedBy != weaging.UpdatedBy {
			t.Fatalf("unexpected audit fields: %#v", weaging)
		}
	})

	t.Run("clears empty optional strings", func(t *testing.T) {
		params := validParams()
		empty := "   "
		params.Destination = &empty
		params.Note = &empty

		weaging, err := weagingdomain.NewWeaging(params, farrowDate, lastEventDate, nowReference)

		if err != nil {
			t.Fatalf("NewWeaging() error = %v", err)
		}
		if weaging.Destination != nil || weaging.Note != nil {
			t.Fatalf("expected nil destination and note: %#v", weaging)
		}
	})

	t.Run("rejects invalid required fields", func(t *testing.T) {
		for _, test := range []struct {
			name string
			set  func(*weagingdomain.NewWeagingParams)
			want error
		}{
			{"id", func(p *weagingdomain.NewWeagingParams) { p.ID = uuid.Nil }, weagingdomain.ErrInvalidID},
			{"farrowing", func(p *weagingdomain.NewWeagingParams) { p.FarrowingID = uuid.Nil }, weagingdomain.ErrInvalidFarrowing},
			{"sow", func(p *weagingdomain.NewWeagingParams) { p.SowID = uuid.Nil }, weagingdomain.ErrInvalidSow},
			{"weaging date", func(p *weagingdomain.NewWeagingParams) { p.WeagingDate = time.Time{} }, weagingdomain.ErrInvalidWeagingDate},
			{"created by", func(p *weagingdomain.NewWeagingParams) { p.CreatedBy = uuid.Nil }, weagingdomain.ErrInvalidCreatedBy},
		} {
			t.Run(test.name, func(t *testing.T) {
				params := validParams()
				test.set(&params)

				_, err := weagingdomain.NewWeaging(params, farrowDate, lastEventDate, nowReference)

				if !errors.Is(err, test.want) {
					t.Fatalf("error = %v, want %v", err, test.want)
				}
			})
		}
	})

	t.Run("rejects a nil farrow date", func(t *testing.T) {
		params := validParams()

		_, err := weagingdomain.NewWeaging(params, time.Time{}, lastEventDate, nowReference)

		if !errors.Is(err, weagingdomain.ErrInvalidFarrowDate) {
			t.Fatalf("error = %v, want ErrInvalidFarrowDate", err)
		}
	})

	t.Run("rejects a non-positive quantity", func(t *testing.T) {
		params := validParams()
		params.Quantity = 0

		_, err := weagingdomain.NewWeaging(params, farrowDate, lastEventDate, nowReference)

		if !errors.Is(err, weagingdomain.ErrInvalidQuantity) {
			t.Fatalf("error = %v, want ErrInvalidQuantity", err)
		}
	})

	t.Run("rejects a non-positive weight", func(t *testing.T) {
		params := validParams()
		weight := 0.0
		params.TotalWeight = &weight

		_, err := weagingdomain.NewWeaging(params, farrowDate, lastEventDate, nowReference)

		if !errors.Is(err, weagingdomain.ErrInvalidTotalWeight) {
			t.Fatalf("error = %v, want ErrInvalidTotalWeight", err)
		}
	})

	t.Run("accepts the maximum quantity and weight", func(t *testing.T) {
		params := validParams()
		params.Quantity = 30
		weight := 200.0
		params.TotalWeight = &weight

		weaging, err := weagingdomain.NewWeaging(params, farrowDate, lastEventDate, nowReference)

		if err != nil {
			t.Fatalf("NewWeaging() error = %v", err)
		}
		if weaging.Quantity != 30 {
			t.Fatalf("quantity = %d, want 30", weaging.Quantity)
		}
		if weaging.TotalWeight == nil || *weaging.TotalWeight != 200 {
			t.Fatalf("total weight = %v, want 200", weaging.TotalWeight)
		}
	})

	t.Run("rejects a quantity above the maximum", func(t *testing.T) {
		params := validParams()
		params.Quantity = 31

		_, err := weagingdomain.NewWeaging(params, farrowDate, lastEventDate, nowReference)

		if !errors.Is(err, weagingdomain.ErrQuantityExceedsMax) {
			t.Fatalf("error = %v, want ErrQuantityExceedsMax", err)
		}
	})

	t.Run("accepts a missing weight", func(t *testing.T) {
		params := validParams()
		params.TotalWeight = nil

		weaging, err := weagingdomain.NewWeaging(params, farrowDate, lastEventDate, nowReference)

		if err != nil {
			t.Fatalf("NewWeaging() error = %v", err)
		}
		if weaging.TotalWeight != nil {
			t.Fatalf("total weight = %v, want nil", weaging.TotalWeight)
		}
	})

	t.Run("rejects a weight above the maximum", func(t *testing.T) {
		params := validParams()
		weight := 200.01
		params.TotalWeight = &weight

		_, err := weagingdomain.NewWeaging(params, farrowDate, lastEventDate, nowReference)

		if !errors.Is(err, weagingdomain.ErrTotalWeightExceedsMax) {
			t.Fatalf("error = %v, want ErrTotalWeightExceedsMax", err)
		}
	})

	t.Run("rejects a destination that is too long", func(t *testing.T) {
		params := validParams()
		long := string(make([]rune, 101))
		params.Destination = &long

		_, err := weagingdomain.NewWeaging(params, farrowDate, lastEventDate, nowReference)

		if !errors.Is(err, weagingdomain.ErrInvalidDestination) {
			t.Fatalf("error = %v, want ErrInvalidDestination", err)
		}
	})

	t.Run("rejects a note that is too long", func(t *testing.T) {
		params := validParams()
		long := string(make([]rune, 501))
		params.Note = &long

		_, err := weagingdomain.NewWeaging(params, farrowDate, lastEventDate, nowReference)

		if !errors.Is(err, weagingdomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
	})

	t.Run("rejects a weaging date on or before the farrow date", func(t *testing.T) {
		params := validParams()
		params.WeagingDate = farrowDate

		_, err := weagingdomain.NewWeaging(params, farrowDate, lastEventDate, nowReference)

		if !errors.Is(err, weagingdomain.ErrWeagingDateBeforeFarrowing) {
			t.Fatalf("error = %v, want ErrWeagingDateBeforeFarrowing", err)
		}
	})

	t.Run("accepts a weaging date on the last event day", func(t *testing.T) {
		params := validParams()
		params.WeagingDate = lastEventDate

		_, err := weagingdomain.NewWeaging(params, farrowDate, lastEventDate, nowReference)

		if err != nil {
			t.Fatalf("NewWeaging() error = %v", err)
		}
	})

	t.Run("rejects a weaging date before the last event", func(t *testing.T) {
		params := validParams()
		eventDate := params.WeagingDate.AddDate(0, 0, 1)

		_, err := weagingdomain.NewWeaging(params, farrowDate, eventDate, nowReference)

		if !errors.Is(err, weagingdomain.ErrWeagingDateBeforeEvents) {
			t.Fatalf("error = %v, want ErrWeagingDateBeforeEvents", err)
		}
	})

	t.Run("skips the event check when there is no event", func(t *testing.T) {
		params := validParams()

		if _, err := weagingdomain.NewWeaging(params, farrowDate, time.Time{}, nowReference); err != nil {
			t.Fatalf("NewWeaging() error = %v", err)
		}
	})

	t.Run("rejects a weaging date in the future", func(t *testing.T) {
		params := validParams()
		params.WeagingDate = nowReference.AddDate(0, 0, 1)

		_, err := weagingdomain.NewWeaging(params, farrowDate, lastEventDate, nowReference)

		if !errors.Is(err, weagingdomain.ErrWeagingDateInFuture) {
			t.Fatalf("error = %v, want ErrWeagingDateInFuture", err)
		}
	})
}

func updateTestWeaging() *weagingdomain.Weaging {
	weight := 120.5
	destination := "Nave 2"
	note := "Destete completo de la camada"
	actor := uuid.New()
	return &weagingdomain.Weaging{
		ID:          uuid.New(),
		FarrowingID: uuid.New(),
		SowID:       uuid.New(),
		WeagingDate: time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC),
		Quantity:    8,
		TotalWeight: &weight,
		Destination: &destination,
		Note:        &note,
		CreatedAt:   nowReference,
		UpdatedAt:   nowReference,
		CreatedBy:   actor,
		UpdatedBy:   actor,
	}
}

func TestUpdateWeaging(t *testing.T) {
	updatedBy := uuid.New()

	t.Run("updates the editable fields", func(t *testing.T) {
		weaging := updateTestWeaging()
		farrowingID := weaging.FarrowingID
		sowID := weaging.SowID
		weagingDate := weaging.WeagingDate
		weight := 55.5
		destination := "  Nave 3  "
		note := "  Nueva nota  "

		err := weaging.Update(weagingdomain.UpdateWeagingParams{
			TotalWeight: &weight,
			Destination: &destination,
			Note:        &note,
		}, updatedBy, nowReference)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if weaging.TotalWeight == nil || *weaging.TotalWeight != 55.5 {
			t.Fatalf("unexpected weight: %#v", weaging.TotalWeight)
		}
		if weaging.Destination == nil || *weaging.Destination != "Nave 3" {
			t.Fatalf("unexpected destination: %#v", weaging.Destination)
		}
		if weaging.Note == nil || *weaging.Note != "Nueva nota" {
			t.Fatalf("unexpected note: %#v", weaging.Note)
		}
		if weaging.UpdatedBy != updatedBy || !weaging.UpdatedAt.Equal(nowReference) {
			t.Fatalf("unexpected audit fields: %#v", weaging)
		}
		if weaging.FarrowingID != farrowingID || weaging.SowID != sowID ||
			!weaging.WeagingDate.Equal(weagingDate) || weaging.Quantity != 8 {
			t.Fatalf("immutable fields changed: %#v", weaging)
		}
	})

	t.Run("clears the weight", func(t *testing.T) {
		weaging := updateTestWeaging()

		err := weaging.Update(weagingdomain.UpdateWeagingParams{
			ClearTotalWeight: true,
		}, updatedBy, nowReference)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if weaging.TotalWeight != nil {
			t.Fatalf("weight = %v, want nil", weaging.TotalWeight)
		}
	})

	t.Run("clears the destination with a blank value", func(t *testing.T) {
		weaging := updateTestWeaging()
		blank := "   "

		err := weaging.Update(weagingdomain.UpdateWeagingParams{
			Destination: &blank,
		}, updatedBy, nowReference)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if weaging.Destination != nil {
			t.Fatalf("destination = %v, want nil", weaging.Destination)
		}
	})

	t.Run("clears the note with a blank value", func(t *testing.T) {
		weaging := updateTestWeaging()
		blank := "   "

		err := weaging.Update(weagingdomain.UpdateWeagingParams{
			Note: &blank,
		}, updatedBy, nowReference)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if weaging.Note != nil {
			t.Fatalf("note = %v, want nil", weaging.Note)
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		weaging := updateTestWeaging()

		err := weaging.Update(weagingdomain.UpdateWeagingParams{}, updatedBy, nowReference)

		if !errors.Is(err, weagingdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects setting and clearing the weight together", func(t *testing.T) {
		weaging := updateTestWeaging()
		weight := 10.0

		err := weaging.Update(weagingdomain.UpdateWeagingParams{
			TotalWeight:      &weight,
			ClearTotalWeight: true,
		}, updatedBy, nowReference)

		if !errors.Is(err, weagingdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects a nil weaging", func(t *testing.T) {
		var weaging *weagingdomain.Weaging

		err := weaging.Update(weagingdomain.UpdateWeagingParams{
			ClearTotalWeight: true,
		}, updatedBy, nowReference)

		if !errors.Is(err, weagingdomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("rejects a missing actor", func(t *testing.T) {
		weaging := updateTestWeaging()

		err := weaging.Update(weagingdomain.UpdateWeagingParams{
			ClearTotalWeight: true,
		}, uuid.Nil, nowReference)

		if !errors.Is(err, weagingdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})

	t.Run("validates the weight", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			weight float64
			want   error
		}{
			{"zero", 0, weagingdomain.ErrInvalidTotalWeight},
			{"negative", -1, weagingdomain.ErrInvalidTotalWeight},
			{"above maximum", 200.5, weagingdomain.ErrTotalWeightExceedsMax},
		} {
			t.Run(test.name, func(t *testing.T) {
				weaging := updateTestWeaging()
				weight := test.weight

				err := weaging.Update(weagingdomain.UpdateWeagingParams{
					TotalWeight: &weight,
				}, updatedBy, nowReference)

				if !errors.Is(err, test.want) {
					t.Fatalf("error = %v, want %v", err, test.want)
				}
			})
		}
	})

	t.Run("accepts a weight of 200", func(t *testing.T) {
		weaging := updateTestWeaging()
		weight := 200.0

		err := weaging.Update(weagingdomain.UpdateWeagingParams{
			TotalWeight: &weight,
		}, updatedBy, nowReference)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
	})

	t.Run("rejects a destination that is too long", func(t *testing.T) {
		weaging := updateTestWeaging()
		long := strings.Repeat("a", 101)

		err := weaging.Update(weagingdomain.UpdateWeagingParams{
			Destination: &long,
		}, updatedBy, nowReference)

		if !errors.Is(err, weagingdomain.ErrInvalidDestination) {
			t.Fatalf("error = %v, want ErrInvalidDestination", err)
		}
	})

	t.Run("rejects a note that is too long", func(t *testing.T) {
		weaging := updateTestWeaging()
		long := strings.Repeat("a", 501)

		err := weaging.Update(weagingdomain.UpdateWeagingParams{
			Note: &long,
		}, updatedBy, nowReference)

		if !errors.Is(err, weagingdomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
	})
}
