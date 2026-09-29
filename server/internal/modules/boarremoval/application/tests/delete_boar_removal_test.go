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

func removedBoar(id uuid.UUID, state boardomain.State) *boardomain.Boar {
	boar := testBoar(id, state)
	boar.Active = false
	return boar
}

func TestDeleteBoarRemovalExecute(t *testing.T) {
	removalID := uuid.New()
	boarID := uuid.New()
	actorID := uuid.New()
	deleteNow := nowReference.Add(2 * time.Hour)

	t.Run("restores the boar state and reactivates it", func(t *testing.T) {
		removal := testRemoval(removalID, boarID, boarremovaldomain.TypeDeath)
		repository := &fakeBoarRemovalRepository{
			removalByID: removal,
			boar:        removedBoar(boarID, boardomain.StateDead),
		}
		useCase := boarremovalapplication.NewDeleteBoarRemovalService(repository, func() time.Time { return deleteNow })

		err := useCase.Execute(context.Background(), boarremovalapplication.DeleteBoarRemovalCommand{
			ID:        removalID,
			DeletedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deleted != removal {
			t.Fatalf("removal was not deleted")
		}
		if repository.deletedBoar == nil || repository.deletedBoar.State != boardomain.StateAlive {
			t.Fatalf("boar state = %#v", repository.deletedBoar)
		}
		if !repository.deletedBoar.Active {
			t.Fatalf("boar must be active after delete")
		}
		if !repository.deletedBoar.UpdatedAt.Equal(deleteNow) || repository.deletedBoar.UpdatedBy != actorID {
			t.Fatalf("unexpected boar audit: %#v", repository.deletedBoar)
		}
	})

	t.Run("restores a discarded boar to alive", func(t *testing.T) {
		removal := testRemoval(removalID, boarID, boarremovaldomain.TypeDiscard)
		repository := &fakeBoarRemovalRepository{
			removalByID: removal,
			boar:        removedBoar(boarID, boardomain.StateDiscarded),
		}
		useCase := boarremovalapplication.NewDeleteBoarRemovalService(repository, func() time.Time { return deleteNow })

		if err := useCase.Execute(context.Background(), boarremovalapplication.DeleteBoarRemovalCommand{ID: removalID, DeletedBy: actorID}); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deletedBoar.State != boardomain.StateAlive {
			t.Fatalf("boar state = %v", repository.deletedBoar.State)
		}
	})

	t.Run("rejects an active boar", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{
			removalByID: testRemoval(removalID, boarID, boarremovaldomain.TypeDeath),
			boar:        testBoar(boarID, boardomain.StateDead),
		}
		useCase := boarremovalapplication.NewDeleteBoarRemovalService(repository, func() time.Time { return deleteNow })

		err := useCase.Execute(context.Background(), boarremovalapplication.DeleteBoarRemovalCommand{ID: removalID, DeletedBy: actorID})

		if !errors.Is(err, boarremovalapplication.ErrBoarStateMismatch) {
			t.Fatalf("error = %v, want ErrBoarStateMismatch", err)
		}
		if repository.deleted != nil {
			t.Fatalf("removal must not be deleted")
		}
	})

	t.Run("rejects a boar whose state does not match the removal type", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{
			removalByID: testRemoval(removalID, boarID, boarremovaldomain.TypeDeath),
			boar:        removedBoar(boarID, boardomain.StateSacrificed),
		}
		useCase := boarremovalapplication.NewDeleteBoarRemovalService(repository, func() time.Time { return deleteNow })

		err := useCase.Execute(context.Background(), boarremovalapplication.DeleteBoarRemovalCommand{ID: removalID, DeletedBy: actorID})

		if !errors.Is(err, boarremovalapplication.ErrBoarStateMismatch) {
			t.Fatalf("error = %v, want ErrBoarStateMismatch", err)
		}
	})

	t.Run("propagates a missing removal", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{removalByIDErr: ports.ErrBoarRemovalNotFound}
		useCase := boarremovalapplication.NewDeleteBoarRemovalService(repository, func() time.Time { return deleteNow })

		err := useCase.Execute(context.Background(), boarremovalapplication.DeleteBoarRemovalCommand{ID: removalID, DeletedBy: actorID})

		if !errors.Is(err, ports.ErrBoarRemovalNotFound) {
			t.Fatalf("error = %v, want ErrBoarRemovalNotFound", err)
		}
	})

	t.Run("propagates a missing boar", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{
			removalByID: testRemoval(removalID, boarID, boarremovaldomain.TypeDeath),
			boarErr:     ports.ErrBoarNotFound,
		}
		useCase := boarremovalapplication.NewDeleteBoarRemovalService(repository, func() time.Time { return deleteNow })

		err := useCase.Execute(context.Background(), boarremovalapplication.DeleteBoarRemovalCommand{ID: removalID, DeletedBy: actorID})

		if !errors.Is(err, ports.ErrBoarNotFound) {
			t.Fatalf("error = %v, want ErrBoarNotFound", err)
		}
	})

	t.Run("rejects an invalid last state", func(t *testing.T) {
		removal := testRemoval(removalID, boarID, boarremovaldomain.TypeDeath)
		removal.LastState = "Invalido"
		repository := &fakeBoarRemovalRepository{
			removalByID: removal,
			boar:        removedBoar(boarID, boardomain.StateDead),
		}
		useCase := boarremovalapplication.NewDeleteBoarRemovalService(repository, func() time.Time { return deleteNow })

		err := useCase.Execute(context.Background(), boarremovalapplication.DeleteBoarRemovalCommand{ID: removalID, DeletedBy: actorID})

		if !errors.Is(err, boardomain.ErrInvalidState) {
			t.Fatalf("error = %v, want ErrInvalidState", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{
			removalByID: testRemoval(removalID, boarID, boarremovaldomain.TypeDeath),
			boar:        removedBoar(boarID, boardomain.StateDead),
			deleteErr:   errors.New("boom"),
		}
		useCase := boarremovalapplication.NewDeleteBoarRemovalService(repository, func() time.Time { return deleteNow })

		err := useCase.Execute(context.Background(), boarremovalapplication.DeleteBoarRemovalCommand{ID: removalID, DeletedBy: actorID})

		if !errors.Is(err, repository.deleteErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
