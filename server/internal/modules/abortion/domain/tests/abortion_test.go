package tests

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	abortiondomain "server/internal/modules/abortion/domain"
)

var (
	lastMountDay = time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	abortionDay  = time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	nowReference = time.Date(2026, time.February, 1, 3, 4, 5, 0, time.UTC)
)

func validAbortionParams() abortiondomain.NewAbortionParams {
	note := "  Aborto espontáneo  "
	return abortiondomain.NewAbortionParams{
		ID:           uuid.New(),
		SowID:        uuid.New(),
		ServiceID:    uuid.New(),
		AbortionDate: abortionDay,
		Cause:        abortiondomain.CauseInfectious,
		Note:         &note,
		CreatedBy:    uuid.New(),
	}
}

func TestNewAbortion(t *testing.T) {
	t.Run("creates a valid abortion", func(t *testing.T) {
		params := validAbortionParams()

		abortion, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{LastMountDate: lastMountDay}, nowReference)

		if err != nil {
			t.Fatalf("NewAbortion() error = %v", err)
		}
		if abortion.ID != params.ID || abortion.SowID != params.SowID || abortion.ServiceID != params.ServiceID {
			t.Fatalf("unexpected identifiers: %#v", abortion)
		}
		if !abortion.AbortionDate.Equal(abortionDay) || abortion.Cause != abortiondomain.CauseInfectious {
			t.Fatalf("unexpected date or cause: %#v", abortion)
		}
		if abortion.Note == nil || *abortion.Note != "Aborto espontáneo" {
			t.Fatalf("unexpected note: %#v", abortion.Note)
		}
		if abortion.CreatedBy != params.CreatedBy || abortion.UpdatedBy != params.CreatedBy {
			t.Fatalf("unexpected audit: %#v", abortion)
		}
		if !abortion.CreatedAt.Equal(nowReference) || !abortion.UpdatedAt.Equal(nowReference) {
			t.Fatalf("unexpected timestamps: %#v", abortion)
		}
	})

	t.Run("allows a nil note", func(t *testing.T) {
		params := validAbortionParams()
		params.Note = nil

		abortion, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{LastMountDate: lastMountDay}, nowReference)

		if err != nil || abortion.Note != nil {
			t.Fatalf("unexpected result: err=%v abortion=%#v", err, abortion)
		}
	})

	t.Run("normalizes an empty note to nil", func(t *testing.T) {
		params := validAbortionParams()
		empty := "   "
		params.Note = &empty

		abortion, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{LastMountDate: lastMountDay}, nowReference)

		if err != nil || abortion.Note != nil {
			t.Fatalf("unexpected result: err=%v abortion=%#v", err, abortion)
		}
	})

	t.Run("accepts an abortion dated today", func(t *testing.T) {
		params := validAbortionParams()
		params.AbortionDate = time.Date(2026, time.January, 15, 18, 0, 0, 0, time.UTC)
		now := time.Date(2026, time.January, 15, 20, 0, 0, 0, time.UTC)

		_, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{LastMountDate: lastMountDay}, now)

		if err != nil {
			t.Fatalf("NewAbortion() error = %v", err)
		}
	})

	t.Run("rejects a nil id", func(t *testing.T) {
		params := validAbortionParams()
		params.ID = uuid.Nil

		_, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{LastMountDate: lastMountDay}, nowReference)
		if !errors.Is(err, abortiondomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("rejects a nil sow", func(t *testing.T) {
		params := validAbortionParams()
		params.SowID = uuid.Nil

		_, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{LastMountDate: lastMountDay}, nowReference)
		if !errors.Is(err, abortiondomain.ErrInvalidSow) {
			t.Fatalf("error = %v, want ErrInvalidSow", err)
		}
	})

	t.Run("rejects a nil service", func(t *testing.T) {
		params := validAbortionParams()
		params.ServiceID = uuid.Nil

		_, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{LastMountDate: lastMountDay}, nowReference)
		if !errors.Is(err, abortiondomain.ErrInvalidService) {
			t.Fatalf("error = %v, want ErrInvalidService", err)
		}
	})

	t.Run("rejects a zero abortion date", func(t *testing.T) {
		params := validAbortionParams()
		params.AbortionDate = time.Time{}

		_, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{LastMountDate: lastMountDay}, nowReference)
		if !errors.Is(err, abortiondomain.ErrInvalidAbortionDate) {
			t.Fatalf("error = %v, want ErrInvalidAbortionDate", err)
		}
	})

	t.Run("rejects a zero last mount date", func(t *testing.T) {
		params := validAbortionParams()

		_, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{LastMountDate: time.Time{}}, nowReference)
		if !errors.Is(err, abortiondomain.ErrInvalidLastMount) {
			t.Fatalf("error = %v, want ErrInvalidLastMount", err)
		}
	})

	t.Run("rejects an invalid cause", func(t *testing.T) {
		params := validAbortionParams()
		params.Cause = "Mágica"

		_, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{LastMountDate: lastMountDay}, nowReference)
		if !errors.Is(err, abortiondomain.ErrInvalidCause) {
			t.Fatalf("error = %v, want ErrInvalidCause", err)
		}
	})

	t.Run("rejects a nil creator", func(t *testing.T) {
		params := validAbortionParams()
		params.CreatedBy = uuid.Nil

		_, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{LastMountDate: lastMountDay}, nowReference)
		if !errors.Is(err, abortiondomain.ErrInvalidCreatedBy) {
			t.Fatalf("error = %v, want ErrInvalidCreatedBy", err)
		}
	})

	t.Run("rejects an over-length note", func(t *testing.T) {
		params := validAbortionParams()
		longNote := strings.Repeat("a", 501)
		params.Note = &longNote

		_, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{LastMountDate: lastMountDay}, nowReference)
		if !errors.Is(err, abortiondomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
	})

	t.Run("rejects an abortion dated on the last mount", func(t *testing.T) {
		params := validAbortionParams()
		params.AbortionDate = lastMountDay

		_, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{LastMountDate: lastMountDay}, nowReference)
		if !errors.Is(err, abortiondomain.ErrAbortionDateBeforeMount) {
			t.Fatalf("error = %v, want ErrAbortionDateBeforeMount", err)
		}
	})

	t.Run("rejects an abortion dated before the last mount", func(t *testing.T) {
		params := validAbortionParams()
		params.AbortionDate = time.Date(2026, time.January, 9, 0, 0, 0, 0, time.UTC)

		_, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{LastMountDate: lastMountDay}, nowReference)
		if !errors.Is(err, abortiondomain.ErrAbortionDateBeforeMount) {
			t.Fatalf("error = %v, want ErrAbortionDateBeforeMount", err)
		}
	})

	t.Run("rejects a future abortion date", func(t *testing.T) {
		params := validAbortionParams()
		params.AbortionDate = time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC)

		_, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{LastMountDate: lastMountDay}, time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC))
		if !errors.Is(err, abortiondomain.ErrAbortionDateInFuture) {
			t.Fatalf("error = %v, want ErrAbortionDateInFuture", err)
		}
	})

	t.Run("rejects an abortion dated before the sow entry date", func(t *testing.T) {
		params := validAbortionParams()
		params.AbortionDate = time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC)

		_, err := abortiondomain.NewAbortion(params, abortiondomain.Reference{
			EntryDate:     time.Date(2026, time.January, 8, 0, 0, 0, 0, time.UTC),
			LastMountDate: lastMountDay,
		}, nowReference)

		if !errors.Is(err, abortiondomain.ErrAbortionDateBeforeEntry) {
			t.Fatalf("error = %v, want ErrAbortionDateBeforeEntry", err)
		}
	})
}

