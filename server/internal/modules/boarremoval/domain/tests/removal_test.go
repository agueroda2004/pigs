package tests

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	boarremovaldomain "server/internal/modules/boarremoval/domain"
)

var (
	entryDay     = time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)
	lastMountDay = time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	removalDay   = time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	nowReference = time.Date(2026, time.February, 1, 3, 4, 5, 0, time.UTC)
)

func validRemovalParams() boarremovaldomain.NewBoarRemovalParams {
	note := "  Baja por enfermedad  "
	return boarremovaldomain.NewBoarRemovalParams{
		ID:          uuid.New(),
		BoarID:      uuid.New(),
		RemovalDate: removalDay,
		Type:        boarremovaldomain.TypeDeath,
		Reason:      boarremovaldomain.ReasonDisease,
		Note:        &note,
		LastState:   "Vivo",
		CreatedBy:   uuid.New(),
	}
}

func validReference() boarremovaldomain.Reference {
	return boarremovaldomain.Reference{
		EntryDate:     entryDay,
		LastMountDate: lastMountDay,
	}
}

func TestNewBoarRemoval(t *testing.T) {
	t.Run("creates a valid removal", func(t *testing.T) {
		params := validRemovalParams()

		removal, err := boarremovaldomain.NewBoarRemoval(params, validReference(), nowReference)

		if err != nil {
			t.Fatalf("NewBoarRemoval() error = %v", err)
		}
		if removal.ID != params.ID || removal.BoarID != params.BoarID {
			t.Fatalf("unexpected identifiers: %#v", removal)
		}
		if !removal.RemovalDate.Equal(removalDay) || removal.Type != boarremovaldomain.TypeDeath {
			t.Fatalf("unexpected date or type: %#v", removal)
		}
		if removal.Reason != boarremovaldomain.ReasonDisease || removal.LastState != "Vivo" {
			t.Fatalf("unexpected reason or last state: %#v", removal)
		}
		if removal.Note == nil || *removal.Note != "Baja por enfermedad" {
			t.Fatalf("unexpected note: %#v", removal.Note)
		}
		if removal.CreatedBy != params.CreatedBy || removal.UpdatedBy != params.CreatedBy {
			t.Fatalf("unexpected audit: %#v", removal)
		}
		if !removal.CreatedAt.Equal(nowReference) || !removal.UpdatedAt.Equal(nowReference) {
			t.Fatalf("unexpected timestamps: %#v", removal)
		}
	})

	t.Run("allows a nil note", func(t *testing.T) {
		params := validRemovalParams()
		params.Note = nil

		removal, err := boarremovaldomain.NewBoarRemoval(params, validReference(), nowReference)

		if err != nil {
			t.Fatalf("NewBoarRemoval() error = %v", err)
		}
		if removal.Note != nil {
			t.Fatalf("expected nil note, got %#v", removal.Note)
		}
	})

	t.Run("clears an empty note", func(t *testing.T) {
		params := validRemovalParams()
		empty := "   "
		params.Note = &empty

		removal, err := boarremovaldomain.NewBoarRemoval(params, validReference(), nowReference)

		if err != nil {
			t.Fatalf("NewBoarRemoval() error = %v", err)
		}
		if removal.Note != nil {
			t.Fatalf("expected nil note, got %#v", removal.Note)
		}
	})

	t.Run("rejects an invalid id", func(t *testing.T) {
		params := validRemovalParams()
		params.ID = uuid.Nil

		if _, err := boarremovaldomain.NewBoarRemoval(params, validReference(), nowReference); !errors.Is(err, boarremovaldomain.ErrInvalidID) {
			t.Fatalf("expected ErrInvalidID, got %v", err)
		}
	})

	t.Run("rejects an invalid boar", func(t *testing.T) {
		params := validRemovalParams()
		params.BoarID = uuid.Nil

		if _, err := boarremovaldomain.NewBoarRemoval(params, validReference(), nowReference); !errors.Is(err, boarremovaldomain.ErrInvalidBoar) {
			t.Fatalf("expected ErrInvalidBoar, got %v", err)
		}
	})

	t.Run("rejects a zero removal date", func(t *testing.T) {
		params := validRemovalParams()
		params.RemovalDate = time.Time{}

		if _, err := boarremovaldomain.NewBoarRemoval(params, validReference(), nowReference); !errors.Is(err, boarremovaldomain.ErrInvalidRemovalDate) {
			t.Fatalf("expected ErrInvalidRemovalDate, got %v", err)
		}
	})

	t.Run("rejects a non removable state", func(t *testing.T) {
		params := validRemovalParams()
		params.LastState = "Muerto"

		if _, err := boarremovaldomain.NewBoarRemoval(params, validReference(), nowReference); !errors.Is(err, boarremovaldomain.ErrBoarNotRemovable) {
			t.Fatalf("expected ErrBoarNotRemovable, got %v", err)
		}
	})

	t.Run("rejects a future removal date", func(t *testing.T) {
		params := validRemovalParams()
		params.RemovalDate = nowReference.AddDate(0, 0, 1)

		if _, err := boarremovaldomain.NewBoarRemoval(params, validReference(), nowReference); !errors.Is(err, boarremovaldomain.ErrRemovalDateInFuture) {
			t.Fatalf("expected ErrRemovalDateInFuture, got %v", err)
		}
	})

	t.Run("rejects an invalid type", func(t *testing.T) {
		params := validRemovalParams()
		params.Type = "Otro"

		if _, err := boarremovaldomain.NewBoarRemoval(params, validReference(), nowReference); !errors.Is(err, boarremovaldomain.ErrInvalidType) {
			t.Fatalf("expected ErrInvalidType, got %v", err)
		}
	})

	t.Run("rejects an invalid reason", func(t *testing.T) {
		params := validRemovalParams()
		params.Reason = "Invalido"

		if _, err := boarremovaldomain.NewBoarRemoval(params, validReference(), nowReference); !errors.Is(err, boarremovaldomain.ErrInvalidReason) {
			t.Fatalf("expected ErrInvalidReason, got %v", err)
		}
	})

	t.Run("rejects a too long note", func(t *testing.T) {
		params := validRemovalParams()
		note := strings.Repeat("a", 501)
		params.Note = &note

		if _, err := boarremovaldomain.NewBoarRemoval(params, validReference(), nowReference); !errors.Is(err, boarremovaldomain.ErrInvalidNote) {
			t.Fatalf("expected ErrInvalidNote, got %v", err)
		}
	})

	t.Run("rejects an invalid creator", func(t *testing.T) {
		params := validRemovalParams()
		params.CreatedBy = uuid.Nil

		if _, err := boarremovaldomain.NewBoarRemoval(params, validReference(), nowReference); !errors.Is(err, boarremovaldomain.ErrInvalidCreatedBy) {
			t.Fatalf("expected ErrInvalidCreatedBy, got %v", err)
		}
	})
}

