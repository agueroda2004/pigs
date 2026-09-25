package tests

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	servicedomain "server/internal/modules/service/domain"
)

func validMountParams() servicedomain.NewMountParams {
	note := "  Monta de prueba  "
	return servicedomain.NewMountParams{
		ID:         uuid.New(),
		BoarID:     uuid.New(),
		OperatorID: uuid.New(),
		MountDate:  time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC),
		Type:       servicedomain.MountTypeNatural,
		Note:       &note,
	}
}

func TestNewMount(t *testing.T) {
	now := time.Date(2026, time.February, 1, 3, 4, 5, 0, time.UTC)
	serviceID := uuid.New()
	createdBy := uuid.New()

	t.Run("creates a valid mount", func(t *testing.T) {
		params := validMountParams()

		mount, err := servicedomain.NewMount(params, serviceID, 1, createdBy, now)

		if err != nil {
			t.Fatalf("NewMount() error = %v", err)
		}
		if mount.ID != params.ID || mount.ServiceID != serviceID {
			t.Fatalf("unexpected identifiers: %#v", mount)
		}
		if mount.BoarID != params.BoarID || mount.OperatorID != params.OperatorID {
			t.Fatalf("unexpected references: %#v", mount)
		}
		if mount.MountNumber != 1 || mount.Type != servicedomain.MountTypeNatural {
			t.Fatalf("unexpected mount: %#v", mount)
		}
		if !mount.MountDate.Equal(params.MountDate) {
			t.Fatalf("unexpected date: %#v", mount.MountDate)
		}
		if mount.Note == nil || *mount.Note != "Monta de prueba" {
			t.Fatalf("unexpected note: %#v", mount.Note)
		}
		if mount.CreatedBy != createdBy || mount.UpdatedBy != createdBy {
			t.Fatalf("unexpected audit: %#v", mount)
		}
		if !mount.CreatedAt.Equal(now) || !mount.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected timestamps: %#v", mount)
		}
	})

	t.Run("defaults an empty type to artificial", func(t *testing.T) {
		params := validMountParams()
		params.Type = ""

		mount, err := servicedomain.NewMount(params, serviceID, 2, createdBy, now)

		if err != nil || mount.Type != servicedomain.MountTypeArtificial {
			t.Fatalf("unexpected result: err=%v mount=%#v", err, mount)
		}
	})

	t.Run("allows nil note", func(t *testing.T) {
		params := validMountParams()
		params.Note = nil

		mount, err := servicedomain.NewMount(params, serviceID, 1, createdBy, now)

		if err != nil || mount.Note != nil {
			t.Fatalf("unexpected result: err=%v mount=%#v", err, mount)
		}
	})

	t.Run("normalizes an empty note to nil", func(t *testing.T) {
		params := validMountParams()
		empty := "   "
		params.Note = &empty

		mount, err := servicedomain.NewMount(params, serviceID, 1, createdBy, now)

		if err != nil || mount.Note != nil {
			t.Fatalf("unexpected result: err=%v mount=%#v", err, mount)
		}
	})

	t.Run("rejects nil id", func(t *testing.T) {
		params := validMountParams()
		params.ID = uuid.Nil

		_, err := servicedomain.NewMount(params, serviceID, 1, createdBy, now)
		if !errors.Is(err, servicedomain.ErrInvalidMountID) {
			t.Fatalf("error = %v, want ErrInvalidMountID", err)
		}
	})

	t.Run("rejects nil service", func(t *testing.T) {
		params := validMountParams()

		_, err := servicedomain.NewMount(params, uuid.Nil, 1, createdBy, now)
		if !errors.Is(err, servicedomain.ErrInvalidMountService) {
			t.Fatalf("error = %v, want ErrInvalidMountService", err)
		}
	})

	t.Run("rejects nil boar", func(t *testing.T) {
		params := validMountParams()
		params.BoarID = uuid.Nil

		_, err := servicedomain.NewMount(params, serviceID, 1, createdBy, now)
		if !errors.Is(err, servicedomain.ErrInvalidMountBoar) {
			t.Fatalf("error = %v, want ErrInvalidMountBoar", err)
		}
	})

	t.Run("rejects nil operator", func(t *testing.T) {
		params := validMountParams()
		params.OperatorID = uuid.Nil

		_, err := servicedomain.NewMount(params, serviceID, 1, createdBy, now)
		if !errors.Is(err, servicedomain.ErrInvalidMountOperator) {
			t.Fatalf("error = %v, want ErrInvalidMountOperator", err)
		}
	})

	t.Run("rejects a non positive mount number", func(t *testing.T) {
		params := validMountParams()

		_, err := servicedomain.NewMount(params, serviceID, 0, createdBy, now)
		if !errors.Is(err, servicedomain.ErrInvalidMountNumber) {
			t.Fatalf("error = %v, want ErrInvalidMountNumber", err)
		}

		_, err = servicedomain.NewMount(params, serviceID, -1, createdBy, now)
		if !errors.Is(err, servicedomain.ErrInvalidMountNumber) {
			t.Fatalf("error = %v, want ErrInvalidMountNumber", err)
		}
	})

	t.Run("rejects a zero mount date", func(t *testing.T) {
		params := validMountParams()
		params.MountDate = time.Time{}

		_, err := servicedomain.NewMount(params, serviceID, 1, createdBy, now)
		if !errors.Is(err, servicedomain.ErrInvalidMountDate) {
			t.Fatalf("error = %v, want ErrInvalidMountDate", err)
		}
	})

	t.Run("rejects an invalid type", func(t *testing.T) {
		params := validMountParams()
		params.Type = servicedomain.MountType("Desconocido")

		_, err := servicedomain.NewMount(params, serviceID, 1, createdBy, now)
		if !errors.Is(err, servicedomain.ErrInvalidMountType) {
			t.Fatalf("error = %v, want ErrInvalidMountType", err)
		}
	})

	t.Run("rejects an over-length note", func(t *testing.T) {
		params := validMountParams()
		longNote := strings.Repeat("a", 501)
		params.Note = &longNote

		_, err := servicedomain.NewMount(params, serviceID, 1, createdBy, now)
		if !errors.Is(err, servicedomain.ErrInvalidMountNote) {
			t.Fatalf("error = %v, want ErrInvalidMountNote", err)
		}
	})

	t.Run("accepts a note at exactly max length", func(t *testing.T) {
		params := validMountParams()
		note := strings.Repeat("a", 500)
		params.Note = &note

		mount, err := servicedomain.NewMount(params, serviceID, 1, createdBy, now)
		if err != nil || mount.Note == nil || len([]rune(*mount.Note)) != 500 {
			t.Fatalf("unexpected result: err=%v mount=%#v", err, mount)
		}
	})

	t.Run("rejects nil created by", func(t *testing.T) {
		params := validMountParams()

		_, err := servicedomain.NewMount(params, serviceID, 1, uuid.Nil, now)
		if !errors.Is(err, servicedomain.ErrInvalidMountCreatedBy) {
			t.Fatalf("error = %v, want ErrInvalidMountCreatedBy", err)
		}
	})
}

func TestParseMountType(t *testing.T) {
	t.Run("parses a known type", func(t *testing.T) {
		mountType, err := servicedomain.ParseMountType("  Natural  ")

		if err != nil || mountType != servicedomain.MountTypeNatural {
			t.Fatalf("unexpected result: err=%v type=%v", err, mountType)
		}
	})

	t.Run("rejects an unknown type", func(t *testing.T) {
		_, err := servicedomain.ParseMountType("Desconocido")

		if !errors.Is(err, servicedomain.ErrInvalidMountType) {
			t.Fatalf("error = %v, want ErrInvalidMountType", err)
		}
	})

	t.Run("rejects an empty type", func(t *testing.T) {
		_, err := servicedomain.ParseMountType("   ")

		if !errors.Is(err, servicedomain.ErrInvalidMountType) {
			t.Fatalf("error = %v, want ErrInvalidMountType", err)
		}
	})
}
