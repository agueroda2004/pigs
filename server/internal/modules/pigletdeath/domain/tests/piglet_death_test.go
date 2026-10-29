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

		death, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now)

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

	t.Run("accepts the Sin Turno shift", func(t *testing.T) {
		params := validDeathParams()
		params.Turn = pigletdeathdomain.TurnNoShift

		death, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now)

		if err != nil {
			t.Fatalf("NewPigletDeath() error = %v", err)
		}
		if death.Turn != pigletdeathdomain.TurnNoShift {
			t.Fatalf("turn = %v, want Sin Turno", death.Turn)
		}
	})

	t.Run("trims and clears an empty note", func(t *testing.T) {
		params := validDeathParams()
		blank := "   "
		params.Note = &blank

		death, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now)

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

				_, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now)

				if !errors.Is(err, test.want) {
					t.Fatalf("error = %v, want %v", err, test.want)
				}
			})
		}
	})

	t.Run("rejects a zero farrow date", func(t *testing.T) {
		_, err := pigletdeathdomain.NewPigletDeath(validDeathParams(), time.Time{}, time.Time{}, now)

		if !errors.Is(err, pigletdeathdomain.ErrInvalidFarrowDate) {
			t.Fatalf("error = %v, want ErrInvalidFarrowDate", err)
		}
	})

	t.Run("rejects a death date before the farrowing", func(t *testing.T) {
		params := validDeathParams()
		params.DeathDate = farrowDate.AddDate(0, 0, -1)

		_, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now)

		if !errors.Is(err, pigletdeathdomain.ErrDeathDateBeforeFarrowing) {
			t.Fatalf("error = %v, want ErrDeathDateBeforeFarrowing", err)
		}
	})

	t.Run("accepts a death date on the farrowing day", func(t *testing.T) {
		params := validDeathParams()
		params.DeathDate = farrowDate

		death, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now)

		if err != nil {
			t.Fatalf("NewPigletDeath() error = %v", err)
		}
		if !death.DeathDate.Equal(farrowDate) {
			t.Fatalf("death date = %v, want %v", death.DeathDate, farrowDate)
		}
	})

	t.Run("rejects a death date in the future", func(t *testing.T) {
		params := validDeathParams()
		params.DeathDate = now.AddDate(0, 0, 1)

		_, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now)

		if !errors.Is(err, pigletdeathdomain.ErrDeathDateInFuture) {
			t.Fatalf("error = %v, want ErrDeathDateInFuture", err)
		}
	})

	t.Run("rejects a non-positive quantity", func(t *testing.T) {
		params := validDeathParams()
		params.Quantity = 0

		_, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now)

		if !errors.Is(err, pigletdeathdomain.ErrInvalidQuantity) {
			t.Fatalf("error = %v, want ErrInvalidQuantity", err)
		}
	})

	t.Run("rejects a quantity greater than ten", func(t *testing.T) {
		params := validDeathParams()
		params.Quantity = 11

		_, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now)

		if !errors.Is(err, pigletdeathdomain.ErrQuantityTooHigh) {
			t.Fatalf("error = %v, want ErrQuantityTooHigh", err)
		}
	})

	t.Run("rejects a non-positive weight", func(t *testing.T) {
		params := validDeathParams()
		zero := 0.0
		params.Weight = &zero

		_, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now)

		if !errors.Is(err, pigletdeathdomain.ErrInvalidWeight) {
			t.Fatalf("error = %v, want ErrInvalidWeight", err)
		}
	})

	t.Run("accepts a missing weight", func(t *testing.T) {
		params := validDeathParams()
		params.Weight = nil

		death, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now)

		if err != nil {
			t.Fatalf("NewPigletDeath() error = %v", err)
		}
		if death.Weight != nil {
			t.Fatalf("weight = %v, want nil", death.Weight)
		}
	})

	t.Run("rejects a weight greater than twenty kilograms", func(t *testing.T) {
		params := validDeathParams()
		tooHeavy := 20.01
		params.Weight = &tooHeavy

		_, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now)

		if !errors.Is(err, pigletdeathdomain.ErrWeightTooHigh) {
			t.Fatalf("error = %v, want ErrWeightTooHigh", err)
		}
	})

	t.Run("rejects a death date before the last partial weaging", func(t *testing.T) {
		params := validDeathParams()
		lastWeaging := params.DeathDate.AddDate(0, 0, 1)

		_, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, lastWeaging, now)

		if !errors.Is(err, pigletdeathdomain.ErrDeathDateBeforeWeaging) {
			t.Fatalf("error = %v, want ErrDeathDateBeforeWeaging", err)
		}
	})

	t.Run("accepts a death date after the last partial weaging", func(t *testing.T) {
		params := validDeathParams()
		lastWeaging := farrowDate.AddDate(0, 0, 1)

		death, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, lastWeaging, now)

		if err != nil {
			t.Fatalf("NewPigletDeath() error = %v", err)
		}
		if !death.DeathDate.Equal(params.DeathDate) {
			t.Fatalf("death date = %v, want %v", death.DeathDate, params.DeathDate)
		}
	})

	t.Run("accepts a death date on the last partial weaging day", func(t *testing.T) {
		params := validDeathParams()
		lastWeaging := params.DeathDate.Add(time.Hour)

		death, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, lastWeaging, now)

		if err != nil {
			t.Fatalf("NewPigletDeath() error = %v", err)
		}
		if !death.DeathDate.Equal(params.DeathDate) {
			t.Fatalf("death date = %v, want %v", death.DeathDate, params.DeathDate)
		}
	})

	t.Run("accepts the Desaparecido cause", func(t *testing.T) {
		params := validDeathParams()
		params.Cause = pigletdeathdomain.CauseMissing

		death, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now)

		if err != nil {
			t.Fatalf("NewPigletDeath() error = %v", err)
		}
		if death.Cause != pigletdeathdomain.CauseMissing {
			t.Fatalf("cause = %v, want Desaparecido", death.Cause)
		}
	})

	t.Run("rejects an invalid cause and turn", func(t *testing.T) {
		params := validDeathParams()
		params.Cause = "Desconocida"

		if _, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now); !errors.Is(err, pigletdeathdomain.ErrInvalidCause) {
			t.Fatalf("error = %v, want ErrInvalidCause", err)
		}

		params = validDeathParams()
		params.Turn = "Noche"

		if _, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now); !errors.Is(err, pigletdeathdomain.ErrInvalidTurn) {
			t.Fatalf("error = %v, want ErrInvalidTurn", err)
		}
	})

	t.Run("rejects a too long note", func(t *testing.T) {
		params := validDeathParams()
		long := string(make([]rune, 501))
		params.Note = &long

		_, err := pigletdeathdomain.NewPigletDeath(params, farrowDate, time.Time{}, now)

		if !errors.Is(err, pigletdeathdomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
	})
}