func TestNewBoarRemovalDates(t *testing.T) {
	t.Run("accepts a removal on the entry date", func(t *testing.T) {
		params := validRemovalParams()
		params.RemovalDate = entryDay

		if _, err := boarremovaldomain.NewBoarRemoval(params, boarremovaldomain.Reference{EntryDate: entryDay}, nowReference); err != nil {
			t.Fatalf("NewBoarRemoval() error = %v", err)
		}
	})

	t.Run("accepts a removal on the last mount date", func(t *testing.T) {
		params := validRemovalParams()
		params.RemovalDate = lastMountDay

		if _, err := boarremovaldomain.NewBoarRemoval(params, validReference(), nowReference); err != nil {
			t.Fatalf("NewBoarRemoval() error = %v", err)
		}
	})

	t.Run("rejects a removal before the entry date", func(t *testing.T) {
		params := validRemovalParams()
		params.RemovalDate = entryDay.AddDate(0, 0, -1)

		if _, err := boarremovaldomain.NewBoarRemoval(params, validReference(), nowReference); !errors.Is(err, boarremovaldomain.ErrRemovalDateBeforeEntry) {
			t.Fatalf("expected ErrRemovalDateBeforeEntry, got %v", err)
		}
	})

	t.Run("rejects a removal before the last mount date", func(t *testing.T) {
		params := validRemovalParams()
		params.RemovalDate = lastMountDay.AddDate(0, 0, -1)

		if _, err := boarremovaldomain.NewBoarRemoval(params, validReference(), nowReference); !errors.Is(err, boarremovaldomain.ErrRemovalDateBeforeMount) {
			t.Fatalf("expected ErrRemovalDateBeforeMount, got %v", err)
		}
	})

	t.Run("skips missing references", func(t *testing.T) {
		params := validRemovalParams()
		params.RemovalDate = removalDay

		if _, err := boarremovaldomain.NewBoarRemoval(params, boarremovaldomain.Reference{}, nowReference); err != nil {
			t.Fatalf("NewBoarRemoval() error = %v", err)
		}
	})
}