func TestUpdateAbortion(t *testing.T) {
	newAbortion := func() *abortiondomain.Abortion {
		abortion, err := abortiondomain.NewAbortion(validAbortionParams(), abortiondomain.Reference{LastMountDate: lastMountDay}, nowReference)
		if err != nil {
			t.Fatalf("NewAbortion() error = %v", err)
		}
		return abortion
	}
	reference := abortiondomain.Reference{
		EntryDate:     time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC),
		LastMountDate: lastMountDay,
	}
	updatedAt := time.Date(2026, time.February, 2, 8, 0, 0, 0, time.UTC)

	t.Run("updates the editable fields and the audit", func(t *testing.T) {
		abortion := newAbortion()
		updatedBy := uuid.New()
		newDate := time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC)
		newCause := abortiondomain.CauseTrauma
		note := "  actualizado  "

		err := abortion.Update(abortiondomain.UpdateAbortionParams{
			AbortionDate: &newDate,
			Cause:        &newCause,
			Note:         &note,
		}, reference, updatedBy, updatedAt)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if !abortion.AbortionDate.Equal(newDate) || abortion.Cause != newCause {
			t.Fatalf("unexpected fields: %#v", abortion)
		}
		if abortion.Note == nil || *abortion.Note != "actualizado" {
			t.Fatalf("unexpected note: %#v", abortion.Note)
		}
		if !abortion.UpdatedAt.Equal(updatedAt) || abortion.UpdatedBy != updatedBy {
			t.Fatalf("unexpected audit: %#v", abortion)
		}
	})

	t.Run("leaves the fields unchanged when nil", func(t *testing.T) {
		abortion := newAbortion()
		original := *abortion
		cause := abortiondomain.CauseOther

		err := abortion.Update(abortiondomain.UpdateAbortionParams{
			Cause: &cause,
		}, reference, uuid.New(), updatedAt)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if !abortion.AbortionDate.Equal(original.AbortionDate) || abortion.Note != original.Note {
			t.Fatalf("unchanged fields were modified: %#v", abortion)
		}
	})

	t.Run("rejects an update without fields", func(t *testing.T) {
		abortion := newAbortion()

		err := abortion.Update(abortiondomain.UpdateAbortionParams{}, reference, uuid.New(), updatedAt)

		if !errors.Is(err, abortiondomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects a changed date before the entry date", func(t *testing.T) {
		abortion := newAbortion()
		newDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)

		err := abortion.Update(abortiondomain.UpdateAbortionParams{AbortionDate: &newDate}, reference, uuid.New(), updatedAt)

		if !errors.Is(err, abortiondomain.ErrAbortionDateBeforeEntry) {
			t.Fatalf("error = %v, want ErrAbortionDateBeforeEntry", err)
		}
	})

	t.Run("rejects a changed date on or before the last mount", func(t *testing.T) {
		abortion := newAbortion()

		err := abortion.Update(abortiondomain.UpdateAbortionParams{AbortionDate: &lastMountDay}, reference, uuid.New(), updatedAt)

		if !errors.Is(err, abortiondomain.ErrAbortionDateBeforeMount) {
			t.Fatalf("error = %v, want ErrAbortionDateBeforeMount", err)
		}
	})

	t.Run("rejects a changed date in the future", func(t *testing.T) {
		abortion := newAbortion()
		newDate := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)

		err := abortion.Update(abortiondomain.UpdateAbortionParams{AbortionDate: &newDate}, reference, uuid.New(), updatedAt)

		if !errors.Is(err, abortiondomain.ErrAbortionDateInFuture) {
			t.Fatalf("error = %v, want ErrAbortionDateInFuture", err)
		}
	})

	t.Run("rejects an invalid cause", func(t *testing.T) {
		abortion := newAbortion()
		invalid := abortiondomain.Cause("Mágica")

		err := abortion.Update(abortiondomain.UpdateAbortionParams{Cause: &invalid}, reference, uuid.New(), updatedAt)

		if !errors.Is(err, abortiondomain.ErrInvalidCause) {
			t.Fatalf("error = %v, want ErrInvalidCause", err)
		}
	})

	t.Run("rejects a nil identifier", func(t *testing.T) {
		abortion := newAbortion()
		abortion.ID = uuid.Nil
		cause := abortiondomain.CauseOther

		err := abortion.Update(abortiondomain.UpdateAbortionParams{Cause: &cause}, reference, uuid.New(), updatedAt)

		if !errors.Is(err, abortiondomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("rejects a nil updater", func(t *testing.T) {
		abortion := newAbortion()
		cause := abortiondomain.CauseOther

		err := abortion.Update(abortiondomain.UpdateAbortionParams{Cause: &cause}, reference, uuid.Nil, updatedAt)

		if !errors.Is(err, abortiondomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})
}

func TestParseCause(t *testing.T) {
	t.Run("parses a known cause", func(t *testing.T) {
		cause, err := abortiondomain.ParseCause("  Traumatismo  ")

		if err != nil || cause != abortiondomain.CauseTrauma {
			t.Fatalf("unexpected result: err=%v cause=%v", err, cause)
		}
	})

	t.Run("parses every valid cause", func(t *testing.T) {
		for _, cause := range []abortiondomain.Cause{
			abortiondomain.CauseUnknown,
			abortiondomain.CauseInfectious,
			abortiondomain.CauseTrauma,
			abortiondomain.CauseManagement,
			abortiondomain.CauseNutritional,
			abortiondomain.CauseOther,
		} {
			parsed, err := abortiondomain.ParseCause(string(cause))
			if err != nil || parsed != cause {
				t.Fatalf("cause=%v err=%v parsed=%v", cause, err, parsed)
			}
		}
	})

	t.Run("rejects an unknown cause", func(t *testing.T) {
		_, err := abortiondomain.ParseCause("Mágica")

		if !errors.Is(err, abortiondomain.ErrInvalidCause) {
			t.Fatalf("error = %v, want ErrInvalidCause", err)
		}
	})
}
