package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	servicedomain "server/internal/modules/service/domain"
)

func confirmedService(t *testing.T, mounts []servicedomain.NewMountParams) *servicedomain.Service {
	t.Helper()
	now := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	service, err := servicedomain.NewService(validServiceParams(), mounts, servicedomain.Reference{}, now)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return service
}

func createMount(date time.Time) servicedomain.NewMountParams {
	return servicedomain.NewMountParams{
		ID:         uuid.New(),
		BoarID:     uuid.New(),
		OperatorID: uuid.New(),
		MountDate:  date,
	}
}

func updateMount(id uuid.UUID, date time.Time) servicedomain.UpdateMountParams {
	return servicedomain.UpdateMountParams{
		ID:         id,
		BoarID:     uuid.New(),
		OperatorID: uuid.New(),
		MountDate:  date,
	}
}

func TestNewServiceReferences(t *testing.T) {
	now := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name      string
		reference servicedomain.Reference
		want      error
	}{
		{"entry date", servicedomain.Reference{EntryDate: mountDayTwo}, servicedomain.ErrMountBeforeEntryDate},
		{"previous service", servicedomain.Reference{PreviousMountDate: mountDayTwo}, servicedomain.ErrMountBeforePreviousService},
		{"last abortion", servicedomain.Reference{LastAbortionDate: mountDayTwo}, servicedomain.ErrMountBeforeAbortion},
	}
	for _, test := range cases {
		t.Run("rejects a mount before the "+test.name, func(t *testing.T) {
			_, err := servicedomain.NewService(
				validServiceParams(),
				[]servicedomain.NewMountParams{mountAt(mountDayOne)},
				test.reference,
				now,
			)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}

	t.Run("accepts a mount after every reference", func(t *testing.T) {
		_, err := servicedomain.NewService(
			validServiceParams(),
			[]servicedomain.NewMountParams{mountAt(mountDayTwo)},
			servicedomain.Reference{
				EntryDate:         mountDayOne,
				PreviousMountDate: mountDayOne,
				LastAbortionDate:  mountDayOne,
			},
			now,
		)
		if err != nil {
			t.Fatalf("NewService() error = %v", err)
		}
	})
}

func TestServiceEnsureEditable(t *testing.T) {
	t.Run("allows a confirmed service", func(t *testing.T) {
		service := confirmedService(t, []servicedomain.NewMountParams{mountAt(mountDayOne)})

		if err := service.EnsureEditable(); err != nil {
			t.Fatalf("EnsureEditable() error = %v", err)
		}
	})

	t.Run("blocks every non-confirmed state", func(t *testing.T) {
		for _, state := range []servicedomain.State{
			servicedomain.StateFailed,
			servicedomain.StateAborted,
			servicedomain.StateFinished,
		} {
			service := confirmedService(t, []servicedomain.NewMountParams{mountAt(mountDayOne)})
			service.State = state

			if err := service.EnsureEditable(); !errors.Is(err, servicedomain.ErrServiceNotEditable) {
				t.Fatalf("state=%v error = %v, want ErrServiceNotEditable", state, err)
			}
		}
	})

	t.Run("rejects a nil identifier", func(t *testing.T) {
		service := confirmedService(t, []servicedomain.NewMountParams{mountAt(mountDayOne)})
		service.ID = uuid.Nil

		if err := service.EnsureEditable(); !errors.Is(err, servicedomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})
}

func TestServiceUpdate(t *testing.T) {
	now := time.Date(2026, time.February, 2, 3, 4, 5, 0, time.UTC)
	actor := uuid.New()

	t.Run("applies create, update and delete operations", func(t *testing.T) {
		service := confirmedService(t, []servicedomain.NewMountParams{
			mountAt(mountDayOne),
			mountAt(mountDayTwo),
		})
		first := service.Mounts[0]
		second := service.Mounts[1]

		location := "Nave 2"
		note := "actualizado"
		err := service.Update(servicedomain.UpdateServiceParams{
			Location:       &location,
			Note:           &note,
			DeleteMountIDs: []uuid.UUID{first.ID},
			UpdateMounts:   []servicedomain.UpdateMountParams{updateMount(second.ID, mountDayTwo)},
			CreateMounts:   []servicedomain.NewMountParams{createMount(mountDayLate)},
		}, servicedomain.Reference{}, actor, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if len(service.Mounts) != 2 {
			t.Fatalf("mounts = %d, want 2", len(service.Mounts))
		}
		for index, mount := range service.Mounts {
			if mount.MountNumber != index+1 {
				t.Fatalf("mount[%d] number = %d, want %d", index, mount.MountNumber, index+1)
			}
		}
		if service.Mounts[1].MountDate != mountDayLate {
			t.Fatalf("unexpected last mount date: %v", service.Mounts[1].MountDate)
		}
		want := mountDayLate.AddDate(0, 0, 114)
		if service.ExpectedFarrowingDate == nil || !service.ExpectedFarrowingDate.Equal(want) {
			t.Fatalf("expected farrowing = %#v, want %v", service.ExpectedFarrowingDate, want)
		}
		if service.Location == nil || *service.Location != location {
			t.Fatalf("unexpected location: %#v", service.Location)
		}
		if service.Note == nil || *service.Note != note {
			t.Fatalf("unexpected note: %#v", service.Note)
		}
		if service.UpdatedBy != actor || !service.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected service audit: %#v", service)
		}
		if second.UpdatedBy != actor || !second.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected mount audit: %#v", second)
		}
	})

	t.Run("keeps location and note when omitted", func(t *testing.T) {
		service := confirmedService(t, []servicedomain.NewMountParams{mountAt(mountDayOne)})
		originalLocation := service.Location

		if err := service.Update(
			servicedomain.UpdateServiceParams{},
			servicedomain.Reference{},
			actor,
			now,
		); err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if service.Location != originalLocation {
			t.Fatalf("location changed: %#v", service.Location)
		}
	})

	t.Run("blocks a non-confirmed service", func(t *testing.T) {
		service := confirmedService(t, []servicedomain.NewMountParams{mountAt(mountDayOne)})
		service.State = servicedomain.StateFinished

		err := service.Update(servicedomain.UpdateServiceParams{}, servicedomain.Reference{}, actor, now)

		if !errors.Is(err, servicedomain.ErrServiceNotEditable) {
			t.Fatalf("error = %v, want ErrServiceNotEditable", err)
		}
	})

	t.Run("rejects unknown mount identifiers", func(t *testing.T) {
		service := confirmedService(t, []servicedomain.NewMountParams{mountAt(mountDayOne)})

		err := service.Update(servicedomain.UpdateServiceParams{
			DeleteMountIDs: []uuid.UUID{uuid.New()},
		}, servicedomain.Reference{}, actor, now)
		if !errors.Is(err, servicedomain.ErrMountNotFound) {
			t.Fatalf("delete error = %v, want ErrMountNotFound", err)
		}

		err = service.Update(servicedomain.UpdateServiceParams{
			UpdateMounts: []servicedomain.UpdateMountParams{updateMount(uuid.New(), mountDayOne)},
		}, servicedomain.Reference{}, actor, now)
		if !errors.Is(err, servicedomain.ErrMountNotFound) {
			t.Fatalf("update error = %v, want ErrMountNotFound", err)
		}
	})

	t.Run("rejects a mount before the previous service", func(t *testing.T) {
		service := confirmedService(t, []servicedomain.NewMountParams{mountAt(mountDayTwo)})

		err := service.Update(servicedomain.UpdateServiceParams{
			CreateMounts: []servicedomain.NewMountParams{createMount(mountDayOne)},
		}, servicedomain.Reference{PreviousMountDate: mountDayOne}, actor, now)

		if !errors.Is(err, servicedomain.ErrMountBeforePreviousService) {
			t.Fatalf("error = %v, want ErrMountBeforePreviousService", err)
		}
	})

	t.Run("rejects removing every mount", func(t *testing.T) {
		service := confirmedService(t, []servicedomain.NewMountParams{mountAt(mountDayOne)})

		err := service.Update(servicedomain.UpdateServiceParams{
			DeleteMountIDs: []uuid.UUID{service.Mounts[0].ID},
		}, servicedomain.Reference{}, actor, now)

		if !errors.Is(err, servicedomain.ErrInvalidMounts) {
			t.Fatalf("error = %v, want ErrInvalidMounts", err)
		}
	})

	t.Run("rejects a nil actor", func(t *testing.T) {
		service := confirmedService(t, []servicedomain.NewMountParams{mountAt(mountDayOne)})

		err := service.Update(servicedomain.UpdateServiceParams{}, servicedomain.Reference{}, uuid.Nil, now)

		if !errors.Is(err, servicedomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})
}

func TestMountUpdate(t *testing.T) {
	now := time.Date(2026, time.February, 2, 3, 4, 5, 0, time.UTC)
	actor := uuid.New()

	t.Run("applies the new fields and audit", func(t *testing.T) {
		service := confirmedService(t, []servicedomain.NewMountParams{mountAt(mountDayOne)})
		mount := service.Mounts[0]
		boarID := uuid.New()
		operatorID := uuid.New()
		note := "corregida"

		err := mount.Update(servicedomain.UpdateMountParams{
			ID:         mount.ID,
			BoarID:     boarID,
			OperatorID: operatorID,
			MountDate:  mountDayTwo,
			Type:       servicedomain.MountTypeNatural,
			Note:       &note,
		}, actor, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if mount.BoarID != boarID || mount.OperatorID != operatorID {
			t.Fatalf("unexpected references: %#v", mount)
		}
		if mount.MountDate != mountDayTwo || mount.Type != servicedomain.MountTypeNatural {
			t.Fatalf("unexpected mount: %#v", mount)
		}
		if mount.Note == nil || *mount.Note != note {
			t.Fatalf("unexpected note: %#v", mount.Note)
		}
		if mount.UpdatedBy != actor || !mount.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected audit: %#v", mount)
		}
	})

	t.Run("rejects invalid fields", func(t *testing.T) {
		service := confirmedService(t, []servicedomain.NewMountParams{mountAt(mountDayOne)})
		mount := service.Mounts[0]

		if err := mount.Update(servicedomain.UpdateMountParams{ID: mount.ID}, actor, now); !errors.Is(err, servicedomain.ErrInvalidMountBoar) {
			t.Fatalf("error = %v, want ErrInvalidMountBoar", err)
		}
	})
}
