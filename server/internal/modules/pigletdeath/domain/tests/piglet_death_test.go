package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
)

func validDeathParams() pigletdeathdomain.NewPigletDeathParams {
	weight := 1.5
	return pigletdeathdomain.NewPigletDeathParams{
		ID:           uuid.New(),
		FarrowingID:  uuid.New(),
		SowID:        uuid.New(),
		OperatorID:   uuid.New(),
		OperatorName: "Operador",
		DeathDate:    time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC),
		Quantity:     2,
		Weight:       &weight,
		Cause:        pigletdeathdomain.CauseCrushed,
		Turn:         pigletdeathdomain.TurnMorning,
		CreatedBy:    uuid.New(),
	}
}

func TestNewPigletDeath(t *testing.T) {
	now := time.Date(2026, time.April, 30, 3, 4, 5, 0, time.UTC)
	farrowDate := time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC)

	t.Run("creates a valid piglet death", func(t *testing.T) {
		params := validDeathParams()

		death, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, now)

		if err != nil {
			t.Fatalf("NewPigletDeath() error = %v", err)
		}
		if death.Quantity != 2 || death.Cause != pigletdeathdomain.CauseCrushed || death.Turn != pigletdeathdomain.TurnMorning {
			t.Fatalf("unexpected death: %#v", death)
		}
		if death.OperatorName != "Operador" {
			t.Fatalf("operator name = %q, want Operador", death.OperatorName)
		}
		if death.Weight == nil || *death.Weight != 1.5 {
			t.Fatalf("weight = %v, want 1.5", death.Weight)
		}
		if !death.CreatedAt.Equal(now) || death.CreatedBy != death.UpdatedBy {
			t.Fatalf("unexpected audit fields: %#v", death)
		}
	})

	t.Run("trims and clears an empty note", func(t *testing.T) {
		params := validDeathParams()
		blank := "   "
		params.Note = &blank

		death, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, now)

		if err != nil {
			t.Fatalf("NewPigletDeath() error = %v", err)
		}
		if death.Note != nil {
			t.Fatalf("note = %v, want nil", death.Note)
		}
	})

	t.Run("rejects required identifiers and dates", func(t *testing.T) {
		for _, test := range []struct {
			name string
			set  func(*pigletdeathdomain.NewPigletDeathParams)
			want error
		}{
			{"id", func(p *pigletdeathdomain.NewPigletDeathParams) { p.ID = uuid.Nil }, pigletdeathdomain.ErrInvalidID},
			{"farrowing", func(p *pigletdeathdomain.NewPigletDeathParams) { p.FarrowingID = uuid.Nil }, pigletdeathdomain.ErrInvalidFarrowing},
			{"sow", func(p *pigletdeathdomain.NewPigletDeathParams) { p.SowID = uuid.Nil }, pigletdeathdomain.ErrInvalidSow},
			{"operator", func(p *pigletdeathdomain.NewPigletDeathParams) { p.OperatorID = uuid.Nil }, pigletdeathdomain.ErrInvalidOperator},
			{"date", func(p *pigletdeathdomain.NewPigletDeathParams) { p.DeathDate = time.Time{} }, pigletdeathdomain.ErrInvalidDeathDate},
			{"created by", func(p *pigletdeathdomain.NewPigletDeathParams) { p.CreatedBy = uuid.Nil }, pigletdeathdomain.ErrInvalidCreatedBy},
		} {
			t.Run(test.name, func(t *testing.T) {
				params := validDeathParams()
				test.set(&params)

				_, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, now)

				if !errors.Is(err, test.want) {
					t.Fatalf("error = %v, want %v", err, test.want)
				}
			})
		}
	})

	t.Run("rejects a zero farrow date", func(t *testing.T) {
		_, err := pigletdeathdomain.NewPigletDeath(validDeathParams(), time.Time{}, now)

		if !errors.Is(err, pigletdeathdomain.ErrInvalidFarrowDate) {
			t.Fatalf("error = %v, want ErrInvalidFarrowDate", err)
		}
	})

	t.Run("rejects a death date on or before the farrowing", func(t *testing.T) {
		params := validDeathParams()
		params.DeathDate = farrowDate

		_, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, now)

		if !errors.Is(err, pigletdeathdomain.ErrDeathDateBeforeFarrowing) {
			t.Fatalf("error = %v, want ErrDeathDateBeforeFarrowing", err)
		}
	})

	t.Run("rejects a death date in the future", func(t *testing.T) {
		params := validDeathParams()
		params.DeathDate = now.AddDate(0, 0, 1)

		_, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, now)

		if !errors.Is(err, pigletdeathdomain.ErrDeathDateInFuture) {
			t.Fatalf("error = %v, want ErrDeathDateInFuture", err)
		}
	})

	t.Run("rejects a non-positive quantity", func(t *testing.T) {
		params := validDeathParams()
		params.Quantity = 0

		_, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, now)

		if !errors.Is(err, pigletdeathdomain.ErrInvalidQuantity) {
			t.Fatalf("error = %v, want ErrInvalidQuantity", err)
		}
	})

	t.Run("rejects a non-positive weight", func(t *testing.T) {
		params := validDeathParams()
		zero := 0.0
		params.Weight = &zero

		_, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, now)

		if !errors.Is(err, pigletdeathdomain.ErrInvalidWeight) {
			t.Fatalf("error = %v, want ErrInvalidWeight", err)
		}
	})

	t.Run("accepts a nil weight", func(t *testing.T) {
		params := validDeathParams()
		params.Weight = nil

		death, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, now)

		if err != nil {
			t.Fatalf("NewPigletDeath() error = %v", err)
		}
		if death.Weight != nil {
			t.Fatalf("weight = %v, want nil", death.Weight)
		}
	})

	t.Run("rejects an invalid cause and turn", func(t *testing.T) {
		params := validDeathParams()
		params.Cause = "Desconocida"

		if _, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, now); !errors.Is(err, pigletdeathdomain.ErrInvalidCause) {
			t.Fatalf("error = %v, want ErrInvalidCause", err)
		}

		params = validDeathParams()
		params.Turn = "Noche"

		if _, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, now); !errors.Is(err, pigletdeathdomain.ErrInvalidTurn) {
			t.Fatalf("error = %v, want ErrInvalidTurn", err)
		}
	})

	t.Run("rejects a too long note", func(t *testing.T) {
		params := validDeathParams()
		long := string(make([]rune, 501))
		params.Note = &long

		_, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, now)

		if !errors.Is(err, pigletdeathdomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
	})
}

