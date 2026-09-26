package tests

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	sowremovaldomain "server/internal/modules/sowremoval/domain"
)

var (
	lastMountDay    = time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	lastAbortionDay = time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC)
	weaningDay      = time.Date(2026, time.January, 8, 0, 0, 0, 0, time.UTC)
	removalDay      = time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	nowReference    = time.Date(2026, time.February, 1, 3, 4, 5, 0, time.UTC)
)

func validRemovalParams() sowremovaldomain.NewSowRemovalParams {
	note := "  Baja por enfermedad  "
	return sowremovaldomain.NewSowRemovalParams{
		ID:          uuid.New(),
		SowID:       uuid.New(),
		RemovalDate: removalDay,
		Type:        sowremovaldomain.TypeDeath,
		Reason:      sowremovaldomain.ReasonDisease,
		Note:        &note,
		LastState:   "Viva",
		CreatedBy:   uuid.New(),
	}
}

func validReference() sowremovaldomain.Reference {
	return sowremovaldomain.Reference{
		LastMountDate:    lastMountDay,
		LastAbortionDate: lastAbortionDay,
		WeaningDate:      weaningDay,
	}
}

func TestNewSowRemoval(t *testing.T) {
	t.Run("creates a valid removal", func(t *testing.T) {
		params := validRemovalParams()

		removal, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference)

		if err != nil {
			t.Fatalf("NewSowRemoval() error = %v", err)
		}
		if removal.ID != params.ID || removal.SowID != params.SowID {
			t.Fatalf("unexpected identifiers: %#v", removal)
		}
		if !removal.RemovalDate.Equal(removalDay) || removal.Type != sowremovaldomain.TypeDeath {
			t.Fatalf("unexpected date or type: %#v", removal)
		}
		if removal.Reason != sowremovaldomain.ReasonDisease || removal.LastState != "Viva" {
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

		removal, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference)

		if err != nil {
			t.Fatalf("NewSowRemoval() error = %v", err)
		}
		if removal.Note != nil {
			t.Fatalf("expected nil note, got %#v", removal.Note)
		}
	})

	t.Run("clears an empty note", func(t *testing.T) {
		params := validRemovalParams()
		empty := "   "
		params.Note = &empty

		removal, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference)

		if err != nil {
			t.Fatalf("NewSowRemoval() error = %v", err)
		}
		if removal.Note != nil {
			t.Fatalf("expected nil note, got %#v", removal.Note)
		}
	})

	t.Run("allows every removable state", func(t *testing.T) {
		states := []string{"Viva", "Destetada", "Abortada", "Gestando"}
		for _, state := range states {
			params := validRemovalParams()
			params.LastState = state

			if _, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference); err != nil {
				t.Fatalf("state %s: NewSowRemoval() error = %v", state, err)
			}
		}
	})

	t.Run("rejects an invalid id", func(t *testing.T) {
		params := validRemovalParams()
		params.ID = uuid.Nil

		if _, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference); !errors.Is(err, sowremovaldomain.ErrInvalidID) {
			t.Fatalf("expected ErrInvalidID, got %v", err)
		}
	})

	t.Run("rejects an invalid sow", func(t *testing.T) {
		params := validRemovalParams()
		params.SowID = uuid.Nil

		if _, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference); !errors.Is(err, sowremovaldomain.ErrInvalidSow) {
			t.Fatalf("expected ErrInvalidSow, got %v", err)
		}
	})

	t.Run("rejects a zero removal date", func(t *testing.T) {
		params := validRemovalParams()
		params.RemovalDate = time.Time{}

		if _, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference); !errors.Is(err, sowremovaldomain.ErrInvalidRemovalDate) {
			t.Fatalf("expected ErrInvalidRemovalDate, got %v", err)
		}
	})

	t.Run("rejects a non removable state", func(t *testing.T) {
		params := validRemovalParams()
		params.LastState = "Muerta"

		if _, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference); !errors.Is(err, sowremovaldomain.ErrSowNotRemovable) {
			t.Fatalf("expected ErrSowNotRemovable, got %v", err)
		}
	})

	t.Run("rejects a future removal date", func(t *testing.T) {
		params := validRemovalParams()
		params.RemovalDate = nowReference.AddDate(0, 0, 1)

		if _, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference); !errors.Is(err, sowremovaldomain.ErrRemovalDateInFuture) {
			t.Fatalf("expected ErrRemovalDateInFuture, got %v", err)
		}
	})

	t.Run("rejects an invalid type", func(t *testing.T) {
		params := validRemovalParams()
		params.Type = "Otro"

		if _, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference); !errors.Is(err, sowremovaldomain.ErrInvalidType) {
			t.Fatalf("expected ErrInvalidType, got %v", err)
		}
	})

	t.Run("rejects an invalid reason", func(t *testing.T) {
		params := validRemovalParams()
		params.Reason = "Invalido"

		if _, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference); !errors.Is(err, sowremovaldomain.ErrInvalidReason) {
			t.Fatalf("expected ErrInvalidReason, got %v", err)
		}
	})

	t.Run("rejects a too long note", func(t *testing.T) {
		params := validRemovalParams()
		note := strings.Repeat("a", 501)
		params.Note = &note

		if _, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference); !errors.Is(err, sowremovaldomain.ErrInvalidNote) {
			t.Fatalf("expected ErrInvalidNote, got %v", err)
		}
	})

	t.Run("rejects an invalid creator", func(t *testing.T) {
		params := validRemovalParams()
		params.CreatedBy = uuid.Nil

		if _, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference); !errors.Is(err, sowremovaldomain.ErrInvalidCreatedBy) {
			t.Fatalf("expected ErrInvalidCreatedBy, got %v", err)
		}
	})
}

