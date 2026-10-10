package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	abortionapplication "server/internal/modules/abortion/application"
	abortiondomain "server/internal/modules/abortion/domain"
	"server/internal/modules/abortion/ports"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

func testAbortion(sowID, serviceID uuid.UUID) *abortiondomain.Abortion {
	return &abortiondomain.Abortion{
		ID:           uuid.New(),
		SowID:        sowID,
		ServiceID:    serviceID,
		AbortionDate: abortionDay,
		Cause:        abortiondomain.CauseInfectious,
	}
}

func TestUpdateAbortionExecute(t *testing.T) {
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

	t.Run("applies the fields and persists the abortion", func(t *testing.T) {
		repository := repositoryWith()
		useCase := abortionapplication.NewUpdateAbortionService(repository, func() time.Time { return nowReference })
		newDate := time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC)
		newCause := abortiondomain.CauseTrauma
		note := "actualizado"

		err := useCase.Execute(context.Background(), repository.getAbortion.ID, abortionapplication.UpdateAbortionCommand{
			AbortionDate: &newDate,
			Cause:        &newCause,
			Note:         &note,
			UpdatedBy:    actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.updated == nil {
			t.Fatalf("abortion was not persisted")
		}
		if !repository.updated.AbortionDate.Equal(newDate) || repository.updated.Cause != newCause {
			t.Fatalf("unexpected fields: %#v", repository.updated)
		}
		if repository.updated.UpdatedBy != actorID || !repository.updated.UpdatedAt.Equal(nowReference) {
			t.Fatalf("unexpected audit: %#v", repository.updated)
		}
	})

	t.Run("validates the date against the service last mount", func(t *testing.T) {
		repository := repositoryWith()
		useCase := abortionapplication.NewUpdateAbortionService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), repository.getAbortion.ID, abortionapplication.UpdateAbortionCommand{
			AbortionDate: &mountDayOne,
			UpdatedBy:    actorID,
		})

		if !errors.Is(err, abortiondomain.ErrAbortionDateBeforeMount) {
			t.Fatalf("error = %v, want ErrAbortionDateBeforeMount", err)
		}
		if repository.updated != nil {
			t.Fatalf("abortion must not be persisted on validation error")
		}
	})

	t.Run("validates the date against the sow entry date", func(t *testing.T) {
		repository := repositoryWith()
		useCase := abortionapplication.NewUpdateAbortionService(repository, func() time.Time { return nowReference })
		early := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

		err := useCase.Execute(context.Background(), repository.getAbortion.ID, abortionapplication.UpdateAbortionCommand{
			AbortionDate: &early,
			UpdatedBy:    actorID,
		})

		if !errors.Is(err, abortiondomain.ErrAbortionDateBeforeEntry) {
			t.Fatalf("error = %v, want ErrAbortionDateBeforeEntry", err)
		}
	})

	t.Run("propagates a missing abortion", func(t *testing.T) {
		repository := &fakeAbortionRepository{getAbortionErr: ports.ErrAbortionNotFound}
		useCase := abortionapplication.NewUpdateAbortionService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), uuid.New(), abortionapplication.UpdateAbortionCommand{
			UpdatedBy: actorID,
		})

		if !errors.Is(err, ports.ErrAbortionNotFound) {
			t.Fatalf("error = %v, want ErrAbortionNotFound", err)
		}
	})

	t.Run("propagates a missing sow", func(t *testing.T) {
		repository := &fakeAbortionRepository{getAbortion: testAbortion(uuid.New(), uuid.New()), sowErr: ports.ErrSowNotFound}
		useCase := abortionapplication.NewUpdateAbortionService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), uuid.New(), abortionapplication.UpdateAbortionCommand{
			UpdatedBy: actorID,
		})

		if !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("propagates a missing service", func(t *testing.T) {
		repository := repositoryWith()
		repository.serviceErr = ports.ErrServiceNotFound
		useCase := abortionapplication.NewUpdateAbortionService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), repository.getAbortion.ID, abortionapplication.UpdateAbortionCommand{
			UpdatedBy: actorID,
		})

		if !errors.Is(err, ports.ErrServiceNotFound) {
			t.Fatalf("error = %v, want ErrServiceNotFound", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := repositoryWith()
		repository.updateErr = errors.New("boom")
		useCase := abortionapplication.NewUpdateAbortionService(repository, func() time.Time { return nowReference })
		cause := abortiondomain.CauseOther

		err := useCase.Execute(context.Background(), repository.getAbortion.ID, abortionapplication.UpdateAbortionCommand{
			Cause:     &cause,
			UpdatedBy: actorID,
		})

		if !errors.Is(err, repository.updateErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