func TestParseCause(t *testing.T) {
	t.Run("parses a known cause", func(t *testing.T) {
		cause, err := pigletdeathdomain.ParseCause(" Bacteria ")

		if err != nil {
			t.Fatalf("ParseCause() error = %v", err)
		}
		if cause != pigletdeathdomain.CauseBacteria {
			t.Fatalf("cause = %v, want Bacteria", cause)
		}
	})

	t.Run("rejects an unknown cause", func(t *testing.T) {
		if _, err := pigletdeathdomain.ParseCause("Inexistente"); !errors.Is(err, pigletdeathdomain.ErrInvalidCause) {
			t.Fatalf("error = %v, want ErrInvalidCause", err)
		}
	})
}

func TestParseTurn(t *testing.T) {
	t.Run("parses a known turn", func(t *testing.T) {
		turn, err := pigletdeathdomain.ParseTurn("Madrugada_no_asistida")

		if err != nil {
			t.Fatalf("ParseTurn() error = %v", err)
		}
		if turn != pigletdeathdomain.TurnUnassistedDawn {
			t.Fatalf("turn = %v, want Madrugada_no_asistida", turn)
		}
	})

	t.Run("rejects an unknown turn", func(t *testing.T) {
		if _, err := pigletdeathdomain.ParseTurn("Noche"); !errors.Is(err, pigletdeathdomain.ErrInvalidTurn) {
			t.Fatalf("error = %v, want ErrInvalidTurn", err)
		}
	})
}