func TestBoarRemovalUpdate(t *testing.T) {
	updatedBy := uuid.New()
	updatedAt := nowReference.Add(time.Hour)

	newRemoval := func(t *testing.T) *boarremovaldomain.BoarRemoval {
		t.Helper()
		removal, err := boarremovaldomain.NewBoarRemoval(validRemovalParams(), validReference(), nowReference)
		if err != nil {
			t.Fatalf("NewBoarRemoval() error = %v", err)
		}
		return removal
	}

	t.Run("updates the provided fields and records the auditor", func(t *testing.T) {
		removal := newRemoval(t)
		newDate := time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC)
		newType := boarremovaldomain.TypeSacrifice
		newReason := boarremovaldomain.ReasonOther
		note := "  actualizada  "

		err := removal.Update(boarremovaldomain.UpdateBoarRemovalParams{
			RemovalDate: &newDate,
			Type:        &newType,
			Reason:      &newReason,
			Note:        &note,
		}, validReference(), updatedBy, updatedAt)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if !removal.RemovalDate.Equal(newDate) || removal.Type != newType || removal.Reason != newReason {
			t.Fatalf("unexpected removal: %#v", removal)
		}
		if removal.Note == nil || *removal.Note != "actualizada" {
			t.Fatalf("unexpected note: %#v", removal.Note)
		}
		if !removal.UpdatedAt.Equal(updatedAt) || removal.UpdatedBy != updatedBy {
			t.Fatalf("unexpected audit: %#v", removal)
		}
	})

	t.Run("clears the note with an empty string", func(t *testing.T) {
		removal := newRemoval(t)
		empty := "   "

		if err := removal.Update(boarremovaldomain.UpdateBoarRemovalParams{Note: &empty}, validReference(), updatedBy, updatedAt); err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if removal.Note != nil {
			t.Fatalf("expected nil note, got %#v", removal.Note)
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		removal := newRemoval(t)

		if err := removal.Update(boarremovaldomain.UpdateBoarRemovalParams{}, validReference(), updatedBy, updatedAt); !errors.Is(err, boarremovaldomain.ErrInvalidUpdate) {
			t.Fatalf("expected ErrInvalidUpdate, got %v", err)
		}
	})

	t.Run("rejects an invalid auditor", func(t *testing.T) {
		removal := newRemoval(t)
		reason := boarremovaldomain.ReasonOther

		if err := removal.Update(boarremovaldomain.UpdateBoarRemovalParams{Reason: &reason}, validReference(), uuid.Nil, updatedAt); !errors.Is(err, boarremovaldomain.ErrInvalidUpdatedBy) {
			t.Fatalf("expected ErrInvalidUpdatedBy, got %v", err)
		}
	})

	t.Run("rejects a zero date", func(t *testing.T) {
		removal := newRemoval(t)
		zero := time.Time{}

		if err := removal.Update(boarremovaldomain.UpdateBoarRemovalParams{RemovalDate: &zero}, validReference(), updatedBy, updatedAt); !errors.Is(err, boarremovaldomain.ErrInvalidRemovalDate) {
			t.Fatalf("expected ErrInvalidRemovalDate, got %v", err)
		}
	})

	t.Run("rejects a date before the entry date", func(t *testing.T) {
		removal := newRemoval(t)
		beforeEntry := entryDay.AddDate(0, 0, -1)

		if err := removal.Update(boarremovaldomain.UpdateBoarRemovalParams{RemovalDate: &beforeEntry}, validReference(), updatedBy, updatedAt); !errors.Is(err, boarremovaldomain.ErrRemovalDateBeforeEntry) {
			t.Fatalf("expected ErrRemovalDateBeforeEntry, got %v", err)
		}
	})

	t.Run("rejects a date before the last mount", func(t *testing.T) {
		removal := newRemoval(t)
		beforeMount := lastMountDay.AddDate(0, 0, -1)

		if err := removal.Update(boarremovaldomain.UpdateBoarRemovalParams{RemovalDate: &beforeMount}, validReference(), updatedBy, updatedAt); !errors.Is(err, boarremovaldomain.ErrRemovalDateBeforeMount) {
			t.Fatalf("expected ErrRemovalDateBeforeMount, got %v", err)
		}
	})

	t.Run("rejects an invalid type", func(t *testing.T) {
		removal := newRemoval(t)
		invalid := boarremovaldomain.Type("Otro")

		if err := removal.Update(boarremovaldomain.UpdateBoarRemovalParams{Type: &invalid}, validReference(), updatedBy, updatedAt); !errors.Is(err, boarremovaldomain.ErrInvalidType) {
			t.Fatalf("expected ErrInvalidType, got %v", err)
		}
	})

	t.Run("rejects an invalid reason", func(t *testing.T) {
		removal := newRemoval(t)
		invalid := boarremovaldomain.Reason("Invalido")

		if err := removal.Update(boarremovaldomain.UpdateBoarRemovalParams{Reason: &invalid}, validReference(), updatedBy, updatedAt); !errors.Is(err, boarremovaldomain.ErrInvalidReason) {
			t.Fatalf("expected ErrInvalidReason, got %v", err)
		}
	})

	t.Run("rejects a too long note", func(t *testing.T) {
		removal := newRemoval(t)
		note := strings.Repeat("a", 501)

		if err := removal.Update(boarremovaldomain.UpdateBoarRemovalParams{Note: &note}, validReference(), updatedBy, updatedAt); !errors.Is(err, boarremovaldomain.ErrInvalidNote) {
			t.Fatalf("expected ErrInvalidNote, got %v", err)
		}
	})
}

