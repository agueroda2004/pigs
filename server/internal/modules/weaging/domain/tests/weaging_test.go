package tests

import (
	"errors"
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

	t.Run("rejects a weaging date on or before the last event", func(t *testing.T) {
		params := validParams()
		params.WeagingDate = lastEventDate

		_, err := weagingdomain.NewWeaging(params, farrowDate, lastEventDate, nowReference)

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
