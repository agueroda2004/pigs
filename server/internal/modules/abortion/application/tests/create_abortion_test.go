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

func TestCreateAbortionExecute(t *testing.T) {
	sowID := uuid.New()
	serviceID := uuid.New()
	actorID := uuid.New()

	repositoryWith := func() *fakeAbortionRepository {
		return &fakeAbortionRepository{
			sow:     testSow(sowID, sowdomain.StatePregnant),
			service: testService(serviceID, sowID, servicedomain.StateConfirmed, testMount(serviceID, 1, mountDayOne)),
		}
	}

	t.Run("registers the abortion and aborts the sow and service", func(t *testing.T) {
		repository := repositoryWith()
		useCase := abortionapplication.NewCreateAbortionService(repository, func() time.Time { return nowReference })

		abortion, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if abortion.SowID != sowID || abortion.ServiceID != serviceID {
			t.Fatalf("unexpected references: %#v", abortion)
		}
		if abortion.CreatedBy != actorID || !abortion.CreatedAt.Equal(nowReference) {
			t.Fatalf("unexpected audit: %#v", abortion)
		}
		if repository.created != abortion {
			t.Fatalf("abortion was not persisted")
		}
		if repository.createdSow.State != sowdomain.StateAborted {
			t.Fatalf("sow state = %v, want %v", repository.createdSow.State, sowdomain.StateAborted)
		}
		if repository.createdService.State != servicedomain.StateAborted {
			t.Fatalf("service state = %v, want %v", repository.createdService.State, servicedomain.StateAborted)
		}
	})

	t.Run("uses the latest mount date for the rule", func(t *testing.T) {
		repository := repositoryWith()
		repository.service.Mounts = []*servicedomain.Mount{
			testMount(serviceID, 1, mountDayOne),
			testMount(serviceID, 2, mountDayTwo),
		}
		useCase := abortionapplication.NewCreateAbortionService(repository, func() time.Time { return nowReference })

		command := validCommand(sowID, actorID)
		command.AbortionDate = mountDayTwo

		_, err := useCase.Execute(context.Background(), command)

		if !errors.Is(err, abortiondomain.ErrAbortionDateBeforeMount) {
			t.Fatalf("error = %v, want ErrAbortionDateBeforeMount", err)
		}
	})

	t.Run("propagates a missing sow", func(t *testing.T) {
		repository := &fakeAbortionRepository{sowErr: ports.ErrSowNotFound}
		useCase := abortionapplication.NewCreateAbortionService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("rejects a sow that is not gestating", func(t *testing.T) {
		repository := repositoryWith()
		repository.sow = testSow(sowID, sowdomain.StateWeaned)
		useCase := abortionapplication.NewCreateAbortionService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, abortionapplication.ErrSowNotGestating) {
			t.Fatalf("error = %v, want ErrSowNotGestating", err)
		}
	})

	t.Run("propagates a missing service", func(t *testing.T) {
		repository := repositoryWith()
		repository.serviceErr = ports.ErrServiceNotFound
		useCase := abortionapplication.NewCreateAbortionService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, ports.ErrServiceNotFound) {
			t.Fatalf("error = %v, want ErrServiceNotFound", err)
		}
	})

	t.Run("rejects a service that is not confirmed", func(t *testing.T) {
		repository := repositoryWith()
		repository.service.State = servicedomain.StateFailed
		useCase := abortionapplication.NewCreateAbortionService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, abortionapplication.ErrServiceNotConfirmed) {
			t.Fatalf("error = %v, want ErrServiceNotConfirmed", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := repositoryWith()
		repository.createErr = errors.New("boom")
		useCase := abortionapplication.NewCreateAbortionService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, repository.createErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
