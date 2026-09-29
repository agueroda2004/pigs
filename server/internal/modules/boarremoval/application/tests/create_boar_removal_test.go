package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	boarremovalapplication "server/internal/modules/boarremoval/application"
	boarremovaldomain "server/internal/modules/boarremoval/domain"
	"server/internal/modules/boarremoval/ports"
)

func TestCreateBoarRemovalExecute(t *testing.T) {
	boarID := uuid.New()
	actorID := uuid.New()

	repositoryWith := func() *fakeBoarRemovalRepository {
		return &fakeBoarRemovalRepository{
			boar:      testBoar(boarID, boardomain.StateAlive),
			lastMount: mountDay,
		}
	}

	t.Run("registers the removal, deactivates the boar and changes its state", func(t *testing.T) {
		repository := repositoryWith()
		useCase := boarremovalapplication.NewCreateBoarRemovalService(repository, func() time.Time { return nowReference })

		removal, err := useCase.Execute(context.Background(), validCommand(boarID, actorID))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if removal.BoarID != boarID || removal.LastState != string(boardomain.StateAlive) {
			t.Fatalf("unexpected removal: %#v", removal)
		}
		if removal.CreatedBy != actorID || !removal.CreatedAt.Equal(nowReference) {
			t.Fatalf("unexpected audit: %#v", removal)
		}
		if repository.created != removal {
			t.Fatalf("removal was not persisted")
		}
		if repository.createdBoar.State != boardomain.StateDead {
			t.Fatalf("boar state = %v, want %v", repository.createdBoar.State, boardomain.StateDead)
		}
		if repository.createdBoar.Active {
			t.Fatalf("boar must be inactive after removal")
		}
	})

	t.Run("removes a boar without mounts", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{
			boar: testBoar(boarID, boardomain.StateAlive),
		}
		useCase := boarremovalapplication.NewCreateBoarRemovalService(repository, func() time.Time { return nowReference })

		removal, err := useCase.Execute(context.Background(), validCommand(boarID, actorID))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if removal.LastState != string(boardomain.StateAlive) {
			t.Fatalf("last state = %q", removal.LastState)
		}
	})

	t.Run("propagates a missing boar", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{boarErr: ports.ErrBoarNotFound}
		useCase := boarremovalapplication.NewCreateBoarRemovalService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(boarID, actorID))

		if !errors.Is(err, ports.ErrBoarNotFound) {
			t.Fatalf("error = %v, want ErrBoarNotFound", err)
		}
	})

	t.Run("rejects a boar that is not removable", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{boar: testBoar(boarID, boardomain.StateDead)}
		useCase := boarremovalapplication.NewCreateBoarRemovalService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(boarID, actorID))

		if !errors.Is(err, boarremovalapplication.ErrBoarNotRemovable) {
			t.Fatalf("error = %v, want ErrBoarNotRemovable", err)
		}
		if repository.created != nil {
			t.Fatalf("removal must not be persisted")
		}
	})

	t.Run("validates the removal date against the entry date", func(t *testing.T) {
		repository := repositoryWith()
		useCase := boarremovalapplication.NewCreateBoarRemovalService(repository, func() time.Time { return nowReference })

		command := validCommand(boarID, actorID)
		command.RemovalDate = entryDay.AddDate(0, 0, -1)

		_, err := useCase.Execute(context.Background(), command)

		if !errors.Is(err, boarremovaldomain.ErrRemovalDateBeforeEntry) {
			t.Fatalf("error = %v, want ErrRemovalDateBeforeEntry", err)
		}
	})

	t.Run("validates the removal date against the last mount", func(t *testing.T) {
		repository := repositoryWith()
		useCase := boarremovalapplication.NewCreateBoarRemovalService(repository, func() time.Time { return nowReference })

		command := validCommand(boarID, actorID)
		command.RemovalDate = mountDay.AddDate(0, 0, -1)

		_, err := useCase.Execute(context.Background(), command)

		if !errors.Is(err, boarremovaldomain.ErrRemovalDateBeforeMount) {
			t.Fatalf("error = %v, want ErrRemovalDateBeforeMount", err)
		}
	})

	t.Run("propagates a last mount lookup error", func(t *testing.T) {
		repository := repositoryWith()
		repository.lastMountErr = errors.New("boom")
		useCase := boarremovalapplication.NewCreateBoarRemovalService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(boarID, actorID))

		if !errors.Is(err, repository.lastMountErr) {
			t.Fatalf("error = %v, want the lookup error", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := repositoryWith()
		repository.createErr = errors.New("boom")
		useCase := boarremovalapplication.NewCreateBoarRemovalService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(boarID, actorID))

		if !errors.Is(err, repository.createErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
