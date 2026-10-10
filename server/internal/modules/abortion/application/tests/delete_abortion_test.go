package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	abortionapplication "server/internal/modules/abortion/application"
	"server/internal/modules/abortion/ports"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

func TestDeleteAbortionExecute(t *testing.T) {
	sowID := uuid.New()
	serviceID := uuid.New()
	actorID := uuid.New()

	repositoryWith := func() *fakeAbortionRepository {
		return &fakeAbortionRepository{
			getAbortion: testAbortion(sowID, serviceID),
			sow:         testSow(sowID, sowdomain.StateAborted),
			service:     testService(serviceID, sowID, servicedomain.StateAborted, testMount(serviceID, 1, mountDayOne)),
		}
	}

	t.Run("deletes the abortion and restores the states", func(t *testing.T) {
		repository := repositoryWith()
		useCase := abortionapplication.NewDeleteAbortionService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), abortionapplication.DeleteAbortionCommand{
			ID:        repository.getAbortion.ID,
			DeletedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deleted != repository.getAbortion {
			t.Fatalf("abortion was not deleted")
		}
		if repository.deletedSow.State != sowdomain.StatePregnant {
			t.Fatalf("sow state = %v, want %v", repository.deletedSow.State, sowdomain.StatePregnant)
		}
		if repository.deletedService.State != servicedomain.StateConfirmed {
			t.Fatalf("service state = %v, want %v", repository.deletedService.State, servicedomain.StateConfirmed)
		}
		if repository.deletedSow.UpdatedBy != actorID || !repository.deletedSow.UpdatedAt.Equal(nowReference) {
			t.Fatalf("unexpected sow audit: %#v", repository.deletedSow)
		}
		if !repository.hasFutureEventsAfter.Equal(abortionDay) {
			t.Fatalf("future events checked after %v, want %v", repository.hasFutureEventsAfter, abortionDay)
		}
	})

	t.Run("blocks the deletion when there are future events", func(t *testing.T) {
		repository := repositoryWith()
		repository.hasFutureEvents = true
		useCase := abortionapplication.NewDeleteAbortionService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), abortionapplication.DeleteAbortionCommand{
			ID:        repository.getAbortion.ID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, ports.ErrAbortionNotDeletable) {
			t.Fatalf("error = %v, want ErrAbortionNotDeletable", err)
		}
		if repository.deleted != nil {
			t.Fatalf("abortion must not be deleted")
		}
	})

	t.Run("propagates a missing abortion", func(t *testing.T) {
		repository := &fakeAbortionRepository{getAbortionErr: ports.ErrAbortionNotFound}
		useCase := abortionapplication.NewDeleteAbortionService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), abortionapplication.DeleteAbortionCommand{ID: uuid.New(), DeletedBy: actorID})

		if !errors.Is(err, ports.ErrAbortionNotFound) {
			t.Fatalf("error = %v, want ErrAbortionNotFound", err)
		}
	})

	t.Run("propagates a future events error", func(t *testing.T) {
		repository := repositoryWith()
		repository.hasFutureEventsErr = errors.New("boom")
		useCase := abortionapplication.NewDeleteAbortionService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), abortionapplication.DeleteAbortionCommand{ID: repository.getAbortion.ID, DeletedBy: actorID})

		if !errors.Is(err, repository.hasFutureEventsErr) {
			t.Fatalf("error = %v, want the repository error", err)
		}
	})

	t.Run("propagates a missing sow", func(t *testing.T) {
		repository := repositoryWith()
		repository.sowErr = ports.ErrSowNotFound
		useCase := abortionapplication.NewDeleteAbortionService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), abortionapplication.DeleteAbortionCommand{ID: repository.getAbortion.ID, DeletedBy: actorID})

		if !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("propagates a missing service", func(t *testing.T) {
		repository := repositoryWith()
		repository.serviceErr = ports.ErrServiceNotFound
		useCase := abortionapplication.NewDeleteAbortionService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), abortionapplication.DeleteAbortionCommand{ID: repository.getAbortion.ID, DeletedBy: actorID})

		if !errors.Is(err, ports.ErrServiceNotFound) {
			t.Fatalf("error = %v, want ErrServiceNotFound", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := repositoryWith()
		repository.deleteErr = errors.New("boom")
		useCase := abortionapplication.NewDeleteAbortionService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), abortionapplication.DeleteAbortionCommand{ID: repository.getAbortion.ID, DeletedBy: actorID})

		if !errors.Is(err, repository.deleteErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
