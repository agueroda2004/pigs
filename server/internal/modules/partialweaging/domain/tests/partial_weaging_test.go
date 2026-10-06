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

func validParams() partialweagingdomain.NewPartialWeagingParams {
	weight := 42.5
	note := "Destete parcial de la camada"
	return partialweagingdomain.NewPartialWeagingParams{
		ID:          uuid.New(),
		FarrowingID: uuid.New(),
		SowID:       uuid.New(),
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

		weaging, err := partialweagingdomain.NewPartialWeaging(params, farrowDate, lastEventDate, nowReference)

		if err != nil {
			t.Fatalf("NewPartialWeaging() error = %v", err)
		}
		if weaging.FarrowingID != params.FarrowingID || weaging.SowID != params.SowID {
			t.Fatalf("unexpected references: %#v", weaging)
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

		weaging, err := partialweagingdomain.NewPartialWeaging(params, farrowDate, lastEventDate, nowReference)

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

				_, err := partialweagingdomain.NewPartialWeaging(params, farrowDate, lastEventDate, nowReference)

				if !errors.Is(err, test.want) {
					t.Fatalf("error = %v, want %v", err, test.want)
				}
			})
		}
	})

	t.Run("rejects a nil farrow date", func(t *testing.T) {
		params := validParams()

		_, err := partialweagingdomain.NewPartialWeaging(params, time.Time{}, lastEventDate, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrInvalidFarrowDate) {
			t.Fatalf("error = %v, want ErrInvalidFarrowDate", err)
		}
	})

	t.Run("rejects a non-positive quantity", func(t *testing.T) {
		params := validParams()
		params.Quantity = 0

		_, err := partialweagingdomain.NewPartialWeaging(params, farrowDate, lastEventDate, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrInvalidQuantity) {
			t.Fatalf("error = %v, want ErrInvalidQuantity", err)
		}
	})

	t.Run("rejects a non-positive weight", func(t *testing.T) {
		params := validParams()
		weight := 0.0
		params.TotalWeight = &weight

		_, err := partialweagingdomain.NewPartialWeaging(params, farrowDate, lastEventDate, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrInvalidTotalWeight) {
			t.Fatalf("error = %v, want ErrInvalidTotalWeight", err)
		}
	})

	t.Run("rejects an invalid type", func(t *testing.T) {
		params := validParams()
		params.Type = "Desconocido"

		_, err := partialweagingdomain.NewPartialWeaging(params, farrowDate, lastEventDate, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrInvalidType) {
			t.Fatalf("error = %v, want ErrInvalidType", err)
		}
	})

	t.Run("rejects a note that is too long", func(t *testing.T) {
		params := validParams()
		long := string(make([]rune, 501))
		params.Note = &long

		_, err := partialweagingdomain.NewPartialWeaging(params, farrowDate, lastEventDate, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
	})

	t.Run("rejects a weaging date on or before the farrow date", func(t *testing.T) {
		params := validParams()
		params.WeagingDate = farrowDate

		_, err := partialweagingdomain.NewPartialWeaging(params, farrowDate, lastEventDate, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrWeagingDateBeforeFarrowing) {
			t.Fatalf("error = %v, want ErrWeagingDateBeforeFarrowing", err)
		}
	})

	t.Run("rejects a weaging date on or before the last event", func(t *testing.T) {
		params := validParams()
		params.WeagingDate = lastEventDate

		_, err := partialweagingdomain.NewPartialWeaging(params, farrowDate, lastEventDate, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrWeagingDateBeforeEvents) {
			t.Fatalf("error = %v, want ErrWeagingDateBeforeEvents", err)
		}
	})

	t.Run("skips the event check when there is no event", func(t *testing.T) {
		params := validParams()

		if _, err := partialweagingdomain.NewPartialWeaging(params, farrowDate, time.Time{}, nowReference); err != nil {
			t.Fatalf("NewPartialWeaging() error = %v", err)
		}
	})

	t.Run("rejects a weaging date in the future", func(t *testing.T) {
		params := validParams()
		params.WeagingDate = nowReference.AddDate(0, 0, 1)

		_, err := partialweagingdomain.NewPartialWeaging(params, farrowDate, lastEventDate, nowReference)

		if !errors.Is(err, partialweagingdomain.ErrWeagingDateInFuture) {
			t.Fatalf("error = %v, want ErrWeagingDateInFuture", err)
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