func TestTypeBoarState(t *testing.T) {
	cases := map[boarremovaldomain.Type]string{
		boarremovaldomain.TypeDeath:     "Muerto",
		boarremovaldomain.TypeDiscard:   "Desecho",
		boarremovaldomain.TypeSacrifice: "Sacrificado",
	}

	for removalType, expected := range cases {
		state, err := removalType.BoarState()
		if err != nil {
			t.Fatalf("BoarState() error = %v", err)
		}
		if state != expected {
			t.Fatalf("BoarState() = %q, want %q", state, expected)
		}
	}

	if _, err := boarremovaldomain.Type("Otro").BoarState(); !errors.Is(err, boarremovaldomain.ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
}

func TestParseType(t *testing.T) {
	removalType, err := boarremovaldomain.ParseType("  Muerte  ")
	if err != nil {
		t.Fatalf("ParseType() error = %v", err)
	}
	if removalType != boarremovaldomain.TypeDeath {
		t.Fatalf("ParseType() = %q", removalType)
	}

	if _, err := boarremovaldomain.ParseType("Invalido"); !errors.Is(err, boarremovaldomain.ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
}

func TestParseReason(t *testing.T) {
	reason, err := boarremovaldomain.ParseReason("  Enfermedad  ")
	if err != nil {
		t.Fatalf("ParseReason() error = %v", err)
	}
	if reason != boarremovaldomain.ReasonDisease {
		t.Fatalf("ParseReason() = %q", reason)
	}

	if _, err := boarremovaldomain.ParseReason("Invalido"); !errors.Is(err, boarremovaldomain.ErrInvalidReason) {
		t.Fatalf("expected ErrInvalidReason, got %v", err)
	}
}
