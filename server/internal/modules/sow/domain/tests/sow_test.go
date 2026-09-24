package tests

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	sowdomain "server/internal/modules/sow/domain"
)

func validParams() sowdomain.NewSowParams {
	entry := time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	location := "  Corral A  "
	note := "  Cerda reproductora  "
	return sowdomain.NewSowParams{
		ID:        uuid.New(),
		Code:      "  C-001  ",
		Location:  &location,
		EntryDate: entry,
		Note:      &note,
		Origin:    sowdomain.OriginOwn,
		Parity:    3,
		BreedID:   uuid.New(),
		CreatedBy: uuid.New(),
	}
}

func TestNewSow(t *testing.T) {
	now := time.Date(2026, time.February, 1, 3, 4, 5, 0, time.UTC)

	t.Run("creates a valid active sow with defaults", func(t *testing.T) {
		params := validParams()

		sow, err := sowdomain.NewSow(params, now)

		if err != nil {
			t.Fatalf("NewSow() error = %v", err)
		}
		if sow.ID != params.ID || sow.Code != "C-001" || !sow.Active {
			t.Fatalf("unexpected sow: %#v", sow)
		}
		if sow.Location == nil || *sow.Location != "Corral A" {
			t.Fatalf("unexpected location: %#v", sow.Location)
		}
		if sow.Note == nil || *sow.Note != "Cerda reproductora" {
			t.Fatalf("unexpected note: %#v", sow.Note)
		}
		if sow.State != sowdomain.StateAlive {
			t.Fatalf("state = %v, want %v", sow.State, sowdomain.StateAlive)
		}
		if sow.Origin != sowdomain.OriginOwn {
			t.Fatalf("origin = %v, want %v", sow.Origin, sowdomain.OriginOwn)
		}
		if sow.Parity != 3 {
			t.Fatalf("parity = %d, want 3", sow.Parity)
		}
		if sow.BreedID != params.BreedID || sow.CreatedBy != params.CreatedBy || sow.UpdatedBy != params.CreatedBy {
			t.Fatalf("unexpected references: %#v", sow)
		}
		if !sow.CreatedAt.Equal(now) || !sow.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected timestamps: %#v", sow)
		}
	})

	t.Run("accepts a valid birth date", func(t *testing.T) {
		params := validParams()
		birth := time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)
		params.BirthDate = &birth

		sow, err := sowdomain.NewSow(params, now)

		if err != nil || sow.BirthDate == nil {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("allows nil optional fields", func(t *testing.T) {
		params := validParams()
		params.Location = nil
		params.Note = nil

		sow, err := sowdomain.NewSow(params, now)

		if err != nil || sow.Location != nil || sow.Note != nil {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("normalizes empty optional fields to nil", func(t *testing.T) {
		params := validParams()
		empty := "   "
		params.Location = &empty
		params.Note = &empty

		sow, err := sowdomain.NewSow(params, now)

		if err != nil || sow.Location != nil || sow.Note != nil {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("rejects nil id", func(t *testing.T) {
		params := validParams()
		params.ID = uuid.Nil

		_, err := sowdomain.NewSow(params, now)
		if !errors.Is(err, sowdomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("rejects invalid code", func(t *testing.T) {
		params := validParams()
		params.Code = "  "

		_, err := sowdomain.NewSow(params, now)
		if !errors.Is(err, sowdomain.ErrInvalidCode) {
			t.Fatalf("error = %v, want ErrInvalidCode", err)
		}

		params.Code = strings.Repeat("a", 51)
		_, err = sowdomain.NewSow(params, now)
		if !errors.Is(err, sowdomain.ErrInvalidCode) {
			t.Fatalf("error = %v, want ErrInvalidCode", err)
		}
	})

	t.Run("accepts code at exactly max length", func(t *testing.T) {
		params := validParams()
		params.Code = strings.Repeat("a", 50)

		sow, err := sowdomain.NewSow(params, now)
		if err != nil || sow.Code != params.Code {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("rejects over-length location and note", func(t *testing.T) {
		params := validParams()
		longLocation := strings.Repeat("a", 101)
		params.Location = &longLocation

		_, err := sowdomain.NewSow(params, now)
		if !errors.Is(err, sowdomain.ErrInvalidLocation) {
			t.Fatalf("error = %v, want ErrInvalidLocation", err)
		}

		params = validParams()
		longNote := strings.Repeat("a", 501)
		params.Note = &longNote

		_, err = sowdomain.NewSow(params, now)
		if !errors.Is(err, sowdomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
	})

	t.Run("rejects zero entry date", func(t *testing.T) {
		params := validParams()
		params.EntryDate = time.Time{}

		_, err := sowdomain.NewSow(params, now)
		if !errors.Is(err, sowdomain.ErrInvalidEntryDate) {
			t.Fatalf("error = %v, want ErrInvalidEntryDate", err)
		}
	})

	t.Run("rejects birth date after entry date", func(t *testing.T) {
		params := validParams()
		birth := params.EntryDate.Add(24 * time.Hour)
		params.BirthDate = &birth

		_, err := sowdomain.NewSow(params, now)
		if !errors.Is(err, sowdomain.ErrInvalidBirthDate) {
			t.Fatalf("error = %v, want ErrInvalidBirthDate", err)
		}
	})

	t.Run("accepts an external origin", func(t *testing.T) {
		params := validParams()
		params.Origin = sowdomain.OriginExternal

		sow, err := sowdomain.NewSow(params, now)

		if err != nil || sow.Origin != sowdomain.OriginExternal {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("rejects an invalid origin", func(t *testing.T) {
		params := validParams()
		params.Origin = ""

		_, err := sowdomain.NewSow(params, now)
		if !errors.Is(err, sowdomain.ErrInvalidOrigin) {
			t.Fatalf("error = %v, want ErrInvalidOrigin", err)
		}

		params.Origin = "Desconocido"
		_, err = sowdomain.NewSow(params, now)
		if !errors.Is(err, sowdomain.ErrInvalidOrigin) {
			t.Fatalf("error = %v, want ErrInvalidOrigin", err)
		}
	})

	t.Run("accepts zero parity", func(t *testing.T) {
		params := validParams()
		params.Parity = 0

		sow, err := sowdomain.NewSow(params, now)

		if err != nil || sow.Parity != 0 {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("rejects negative parity", func(t *testing.T) {
		params := validParams()
		params.Parity = -1

		_, err := sowdomain.NewSow(params, now)
		if !errors.Is(err, sowdomain.ErrInvalidParity) {
			t.Fatalf("error = %v, want ErrInvalidParity", err)
		}
	})

	t.Run("rejects nil breed", func(t *testing.T) {
		params := validParams()
		params.BreedID = uuid.Nil

		_, err := sowdomain.NewSow(params, now)
		if !errors.Is(err, sowdomain.ErrInvalidBreed) {
			t.Fatalf("error = %v, want ErrInvalidBreed", err)
		}
	})

	t.Run("rejects nil created by", func(t *testing.T) {
		params := validParams()
		params.CreatedBy = uuid.Nil

		_, err := sowdomain.NewSow(params, now)
		if !errors.Is(err, sowdomain.ErrInvalidCreatedBy) {
			t.Fatalf("error = %v, want ErrInvalidCreatedBy", err)
		}
	})
}

func TestSowUpdate(t *testing.T) {
	now := time.Date(2026, time.March, 5, 6, 7, 8, 0, time.UTC)
	updatedBy := uuid.New()

	newSow := func() *sowdomain.Sow {
		return &sowdomain.Sow{
			ID:        uuid.New(),
			Code:      "C-001",
			Active:    true,
			EntryDate: time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC),
			State:     sowdomain.StateAlive,
			Origin:    sowdomain.OriginOwn,
			Parity:    2,
			BreedID:   uuid.New(),
		}
	}

	t.Run("updates provided fields", func(t *testing.T) {
		sow := newSow()
		code := "  C-002  "
		location := "  Corral B  "
		note := "  Actualizada  "
		active := false
		breedID := uuid.New()
		entry := sow.EntryDate.Add(48 * time.Hour)

		err := sow.Update(sowdomain.UpdateSowParams{
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
		if sow.Code != "C-002" || sow.Active {
			t.Fatalf("unexpected sow: %#v", sow)
		}
		if sow.Location == nil || *sow.Location != "Corral B" || sow.Note == nil || *sow.Note != "Actualizada" {
			t.Fatalf("unexpected optional fields: %#v", sow)
		}
		if !sow.EntryDate.Equal(entry) || sow.BreedID != breedID {
			t.Fatalf("unexpected fields: %#v", sow)
		}
		if sow.UpdatedBy != updatedBy || !sow.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected audit fields: %#v", sow)
		}
	})

	t.Run("clears location and note with empty strings", func(t *testing.T) {
		sow := newSow()
		location := "Corral A"
		note := "Nota"
		sow.Location = &location
		sow.Note = &note
		empty := "   "

		err := sow.Update(sowdomain.UpdateSowParams{
			Location: &empty,
			Note:     &empty,
		}, updatedBy, now)

		if err != nil || sow.Location != nil || sow.Note != nil {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("clears the birth date with ClearBirthDate", func(t *testing.T) {
		sow := newSow()
		birth := sow.EntryDate.Add(-24 * time.Hour)
		sow.BirthDate = &birth

		err := sow.Update(sowdomain.UpdateSowParams{ClearBirthDate: true}, updatedBy, now)

		if err != nil || sow.BirthDate != nil {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
		if sow.UpdatedBy != updatedBy || !sow.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected audit fields: %#v", sow)
		}
	})

	t.Run("never changes the state", func(t *testing.T) {
		sow := newSow()
		code := "C-010"

		err := sow.Update(sowdomain.UpdateSowParams{Code: &code}, updatedBy, now)

		if err != nil || sow.State != sowdomain.StateAlive {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("never changes the parity", func(t *testing.T) {
		sow := newSow()
		code := "C-010"

		err := sow.Update(sowdomain.UpdateSowParams{Code: &code}, updatedBy, now)

		if err != nil || sow.Parity != 2 {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("keeps omitted fields unchanged", func(t *testing.T) {
		sow := newSow()
		code := "C-010"

		err := sow.Update(sowdomain.UpdateSowParams{Code: &code}, updatedBy, now)

		if err != nil || sow.Code != code || !sow.Active || sow.State != sowdomain.StateAlive {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("updates the origin", func(t *testing.T) {
		sow := newSow()
		sow.Origin = sowdomain.OriginOwn
		origin := sowdomain.OriginExternal

		err := sow.Update(sowdomain.UpdateSowParams{Origin: &origin}, updatedBy, now)

		if err != nil || sow.Origin != sowdomain.OriginExternal {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("rejects an invalid origin without mutating the sow", func(t *testing.T) {
		sow := newSow()
		sow.Origin = sowdomain.OriginOwn
		origin := sowdomain.Origin("Desconocido")

		err := sow.Update(sowdomain.UpdateSowParams{Origin: &origin}, updatedBy, now)

		if !errors.Is(err, sowdomain.ErrInvalidOrigin) || sow.Origin != sowdomain.OriginOwn {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("rejects empty update", func(t *testing.T) {
		sow := newSow()

		err := sow.Update(sowdomain.UpdateSowParams{}, updatedBy, now)

		if !errors.Is(err, sowdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects nil updated by", func(t *testing.T) {
		sow := newSow()
		code := "C-011"

		err := sow.Update(sowdomain.UpdateSowParams{Code: &code}, uuid.Nil, now)

		if !errors.Is(err, sowdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})

	t.Run("rejects nil receiver", func(t *testing.T) {
		var sow *sowdomain.Sow
		code := "C-012"

		err := sow.Update(sowdomain.UpdateSowParams{Code: &code}, updatedBy, now)

		if !errors.Is(err, sowdomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("rejects invalid code without mutating the sow", func(t *testing.T) {
		sow := newSow()
		code := "  "

		err := sow.Update(sowdomain.UpdateSowParams{Code: &code}, updatedBy, now)

		if !errors.Is(err, sowdomain.ErrInvalidCode) || sow.Code != "C-001" {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("rejects birth date after updated entry date", func(t *testing.T) {
		sow := newSow()
		birth := sow.EntryDate
		entry := sow.EntryDate.Add(-48 * time.Hour)

		err := sow.Update(sowdomain.UpdateSowParams{
			EntryDate: &entry,
			BirthDate: &birth,
		}, updatedBy, now)

		if !errors.Is(err, sowdomain.ErrInvalidBirthDate) {
			t.Fatalf("error = %v, want ErrInvalidBirthDate", err)
		}
	})

	t.Run("rejects existing birth date after new entry date", func(t *testing.T) {
		sow := newSow()
		birth := sow.EntryDate
		sow.BirthDate = &birth
		entry := sow.EntryDate.Add(-48 * time.Hour)

		err := sow.Update(sowdomain.UpdateSowParams{EntryDate: &entry}, updatedBy, now)

		if !errors.Is(err, sowdomain.ErrInvalidBirthDate) {
			t.Fatalf("error = %v, want ErrInvalidBirthDate", err)
		}
	})

	t.Run("rejects nil breed id", func(t *testing.T) {
		sow := newSow()
		breedID := uuid.Nil

		err := sow.Update(sowdomain.UpdateSowParams{BreedID: &breedID}, updatedBy, now)

		if !errors.Is(err, sowdomain.ErrInvalidBreed) {
			t.Fatalf("error = %v, want ErrInvalidBreed", err)
		}
	})
}

func TestSowChangeState(t *testing.T) {
	now := time.Date(2026, time.April, 1, 2, 3, 4, 0, time.UTC)
	updatedBy := uuid.New()

	newSow := func() *sowdomain.Sow {
		return &sowdomain.Sow{ID: uuid.New(), Code: "C-001", Active: true, State: sowdomain.StateAlive}
	}

	t.Run("changes the state and records the audit", func(t *testing.T) {
		sow := newSow()

		err := sow.ChangeState(sowdomain.StatePregnant, updatedBy, now)

		if err != nil {
			t.Fatalf("ChangeState() error = %v", err)
		}
		if sow.State != sowdomain.StatePregnant || sow.UpdatedBy != updatedBy || !sow.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected sow: %#v", sow)
		}
	})

	t.Run("accepts every valid state", func(t *testing.T) {
		for _, state := range []sowdomain.State{
			sowdomain.StateAlive,
			sowdomain.StateDead,
			sowdomain.StateDiscarded,
			sowdomain.StateSacrificed,
			sowdomain.StateAborted,
			sowdomain.StatePregnant,
			sowdomain.StateLactating,
			sowdomain.StateWeaned,
		} {
			sow := newSow()
			if err := sow.ChangeState(state, updatedBy, now); err != nil || sow.State != state {
				t.Fatalf("state=%v err=%v sow=%#v", state, err, sow)
			}
		}
	})

	t.Run("rejects invalid state without mutating the sow", func(t *testing.T) {
		sow := newSow()

		err := sow.ChangeState(sowdomain.State("Unknown"), updatedBy, now)

		if !errors.Is(err, sowdomain.ErrInvalidState) || sow.State != sowdomain.StateAlive {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("rejects nil updated by", func(t *testing.T) {
		sow := newSow()

		err := sow.ChangeState(sowdomain.StateDead, uuid.Nil, now)

		if !errors.Is(err, sowdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})

	t.Run("rejects nil receiver", func(t *testing.T) {
		var sow *sowdomain.Sow

		err := sow.ChangeState(sowdomain.StateDead, updatedBy, now)

		if !errors.Is(err, sowdomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})
}

func TestParseOrigin(t *testing.T) {
	t.Run("parses a known origin", func(t *testing.T) {
		origin, err := sowdomain.ParseOrigin("  Externo  ")

		if err != nil || origin != sowdomain.OriginExternal {
			t.Fatalf("unexpected result: err=%v origin=%v", err, origin)
		}
	})

	t.Run("rejects an unknown origin", func(t *testing.T) {
		_, err := sowdomain.ParseOrigin("Desconocido")

		if !errors.Is(err, sowdomain.ErrInvalidOrigin) {
			t.Fatalf("error = %v, want ErrInvalidOrigin", err)
		}
	})
}

func TestParseState(t *testing.T) {
	t.Run("parses a known state", func(t *testing.T) {
		state, err := sowdomain.ParseState("  Gestando  ")

		if err != nil || state != sowdomain.StatePregnant {
			t.Fatalf("unexpected result: err=%v state=%v", err, state)
		}
	})

	t.Run("rejects an unknown state", func(t *testing.T) {
		_, err := sowdomain.ParseState("Desconocido")

		if !errors.Is(err, sowdomain.ErrInvalidState) {
			t.Fatalf("error = %v, want ErrInvalidState", err)
		}
	})
}
