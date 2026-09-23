package tests

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
)

func validParams() boardomain.NewBoarParams {
	entry := time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	location := "  Corral A  "
	note := "  Verraco reproductor  "
	return boardomain.NewBoarParams{
		ID:        uuid.New(),
		Code:      "  B-001  ",
		Location:  &location,
		EntryDate: entry,
		Note:      &note,
		Origin:    boardomain.OriginOwn,
		BreedID:   uuid.New(),
		CreatedBy: uuid.New(),
	}
}

func TestNewBoar(t *testing.T) {
	now := time.Date(2026, time.February, 1, 3, 4, 5, 0, time.UTC)

	t.Run("creates a valid active boar with defaults", func(t *testing.T) {
		params := validParams()

		boar, err := boardomain.NewBoar(params, now)

		if err != nil {
			t.Fatalf("NewBoar() error = %v", err)
		}
		if boar.ID != params.ID || boar.Code != "B-001" || !boar.Active {
			t.Fatalf("unexpected boar: %#v", boar)
		}
		if boar.Location == nil || *boar.Location != "Corral A" {
			t.Fatalf("unexpected location: %#v", boar.Location)
		}
		if boar.Note == nil || *boar.Note != "Verraco reproductor" {
			t.Fatalf("unexpected note: %#v", boar.Note)
		}
		if boar.State != boardomain.StateAlive {
			t.Fatalf("state = %v, want %v", boar.State, boardomain.StateAlive)
		}
		if boar.Origin != boardomain.OriginOwn {
			t.Fatalf("origin = %v, want %v", boar.Origin, boardomain.OriginOwn)
		}
		if boar.BreedID != params.BreedID || boar.CreatedBy != params.CreatedBy || boar.UpdatedBy != params.CreatedBy {
			t.Fatalf("unexpected references: %#v", boar)
		}
		if !boar.CreatedAt.Equal(now) || !boar.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected timestamps: %#v", boar)
		}
	})

	t.Run("accepts a valid birth date", func(t *testing.T) {
		params := validParams()
		birth := time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)
		params.BirthDate = &birth

		boar, err := boardomain.NewBoar(params, now)

		if err != nil || boar.BirthDate == nil {
			t.Fatalf("unexpected result: err=%v boar=%#v", err, boar)
		}
	})

	t.Run("allows nil optional fields", func(t *testing.T) {
		params := validParams()
		params.Location = nil
		params.Note = nil

		boar, err := boardomain.NewBoar(params, now)

		if err != nil || boar.Location != nil || boar.Note != nil {
			t.Fatalf("unexpected result: err=%v boar=%#v", err, boar)
		}
	})

	t.Run("normalizes empty optional fields to nil", func(t *testing.T) {
		params := validParams()
		empty := "   "
		params.Location = &empty
		params.Note = &empty

		boar, err := boardomain.NewBoar(params, now)

		if err != nil || boar.Location != nil || boar.Note != nil {
			t.Fatalf("unexpected result: err=%v boar=%#v", err, boar)
		}
	})

	t.Run("rejects nil id", func(t *testing.T) {
		params := validParams()
		params.ID = uuid.Nil

		_, err := boardomain.NewBoar(params, now)
		if !errors.Is(err, boardomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("rejects invalid code", func(t *testing.T) {
		params := validParams()
		params.Code = "  "

		_, err := boardomain.NewBoar(params, now)
		if !errors.Is(err, boardomain.ErrInvalidCode) {
			t.Fatalf("error = %v, want ErrInvalidCode", err)
		}

		params.Code = strings.Repeat("a", 51)
		_, err = boardomain.NewBoar(params, now)
		if !errors.Is(err, boardomain.ErrInvalidCode) {
			t.Fatalf("error = %v, want ErrInvalidCode", err)
		}
	})

	t.Run("accepts code at exactly max length", func(t *testing.T) {
		params := validParams()
		params.Code = strings.Repeat("a", 50)

		boar, err := boardomain.NewBoar(params, now)
		if err != nil || boar.Code != params.Code {
			t.Fatalf("unexpected result: err=%v boar=%#v", err, boar)
		}
	})

	t.Run("rejects over-length location and note", func(t *testing.T) {
		params := validParams()
		longLocation := strings.Repeat("a", 101)
		params.Location = &longLocation

		_, err := boardomain.NewBoar(params, now)
		if !errors.Is(err, boardomain.ErrInvalidLocation) {
			t.Fatalf("error = %v, want ErrInvalidLocation", err)
		}

		params = validParams()
		longNote := strings.Repeat("a", 501)
		params.Note = &longNote

		_, err = boardomain.NewBoar(params, now)
		if !errors.Is(err, boardomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
	})

	t.Run("rejects zero entry date", func(t *testing.T) {
		params := validParams()
		params.EntryDate = time.Time{}

		_, err := boardomain.NewBoar(params, now)
		if !errors.Is(err, boardomain.ErrInvalidEntryDate) {
			t.Fatalf("error = %v, want ErrInvalidEntryDate", err)
		}
	})

	t.Run("rejects birth date after entry date", func(t *testing.T) {
		params := validParams()
		birth := params.EntryDate.Add(24 * time.Hour)
		params.BirthDate = &birth

		_, err := boardomain.NewBoar(params, now)
		if !errors.Is(err, boardomain.ErrInvalidBirthDate) {
			t.Fatalf("error = %v, want ErrInvalidBirthDate", err)
		}
	})

	t.Run("accepts an external origin", func(t *testing.T) {
		params := validParams()
		params.Origin = boardomain.OriginExternal

		boar, err := boardomain.NewBoar(params, now)

		if err != nil || boar.Origin != boardomain.OriginExternal {
			t.Fatalf("unexpected result: err=%v boar=%#v", err, boar)
		}
	})

	t.Run("rejects an invalid origin", func(t *testing.T) {
		params := validParams()
		params.Origin = ""

		_, err := boardomain.NewBoar(params, now)
		if !errors.Is(err, boardomain.ErrInvalidOrigin) {
			t.Fatalf("error = %v, want ErrInvalidOrigin", err)
		}

		params.Origin = "Desconocido"
		_, err = boardomain.NewBoar(params, now)
		if !errors.Is(err, boardomain.ErrInvalidOrigin) {
			t.Fatalf("error = %v, want ErrInvalidOrigin", err)
		}
	})

	t.Run("rejects nil breed", func(t *testing.T) {
		params := validParams()
		params.BreedID = uuid.Nil

		_, err := boardomain.NewBoar(params, now)
		if !errors.Is(err, boardomain.ErrInvalidBreed) {
			t.Fatalf("error = %v, want ErrInvalidBreed", err)
		}
	})

	t.Run("rejects nil created by", func(t *testing.T) {
		params := validParams()
		params.CreatedBy = uuid.Nil

		_, err := boardomain.NewBoar(params, now)
		if !errors.Is(err, boardomain.ErrInvalidCreatedBy) {
			t.Fatalf("error = %v, want ErrInvalidCreatedBy", err)
		}
	})
}

func TestBoarUpdate(t *testing.T) {
	now := time.Date(2026, time.March, 5, 6, 7, 8, 0, time.UTC)
	updatedBy := uuid.New()

	newBoar := func() *boardomain.Boar {
		return &boardomain.Boar{
			ID:        uuid.New(),
			Code:      "B-001",
			Active:    true,
			EntryDate: time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC),
			State:     boardomain.StateAlive,
			BreedID:   uuid.New(),
		}
	}

	t.Run("updates provided fields", func(t *testing.T) {
		boar := newBoar()
		code := "  B-002  "
		location := "  Corral B  "
		note := "  Actualizado  "
		active := false
		breedID := uuid.New()
		entry := boar.EntryDate.Add(48 * time.Hour)

		err := boar.Update(boardomain.UpdateBoarParams{
			Code:      &code,
			Location:  &location,
			Active:    &active,
			EntryDate: &entry,
			Note:      &note,
			BreedID:   &breedID,
		}, updatedBy, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if boar.Code != "B-002" || boar.Active {
			t.Fatalf("unexpected boar: %#v", boar)
		}
		if boar.Location == nil || *boar.Location != "Corral B" || boar.Note == nil || *boar.Note != "Actualizado" {
			t.Fatalf("unexpected optional fields: %#v", boar)
		}
		if !boar.EntryDate.Equal(entry) || boar.BreedID != breedID {
			t.Fatalf("unexpected fields: %#v", boar)
		}
		if boar.UpdatedBy != updatedBy || !boar.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected audit fields: %#v", boar)
		}
	})

	t.Run("clears location and note with empty strings", func(t *testing.T) {
		boar := newBoar()
		location := "Corral A"
		note := "Nota"
		boar.Location = &location
		boar.Note = &note
		empty := "   "

		err := boar.Update(boardomain.UpdateBoarParams{
			Location: &empty,
			Note:     &empty,
		}, updatedBy, now)

		if err != nil || boar.Location != nil || boar.Note != nil {
			t.Fatalf("unexpected result: err=%v boar=%#v", err, boar)
		}
	})

	t.Run("clears the birth date with ClearBirthDate", func(t *testing.T) {
		boar := newBoar()
		birth := boar.EntryDate.Add(-24 * time.Hour)
		boar.BirthDate = &birth

		err := boar.Update(boardomain.UpdateBoarParams{ClearBirthDate: true}, updatedBy, now)

		if err != nil || boar.BirthDate != nil {
			t.Fatalf("unexpected result: err=%v boar=%#v", err, boar)
		}
		if boar.UpdatedBy != updatedBy || !boar.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected audit fields: %#v", boar)
		}
	})

	t.Run("never changes the state", func(t *testing.T) {
		boar := newBoar()
		code := "B-010"

		err := boar.Update(boardomain.UpdateBoarParams{Code: &code}, updatedBy, now)

		if err != nil || boar.State != boardomain.StateAlive {
			t.Fatalf("unexpected result: err=%v boar=%#v", err, boar)
		}
	})

	t.Run("keeps omitted fields unchanged", func(t *testing.T) {
		boar := newBoar()
		code := "B-010"

		err := boar.Update(boardomain.UpdateBoarParams{Code: &code}, updatedBy, now)

		if err != nil || boar.Code != code || !boar.Active || boar.State != boardomain.StateAlive {
			t.Fatalf("unexpected result: err=%v boar=%#v", err, boar)
		}
	})

	t.Run("updates the origin", func(t *testing.T) {
		boar := newBoar()
		boar.Origin = boardomain.OriginOwn
		origin := boardomain.OriginExternal

		err := boar.Update(boardomain.UpdateBoarParams{Origin: &origin}, updatedBy, now)

		if err != nil || boar.Origin != boardomain.OriginExternal {
			t.Fatalf("unexpected result: err=%v boar=%#v", err, boar)
		}
	})

	t.Run("rejects an invalid origin without mutating the boar", func(t *testing.T) {
		boar := newBoar()
		boar.Origin = boardomain.OriginOwn
		origin := boardomain.Origin("Desconocido")

		err := boar.Update(boardomain.UpdateBoarParams{Origin: &origin}, updatedBy, now)

		if !errors.Is(err, boardomain.ErrInvalidOrigin) || boar.Origin != boardomain.OriginOwn {
			t.Fatalf("unexpected result: err=%v boar=%#v", err, boar)
		}
	})

	t.Run("rejects empty update", func(t *testing.T) {
		boar := newBoar()

		err := boar.Update(boardomain.UpdateBoarParams{}, updatedBy, now)

		if !errors.Is(err, boardomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects nil updated by", func(t *testing.T) {
		boar := newBoar()
		code := "B-011"

		err := boar.Update(boardomain.UpdateBoarParams{Code: &code}, uuid.Nil, now)

		if !errors.Is(err, boardomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})

	t.Run("rejects nil receiver", func(t *testing.T) {
		var boar *boardomain.Boar
		code := "B-012"

		err := boar.Update(boardomain.UpdateBoarParams{Code: &code}, updatedBy, now)

		if !errors.Is(err, boardomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("rejects invalid code without mutating the boar", func(t *testing.T) {
		boar := newBoar()
		code := "  "

		err := boar.Update(boardomain.UpdateBoarParams{Code: &code}, updatedBy, now)

		if !errors.Is(err, boardomain.ErrInvalidCode) || boar.Code != "B-001" {
			t.Fatalf("unexpected result: err=%v boar=%#v", err, boar)
		}
	})

	t.Run("rejects birth date after updated entry date", func(t *testing.T) {
		boar := newBoar()
		birth := boar.EntryDate
		entry := boar.EntryDate.Add(-48 * time.Hour)

		err := boar.Update(boardomain.UpdateBoarParams{
			EntryDate: &entry,
			BirthDate: &birth,
		}, updatedBy, now)

		if !errors.Is(err, boardomain.ErrInvalidBirthDate) {
			t.Fatalf("error = %v, want ErrInvalidBirthDate", err)
		}
	})

	t.Run("rejects existing birth date after new entry date", func(t *testing.T) {
		boar := newBoar()
		birth := boar.EntryDate
		boar.BirthDate = &birth
		entry := boar.EntryDate.Add(-48 * time.Hour)

		err := boar.Update(boardomain.UpdateBoarParams{EntryDate: &entry}, updatedBy, now)

		if !errors.Is(err, boardomain.ErrInvalidBirthDate) {
			t.Fatalf("error = %v, want ErrInvalidBirthDate", err)
		}
	})

	t.Run("rejects nil breed id", func(t *testing.T) {
		boar := newBoar()
		breedID := uuid.Nil

		err := boar.Update(boardomain.UpdateBoarParams{BreedID: &breedID}, updatedBy, now)

		if !errors.Is(err, boardomain.ErrInvalidBreed) {
			t.Fatalf("error = %v, want ErrInvalidBreed", err)
		}
	})
}

func TestBoarChangeState(t *testing.T) {
	now := time.Date(2026, time.April, 1, 2, 3, 4, 0, time.UTC)
	updatedBy := uuid.New()

	newBoar := func() *boardomain.Boar {
		return &boardomain.Boar{ID: uuid.New(), Code: "B-001", Active: true, State: boardomain.StateAlive}
	}

	t.Run("changes the state and records the audit", func(t *testing.T) {
		boar := newBoar()

		err := boar.ChangeState(boardomain.StateSacrificed, updatedBy, now)

		if err != nil {
			t.Fatalf("ChangeState() error = %v", err)
		}
		if boar.State != boardomain.StateSacrificed || boar.UpdatedBy != updatedBy || !boar.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected boar: %#v", boar)
		}
	})

	t.Run("accepts every valid state", func(t *testing.T) {
		for _, state := range []boardomain.State{
			boardomain.StateAlive,
			boardomain.StateDead,
			boardomain.StateDiscarded,
			boardomain.StateSacrificed,
		} {
			boar := newBoar()
			if err := boar.ChangeState(state, updatedBy, now); err != nil || boar.State != state {
				t.Fatalf("state=%v err=%v boar=%#v", state, err, boar)
			}
		}
	})

	t.Run("rejects invalid state without mutating the boar", func(t *testing.T) {
		boar := newBoar()

		err := boar.ChangeState(boardomain.State("Unknown"), updatedBy, now)

		if !errors.Is(err, boardomain.ErrInvalidState) || boar.State != boardomain.StateAlive {
			t.Fatalf("unexpected result: err=%v boar=%#v", err, boar)
		}
	})

	t.Run("rejects nil updated by", func(t *testing.T) {
		boar := newBoar()

		err := boar.ChangeState(boardomain.StateDead, uuid.Nil, now)

		if !errors.Is(err, boardomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})

	t.Run("rejects nil receiver", func(t *testing.T) {
		var boar *boardomain.Boar

		err := boar.ChangeState(boardomain.StateDead, updatedBy, now)

		if !errors.Is(err, boardomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})
}