func TestNewSowRemovalDatesByState(t *testing.T) {
	t.Run("alive does not require references", func(t *testing.T) {
		params := validRemovalParams()
		params.LastState = "Viva"

		if _, err := sowremovaldomain.NewSowRemoval(params, sowremovaldomain.Reference{}, nowReference); err != nil {
			t.Fatalf("NewSowRemoval() error = %v", err)
		}
	})

	t.Run("gestating requires date after last mount", func(t *testing.T) {
		params := validRemovalParams()
		params.LastState = "Gestando"
		params.RemovalDate = lastMountDay

		if _, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference); !errors.Is(err, sowremovaldomain.ErrRemovalDateBeforeService) {
			t.Fatalf("expected ErrRemovalDateBeforeService, got %v", err)
		}
	})

	t.Run("gestating requires date after last abortion", func(t *testing.T) {
		params := validRemovalParams()
		params.LastState = "Gestando"
		reference := validReference()
		reference.LastAbortionDate = removalDay

		if _, err := sowremovaldomain.NewSowRemoval(params, reference, nowReference); !errors.Is(err, sowremovaldomain.ErrRemovalDateBeforeAbortion) {
			t.Fatalf("expected ErrRemovalDateBeforeAbortion, got %v", err)
		}
	})

	t.Run("gestating accepts a date after both references", func(t *testing.T) {
		params := validRemovalParams()
		params.LastState = "Gestando"

		if _, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference); err != nil {
			t.Fatalf("NewSowRemoval() error = %v", err)
		}
	})

	t.Run("aborted requires date after last abortion", func(t *testing.T) {
		params := validRemovalParams()
		params.LastState = "Abortada"
		params.RemovalDate = lastAbortionDay

		if _, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference); !errors.Is(err, sowremovaldomain.ErrRemovalDateBeforeAbortion) {
			t.Fatalf("expected ErrRemovalDateBeforeAbortion, got %v", err)
		}
	})

	t.Run("weaned requires date after weaning when provided", func(t *testing.T) {
		params := validRemovalParams()
		params.LastState = "Destetada"
		params.RemovalDate = weaningDay

		if _, err := sowremovaldomain.NewSowRemoval(params, validReference(), nowReference); !errors.Is(err, sowremovaldomain.ErrRemovalDateBeforeWeaning) {
			t.Fatalf("expected ErrRemovalDateBeforeWeaning, got %v", err)
		}
	})
}

func TestTypeSowState(t *testing.T) {
	cases := map[sowremovaldomain.Type]string{
		sowremovaldomain.TypeDeath:     "Muerta",
		sowremovaldomain.TypeDiscard:   "Desecho",
		sowremovaldomain.TypeSacrifice: "Sacrificada",
	}

	for removalType, expected := range cases {
		state, err := removalType.SowState()
		if err != nil {
			t.Fatalf("SowState() error = %v", err)
		}
		if state != expected {
			t.Fatalf("SowState() = %q, want %q", state, expected)
		}
	}

	if _, err := sowremovaldomain.Type("Otro").SowState(); !errors.Is(err, sowremovaldomain.ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
}

func TestParseType(t *testing.T) {
	removalType, err := sowremovaldomain.ParseType("  Muerte  ")
	if err != nil {
		t.Fatalf("ParseType() error = %v", err)
	}
	if removalType != sowremovaldomain.TypeDeath {
		t.Fatalf("ParseType() = %q", removalType)
	}

	if _, err := sowremovaldomain.ParseType("Invalido"); !errors.Is(err, sowremovaldomain.ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
}

func TestParseReason(t *testing.T) {
	reason, err := sowremovaldomain.ParseReason("  Enfermedad  ")
	if err != nil {
		t.Fatalf("ParseReason() error = %v", err)
	}
	if reason != sowremovaldomain.ReasonDisease {
		t.Fatalf("ParseReason() = %q", reason)
	}

	if _, err := sowremovaldomain.ParseReason("Invalido"); !errors.Is(err, sowremovaldomain.ErrInvalidReason) {
		t.Fatalf("expected ErrInvalidReason, got %v", err)
	}
}
