package tests

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	servicedomain "server/internal/modules/service/domain"
)

var (
	mountDayOne  = time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	mountDayTwo  = time.Date(2026, time.January, 11, 0, 0, 0, 0, time.UTC)
	mountDayLate = time.Date(2026, time.January, 12, 0, 0, 0, 0, time.UTC)
)

func validServiceParams() servicedomain.NewServiceParams {
	location := "  Nave 1  "
	note := "  Servicio programado  "
	return servicedomain.NewServiceParams{
		ID:        uuid.New(),
		SowID:     uuid.New(),
		Note:      &note,
		Location:  &location,
		CreatedBy: uuid.New(),
	}
}

func mountAt(date time.Time) servicedomain.NewMountParams {
	return servicedomain.NewMountParams{
		ID:         uuid.New(),
		BoarID:     uuid.New(),
		OperatorID: uuid.New(),
		MountDate:  date,
	}
}

func TestNewService(t *testing.T) {
	now := time.Date(2026, time.February, 1, 3, 4, 5, 0, time.UTC)

	t.Run("creates a valid service with a single mount", func(t *testing.T) {
		params := validServiceParams()

		service, err := servicedomain.NewService(params, []servicedomain.NewMountParams{mountAt(mountDayOne)}, now)

		if err != nil {
			t.Fatalf("NewService() error = %v", err)
		}
		if service.ID != params.ID || service.SowID != params.SowID {
			t.Fatalf("unexpected identifiers: %#v", service)
		}
		if service.State != servicedomain.StateConfirmed {
			t.Fatalf("state = %v, want %v", service.State, servicedomain.StateConfirmed)
		}
		if service.Location == nil || *service.Location != "Nave 1" {
			t.Fatalf("unexpected location: %#v", service.Location)
		}
		if service.Note == nil || *service.Note != "Servicio programado" {
			t.Fatalf("unexpected note: %#v", service.Note)
		}
		if len(service.Mounts) != 1 || service.Mounts[0].ServiceID != params.ID {
			t.Fatalf("unexpected mounts: %#v", service.Mounts)
		}
		if service.Mounts[0].MountNumber != 1 {
			t.Fatalf("mount number = %d, want 1", service.Mounts[0].MountNumber)
		}
		if service.CreatedBy != params.CreatedBy || service.UpdatedBy != params.CreatedBy {
			t.Fatalf("unexpected audit: %#v", service)
		}
		if !service.CreatedAt.Equal(now) || !service.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected timestamps: %#v", service)
		}
	})

	t.Run("computes the expected farrowing date as the last mount plus 114 days", func(t *testing.T) {
		params := validServiceParams()
		mounts := []servicedomain.NewMountParams{mountAt(mountDayOne), mountAt(mountDayTwo)}

		service, err := servicedomain.NewService(params, mounts, now)

		if err != nil {
			t.Fatalf("NewService() error = %v", err)
		}
		want := mountDayTwo.AddDate(0, 0, 114)
		if service.ExpectedFarrowingDate == nil || !service.ExpectedFarrowingDate.Equal(want) {
			t.Fatalf("expected farrowing = %#v, want %v", service.ExpectedFarrowingDate, want)
		}
	})

	t.Run("numbers the mounts in the given order", func(t *testing.T) {
		params := validServiceParams()
		mounts := []servicedomain.NewMountParams{
			mountAt(mountDayOne),
			mountAt(mountDayOne),
			mountAt(mountDayTwo),
		}

		service, err := servicedomain.NewService(params, mounts, now)

		if err != nil {
			t.Fatalf("NewService() error = %v", err)
		}
		if len(service.Mounts) != 3 {
			t.Fatalf("mounts = %d, want 3", len(service.Mounts))
		}
		for index, mount := range service.Mounts {
			if mount.MountNumber != index+1 {
				t.Fatalf("mount[%d] number = %d, want %d", index, mount.MountNumber, index+1)
			}
			if mount.ID != mounts[index].ID {
				t.Fatalf("mount[%d] id = %v, want %v", index, mount.ID, mounts[index].ID)
			}
		}
	})

	t.Run("defaults mount types to artificial", func(t *testing.T) {
		params := validServiceParams()

		service, err := servicedomain.NewService(params, []servicedomain.NewMountParams{mountAt(mountDayOne)}, now)

		if err != nil || service.Mounts[0].Type != servicedomain.MountTypeArtificial {
			t.Fatalf("unexpected result: err=%v service=%#v", err, service)
		}
	})

	t.Run("allows same-day and next-day mounts", func(t *testing.T) {
		params := validServiceParams()
		mounts := []servicedomain.NewMountParams{mountAt(mountDayOne), mountAt(mountDayOne), mountAt(mountDayTwo)}

		_, err := servicedomain.NewService(params, mounts, now)

		if err != nil {
			t.Fatalf("NewService() error = %v", err)
		}
	})

	t.Run("allows nil optional fields", func(t *testing.T) {
		params := validServiceParams()
		params.Location = nil
		params.Note = nil

		service, err := servicedomain.NewService(params, []servicedomain.NewMountParams{mountAt(mountDayOne)}, now)

		if err != nil || service.Location != nil || service.Note != nil {
			t.Fatalf("unexpected result: err=%v service=%#v", err, service)
		}
	})

	t.Run("normalizes empty optional fields to nil", func(t *testing.T) {
		params := validServiceParams()
		empty := "   "
		params.Location = &empty
		params.Note = &empty

		service, err := servicedomain.NewService(params, []servicedomain.NewMountParams{mountAt(mountDayOne)}, now)

		if err != nil || service.Location != nil || service.Note != nil {
			t.Fatalf("unexpected result: err=%v service=%#v", err, service)
		}
	})

	t.Run("rejects nil id", func(t *testing.T) {
		params := validServiceParams()
		params.ID = uuid.Nil

		_, err := servicedomain.NewService(params, []servicedomain.NewMountParams{mountAt(mountDayOne)}, now)
		if !errors.Is(err, servicedomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("rejects nil sow", func(t *testing.T) {
		params := validServiceParams()
		params.SowID = uuid.Nil

		_, err := servicedomain.NewService(params, []servicedomain.NewMountParams{mountAt(mountDayOne)}, now)
		if !errors.Is(err, servicedomain.ErrInvalidSow) {
			t.Fatalf("error = %v, want ErrInvalidSow", err)
		}
	})

	t.Run("rejects over-length location and note", func(t *testing.T) {
		params := validServiceParams()
		longLocation := strings.Repeat("a", 101)
		params.Location = &longLocation

		_, err := servicedomain.NewService(params, []servicedomain.NewMountParams{mountAt(mountDayOne)}, now)
		if !errors.Is(err, servicedomain.ErrInvalidLocation) {
			t.Fatalf("error = %v, want ErrInvalidLocation", err)
		}

		params = validServiceParams()
		longNote := strings.Repeat("a", 501)
		params.Note = &longNote

		_, err = servicedomain.NewService(params, []servicedomain.NewMountParams{mountAt(mountDayOne)}, now)
		if !errors.Is(err, servicedomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
	})

	t.Run("rejects nil created by", func(t *testing.T) {
		params := validServiceParams()
		params.CreatedBy = uuid.Nil

		_, err := servicedomain.NewService(params, []servicedomain.NewMountParams{mountAt(mountDayOne)}, now)
		if !errors.Is(err, servicedomain.ErrInvalidCreatedBy) {
			t.Fatalf("error = %v, want ErrInvalidCreatedBy", err)
		}
	})

	t.Run("rejects an empty mount list", func(t *testing.T) {
		params := validServiceParams()

		_, err := servicedomain.NewService(params, nil, now)
		if !errors.Is(err, servicedomain.ErrInvalidMounts) {
			t.Fatalf("error = %v, want ErrInvalidMounts", err)
		}
	})

	t.Run("rejects more than three mounts", func(t *testing.T) {
		params := validServiceParams()
		mounts := []servicedomain.NewMountParams{
			mountAt(mountDayOne),
			mountAt(mountDayOne),
			mountAt(mountDayOne),
			mountAt(mountDayOne),
		}

		_, err := servicedomain.NewService(params, mounts, now)
		if !errors.Is(err, servicedomain.ErrInvalidMounts) {
			t.Fatalf("error = %v, want ErrInvalidMounts", err)
		}
	})

	t.Run("rejects out-of-order mount dates", func(t *testing.T) {
		params := validServiceParams()
		mounts := []servicedomain.NewMountParams{mountAt(mountDayTwo), mountAt(mountDayOne)}

		_, err := servicedomain.NewService(params, mounts, now)
		if !errors.Is(err, servicedomain.ErrMountDatesNotAscending) {
			t.Fatalf("error = %v, want ErrMountDatesNotAscending", err)
		}
	})

	t.Run("rejects a mount gap larger than 24 hours", func(t *testing.T) {
		params := validServiceParams()
		mounts := []servicedomain.NewMountParams{mountAt(mountDayOne), mountAt(mountDayLate)}

		_, err := servicedomain.NewService(params, mounts, now)
		if !errors.Is(err, servicedomain.ErrMountDateGapTooLarge) {
			t.Fatalf("error = %v, want ErrMountDateGapTooLarge", err)
		}
	})

	t.Run("rejects a future mount date", func(t *testing.T) {
		params := validServiceParams()
		today := time.Date(2026, time.January, 10, 18, 0, 0, 0, time.UTC)
		mounts := []servicedomain.NewMountParams{mountAt(mountDayTwo)}

		_, err := servicedomain.NewService(params, mounts, today)
		if !errors.Is(err, servicedomain.ErrMountDateInFuture) {
			t.Fatalf("error = %v, want ErrMountDateInFuture", err)
		}
	})

	t.Run("accepts a mount dated today", func(t *testing.T) {
		params := validServiceParams()
		today := time.Date(2026, time.January, 10, 18, 0, 0, 0, time.UTC)

		_, err := servicedomain.NewService(params, []servicedomain.NewMountParams{mountAt(mountDayOne)}, today)

		if err != nil {
			t.Fatalf("NewService() error = %v", err)
		}
	})
}

func TestParseState(t *testing.T) {
	t.Run("parses a known state", func(t *testing.T) {
		state, err := servicedomain.ParseState("  Confirmado  ")

		if err != nil || state != servicedomain.StateConfirmed {
			t.Fatalf("unexpected result: err=%v state=%v", err, state)
		}
	})

	t.Run("parses every valid state", func(t *testing.T) {
		for _, state := range []servicedomain.State{
			servicedomain.StateConfirmed,
			servicedomain.StateFailed,
			servicedomain.StateAborted,
			servicedomain.StateFinished,
		} {
			parsed, err := servicedomain.ParseState(string(state))
			if err != nil || parsed != state {
				t.Fatalf("state=%v err=%v parsed=%v", state, err, parsed)
			}
		}
	})

	t.Run("rejects an unknown state", func(t *testing.T) {
		_, err := servicedomain.ParseState("Desconocido")

		if !errors.Is(err, servicedomain.ErrInvalidState) {
			t.Fatalf("error = %v, want ErrInvalidState", err)
		}
	})
}

func validService() *servicedomain.Service {
	now := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	service, err := servicedomain.NewService(
		validServiceParams(),
		[]servicedomain.NewMountParams{mountAt(mountDayOne)},
		now,
	)
	if err != nil {
		panic(err)
	}
	return service
}

func TestServiceChangeState(t *testing.T) {
	now := time.Date(2026, time.February, 2, 3, 4, 5, 0, time.UTC)

	t.Run("applies a valid state and records the audit", func(t *testing.T) {
		service := validService()
		actor := uuid.New()

		err := service.ChangeState(servicedomain.StateAborted, actor, now)

		if err != nil || service.State != servicedomain.StateAborted {
			t.Fatalf("unexpected result: err=%v state=%v", err, service.State)
		}
		if service.UpdatedBy != actor || !service.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected audit: %#v", service)
		}
	})

	t.Run("rejects an invalid state", func(t *testing.T) {
		service := validService()

		err := service.ChangeState("Desconocido", uuid.New(), now)

		if !errors.Is(err, servicedomain.ErrInvalidState) {
			t.Fatalf("error = %v, want ErrInvalidState", err)
		}
	})

	t.Run("rejects a nil actor", func(t *testing.T) {
		service := validService()

		err := service.ChangeState(servicedomain.StateAborted, uuid.Nil, now)

		if !errors.Is(err, servicedomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})
}