func TestPigletDeathUpdate(t *testing.T) {
	now := time.Date(2026, time.April, 30, 3, 4, 5, 0, time.UTC)
	farrowDate := time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC)
	actor := uuid.New()

	newDeath := func(t *testing.T) *pigletdeathdomain.PigletDeath {
		t.Helper()
		death, err := pigletdeathdomain.NewPigletDeath(validDeathParams(), farrowDate, time.Time{}, now)
		if err != nil {
			t.Fatalf("NewPigletDeath() error = %v", err)
		}
		return death
	}

	weightOf := 3.5
	note := "  nueva nota  "

	t.Run("updates the editable fields", func(t *testing.T) {
		death := newDeath(t)
		operatorID := uuid.New()
		operatorName := "Nuevo operador"
		cause := pigletdeathdomain.CauseBacteria
		turn := pigletdeathdomain.TurnAfternoon

		err := death.Update(pigletdeathdomain.UpdatePigletDeathParams{
			OperatorID:   &operatorID,
			OperatorName: &operatorName,
			Weight:       &weightOf,
			Cause:        &cause,
			Turn:         &turn,
			Note:         &note,
		}, actor, now.Add(time.Hour))

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if death.OperatorID != operatorID || death.OperatorName != operatorName {
			t.Fatalf("unexpected operator: %#v", death)
		}
		if death.Weight == nil || *death.Weight != 3.5 || death.Cause != cause || death.Turn != turn {
			t.Fatalf("unexpected fields: %#v", death)
		}
		if death.Note == nil || *death.Note != "nueva nota" {
			t.Fatalf("note = %v, want nueva nota", death.Note)
		}
		if death.UpdatedBy != actor || !death.UpdatedAt.Equal(now.Add(time.Hour)) {
			t.Fatalf("unexpected audit: %#v", death)
		}
	})

	t.Run("clears the note with an empty value", func(t *testing.T) {
		death := newDeath(t)
		empty := ""

		if err := death.Update(pigletdeathdomain.UpdatePigletDeathParams{
			Weight: &weightOf,
			Note:   &empty,
		}, actor, now); err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if death.Note != nil {
			t.Fatalf("note = %v, want nil", death.Note)
		}
	})

	t.Run("accepts the locked fields when they match", func(t *testing.T) {
		death := newDeath(t)
		deathDate := death.DeathDate
		quantity := death.Quantity
		sowID := death.SowID

		err := death.Update(pigletdeathdomain.UpdatePigletDeathParams{
			DeathDate: &deathDate,
			Quantity:  &quantity,
			SowID:     &sowID,
			Weight:    &weightOf,
		}, actor, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
	})

	t.Run("rejects a change to a locked field", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			params func(*pigletdeathdomain.PigletDeath) pigletdeathdomain.UpdatePigletDeathParams
		}{
			{
				"death date",
				func(death *pigletdeathdomain.PigletDeath) pigletdeathdomain.UpdatePigletDeathParams {
					changed := death.DeathDate.AddDate(0, 0, 1)
					return pigletdeathdomain.UpdatePigletDeathParams{DeathDate: &changed, Weight: &weightOf}
				},
			},
			{
				"quantity",
				func(death *pigletdeathdomain.PigletDeath) pigletdeathdomain.UpdatePigletDeathParams {
					changed := death.Quantity + 1
					return pigletdeathdomain.UpdatePigletDeathParams{Quantity: &changed, Weight: &weightOf}
				},
			},
			{
				"sow",
				func(*pigletdeathdomain.PigletDeath) pigletdeathdomain.UpdatePigletDeathParams {
					changed := uuid.New()
					return pigletdeathdomain.UpdatePigletDeathParams{SowID: &changed, Weight: &weightOf}
				},
			},
		} {
			t.Run(test.name, func(t *testing.T) {
				death := newDeath(t)

				err := death.Update(test.params(death), actor, now)

				if !errors.Is(err, pigletdeathdomain.ErrImmutableDeathField) {
					t.Fatalf("error = %v, want ErrImmutableDeathField", err)
				}
			})
		}
	})

	t.Run("keeps the weight when it is omitted", func(t *testing.T) {
		death := newDeath(t)

		err := death.Update(pigletdeathdomain.UpdatePigletDeathParams{Note: &note}, actor, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if death.Weight == nil || *death.Weight != 1.5 {
			t.Fatalf("weight = %v, want 1.5", death.Weight)
		}
	})

	t.Run("clears the weight", func(t *testing.T) {
		death := newDeath(t)

		err := death.Update(pigletdeathdomain.UpdatePigletDeathParams{ClearWeight: true}, actor, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if death.Weight != nil {
			t.Fatalf("weight = %v, want nil", death.Weight)
		}
	})

	t.Run("rejects setting and clearing the weight together", func(t *testing.T) {
		death := newDeath(t)

		err := death.Update(pigletdeathdomain.UpdatePigletDeathParams{
			Weight:      &weightOf,
			ClearWeight: true,
		}, actor, now)

		if !errors.Is(err, pigletdeathdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("accepts the Sin Turno shift", func(t *testing.T) {
		death := newDeath(t)
		turn := pigletdeathdomain.TurnNoShift

		if err := death.Update(pigletdeathdomain.UpdatePigletDeathParams{Turn: &turn}, actor, now); err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if death.Turn != pigletdeathdomain.TurnNoShift {
			t.Fatalf("turn = %v, want Sin Turno", death.Turn)
		}
	})

	t.Run("rejects an invalid weight", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			weight float64
			want   error
		}{
			{"non-positive", 0, pigletdeathdomain.ErrInvalidWeight},
			{"too heavy", 20.01, pigletdeathdomain.ErrWeightTooHigh},
		} {
			t.Run(test.name, func(t *testing.T) {
				death := newDeath(t)
				weight := test.weight

				err := death.Update(pigletdeathdomain.UpdatePigletDeathParams{Weight: &weight}, actor, now)

				if !errors.Is(err, test.want) {
					t.Fatalf("error = %v, want %v", err, test.want)
				}
			})
		}
	})

	t.Run("accepts the Desaparecido cause", func(t *testing.T) {
		death := newDeath(t)
		cause := pigletdeathdomain.CauseMissing

		if err := death.Update(pigletdeathdomain.UpdatePigletDeathParams{Cause: &cause}, actor, now); err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if death.Cause != pigletdeathdomain.CauseMissing {
			t.Fatalf("cause = %v, want Desaparecido", death.Cause)
		}
	})

	t.Run("rejects an invalid cause and turn", func(t *testing.T) {
		death := newDeath(t)
		cause := pigletdeathdomain.PigletDeathCause("Desconocida")
		if err := death.Update(pigletdeathdomain.UpdatePigletDeathParams{
			Weight: &weightOf,
			Cause:  &cause,
		}, actor, now); !errors.Is(err, pigletdeathdomain.ErrInvalidCause) {
			t.Fatalf("error = %v, want ErrInvalidCause", err)
		}

		death = newDeath(t)
		turn := pigletdeathdomain.Turn("Noche")
		if err := death.Update(pigletdeathdomain.UpdatePigletDeathParams{
			Weight: &weightOf,
			Turn:   &turn,
		}, actor, now); !errors.Is(err, pigletdeathdomain.ErrInvalidTurn) {
			t.Fatalf("error = %v, want ErrInvalidTurn", err)
		}
	})

	t.Run("rejects a too long note", func(t *testing.T) {
		death := newDeath(t)
		long := string(make([]rune, 501))

		err := death.Update(pigletdeathdomain.UpdatePigletDeathParams{
			Weight: &weightOf,
			Note:   &long,
		}, actor, now)

		if !errors.Is(err, pigletdeathdomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		death := newDeath(t)

		err := death.Update(pigletdeathdomain.UpdatePigletDeathParams{}, actor, now)

		if !errors.Is(err, pigletdeathdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects a missing actor", func(t *testing.T) {
		death := newDeath(t)

		err := death.Update(pigletdeathdomain.UpdatePigletDeathParams{Weight: &weightOf}, uuid.Nil, now)

		if !errors.Is(err, pigletdeathdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
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

	t.Run("parses the Desaparecido cause", func(t *testing.T) {
		cause, err := pigletdeathdomain.ParseCause(" Desaparecido ")

		if err != nil {
			t.Fatalf("ParseCause() error = %v", err)
		}
		if cause != pigletdeathdomain.CauseMissing {
			t.Fatalf("cause = %v, want Desaparecido", cause)
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

	t.Run("parses the Sin Turno shift", func(t *testing.T) {
		turn, err := pigletdeathdomain.ParseTurn(" Sin Turno ")

		if err != nil {
			t.Fatalf("ParseTurn() error = %v", err)
		}
		if turn != pigletdeathdomain.TurnNoShift {
			t.Fatalf("turn = %v, want Sin Turno", turn)
		}
	})

	t.Run("rejects an unknown turn", func(t *testing.T) {
		if _, err := pigletdeathdomain.ParseTurn("Noche"); !errors.Is(err, pigletdeathdomain.ErrInvalidTurn) {
			t.Fatalf("error = %v, want ErrInvalidTurn", err)
		}
	})
}
