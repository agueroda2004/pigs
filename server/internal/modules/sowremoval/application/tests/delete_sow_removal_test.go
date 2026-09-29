package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
	sowremovalapplication "server/internal/modules/sowremoval/application"
	sowremovaldomain "server/internal/modules/sowremoval/domain"
	"server/internal/modules/sowremoval/ports"
)

func removedSow(id uuid.UUID, state sowdomain.State) *sowdomain.Sow {
	sow := testSow(id, state)
	sow.Active = false
	return sow
}

func TestDeleteSowRemovalExecute(t *testing.T) {
	removalID := uuid.New()
	sowID := uuid.New()
	serviceID := uuid.New()
	actorID := uuid.New()
	deleteNow := nowReference.Add(2 * time.Hour)

	t.Run("restores the sow state and reactivates it", func(t *testing.T) {
		removal := testRemoval(removalID, sowID, sowdomain.StateAlive, sowremovaldomain.TypeDeath)
		repository := &fakeSowRemovalRepository{
			removalByID: removal,
			sow:         removedSow(sowID, sowdomain.StateDead),
			serviceErr:  ports.ErrServiceNotFound,
		}
		useCase := sowremovalapplication.NewDeleteSowRemovalService(repository, func() time.Time { return deleteNow })

		err := useCase.Execute(context.Background(), sowremovalapplication.DeleteSowRemovalCommand{
			ID:        removalID,
			DeletedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deleted != removal {
			t.Fatalf("removal was not deleted")
		}
		if repository.deletedSow == nil || repository.deletedSow.State != sowdomain.StateAlive {
			t.Fatalf("sow state = %#v", repository.deletedSow)
		}
		if !repository.deletedSow.Active {
			t.Fatalf("sow must be active after delete")
		}
		if !repository.deletedSow.UpdatedAt.Equal(deleteNow) || repository.deletedSow.UpdatedBy != actorID {
			t.Fatalf("unexpected sow audit: %#v", repository.deletedSow)
		}
		if repository.deletedService != nil {
			t.Fatalf("expected no service change, got %#v", repository.deletedService)
		}
	})

	t.Run("restores a failed service to confirmed", func(t *testing.T) {
		removal := testRemoval(removalID, sowID, sowdomain.StatePregnant, sowremovaldomain.TypeDeath)
		service := testService(serviceID, sowID, servicedomain.StateFailed, testMount(serviceID, 1, mountDayOne))
		repository := &fakeSowRemovalRepository{
			removalByID: removal,
			sow:         removedSow(sowID, sowdomain.StateDead),
			service:     service,
		}
		useCase := sowremovalapplication.NewDeleteSowRemovalService(repository, func() time.Time { return deleteNow })

		err := useCase.Execute(context.Background(), sowremovalapplication.DeleteSowRemovalCommand{
			ID:        removalID,
			DeletedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deletedSow.State != sowdomain.StatePregnant {
			t.Fatalf("sow state = %v", repository.deletedSow.State)
		}
		if repository.deletedService == nil || repository.deletedService.State != servicedomain.StateConfirmed {
			t.Fatalf("service state = %#v", repository.deletedService)
		}
	})

	t.Run("leaves a service that is not failed", func(t *testing.T) {
		removal := testRemoval(removalID, sowID, sowdomain.StateAlive, sowremovaldomain.TypeDeath)
		service := testService(serviceID, sowID, servicedomain.StateFinished, testMount(serviceID, 1, mountDayOne))
		repository := &fakeSowRemovalRepository{
			removalByID: removal,
			sow:         removedSow(sowID, sowdomain.StateDead),
			service:     service,
		}
		useCase := sowremovalapplication.NewDeleteSowRemovalService(repository, func() time.Time { return deleteNow })

		if err := useCase.Execute(context.Background(), sowremovalapplication.DeleteSowRemovalCommand{ID: removalID, DeletedBy: actorID}); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deletedService != nil {
			t.Fatalf("service must not change, got %#v", repository.deletedService)
		}
		if service.State != servicedomain.StateFinished {
			t.Fatalf("service state changed to %v", service.State)
		}
	})

	t.Run("rejects an active sow", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			removalByID: testRemoval(removalID, sowID, sowdomain.StateAlive, sowremovaldomain.TypeDeath),
			sow:         testSow(sowID, sowdomain.StateDead),
		}
		useCase := sowremovalapplication.NewDeleteSowRemovalService(repository, func() time.Time { return deleteNow })

		err := useCase.Execute(context.Background(), sowremovalapplication.DeleteSowRemovalCommand{ID: removalID, DeletedBy: actorID})

		if !errors.Is(err, sowremovalapplication.ErrSowStateMismatch) {
			t.Fatalf("error = %v, want ErrSowStateMismatch", err)
		}
		if repository.deleted != nil {
			t.Fatalf("removal must not be deleted")
		}
	})

	t.Run("rejects a sow whose state does not match the removal type", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			removalByID: testRemoval(removalID, sowID, sowdomain.StateAlive, sowremovaldomain.TypeDeath),
			sow:         removedSow(sowID, sowdomain.StateDiscarded),
		}
		useCase := sowremovalapplication.NewDeleteSowRemovalService(repository, func() time.Time { return deleteNow })

		err := useCase.Execute(context.Background(), sowremovalapplication.DeleteSowRemovalCommand{ID: removalID, DeletedBy: actorID})

		if !errors.Is(err, sowremovalapplication.ErrSowStateMismatch) {
			t.Fatalf("error = %v, want ErrSowStateMismatch", err)
		}
	})

	t.Run("propagates a missing removal", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{removalByIDErr: ports.ErrSowRemovalNotFound}
		useCase := sowremovalapplication.NewDeleteSowRemovalService(repository, func() time.Time { return deleteNow })

		err := useCase.Execute(context.Background(), sowremovalapplication.DeleteSowRemovalCommand{ID: removalID, DeletedBy: actorID})

		if !errors.Is(err, ports.ErrSowRemovalNotFound) {
			t.Fatalf("error = %v, want ErrSowRemovalNotFound", err)
		}
	})

	t.Run("propagates a missing sow", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			removalByID: testRemoval(removalID, sowID, sowdomain.StateAlive, sowremovaldomain.TypeDeath),
			sowErr:      ports.ErrSowNotFound,
		}
		useCase := sowremovalapplication.NewDeleteSowRemovalService(repository, func() time.Time { return deleteNow })

		err := useCase.Execute(context.Background(), sowremovalapplication.DeleteSowRemovalCommand{ID: removalID, DeletedBy: actorID})

		if !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("rejects an invalid last state", func(t *testing.T) {
		removal := testRemoval(removalID, sowID, sowdomain.StateAlive, sowremovaldomain.TypeDeath)
		removal.LastState = "Invalido"
		repository := &fakeSowRemovalRepository{
			removalByID: removal,
			sow:         removedSow(sowID, sowdomain.StateDead),
		}
		useCase := sowremovalapplication.NewDeleteSowRemovalService(repository, func() time.Time { return deleteNow })

		err := useCase.Execute(context.Background(), sowremovalapplication.DeleteSowRemovalCommand{ID: removalID, DeletedBy: actorID})

		if !errors.Is(err, sowdomain.ErrInvalidState) {
			t.Fatalf("error = %v, want ErrInvalidState", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			removalByID: testRemoval(removalID, sowID, sowdomain.StateAlive, sowremovaldomain.TypeDeath),
			sow:         removedSow(sowID, sowdomain.StateDead),
			serviceErr:  ports.ErrServiceNotFound,
			deleteErr:   errors.New("boom"),
		}
		useCase := sowremovalapplication.NewDeleteSowRemovalService(repository, func() time.Time { return deleteNow })

		err := useCase.Execute(context.Background(), sowremovalapplication.DeleteSowRemovalCommand{ID: removalID, DeletedBy: actorID})

		if !errors.Is(err, repository.deleteErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
