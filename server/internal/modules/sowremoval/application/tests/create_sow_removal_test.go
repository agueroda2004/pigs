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

func TestCreateSowRemovalExecute(t *testing.T) {
	sowID := uuid.New()
	serviceID := uuid.New()
	actorID := uuid.New()

	repositoryWith := func() *fakeSowRemovalRepository {
		return &fakeSowRemovalRepository{
			sow:     testSow(sowID, sowdomain.StatePregnant),
			service: testService(serviceID, sowID, servicedomain.StateConfirmed, testMount(serviceID, 1, mountDayOne)),
		}
	}

	t.Run("registers the removal, deactivates the sow and fails the service", func(t *testing.T) {
		repository := repositoryWith()
		useCase := sowremovalapplication.NewCreateSowRemovalService(repository, func() time.Time { return nowReference })

		removal, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if removal.SowID != sowID || removal.LastState != string(sowdomain.StatePregnant) {
			t.Fatalf("unexpected removal: %#v", removal)
		}
		if removal.CreatedBy != actorID || !removal.CreatedAt.Equal(nowReference) {
			t.Fatalf("unexpected audit: %#v", removal)
		}
		if repository.created != removal {
			t.Fatalf("removal was not persisted")
		}
		if repository.createdSow.State != sowdomain.StateDead {
			t.Fatalf("sow state = %v, want %v", repository.createdSow.State, sowdomain.StateDead)
		}
		if repository.createdSow.Active {
			t.Fatalf("sow must be inactive after removal")
		}
		if repository.createdService == nil || repository.createdService.State != servicedomain.StateFailed {
			t.Fatalf("service was not failed: %#v", repository.createdService)
		}
	})

	t.Run("removes a sow without a service", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			sow:        testSow(sowID, sowdomain.StateAlive),
			serviceErr: ports.ErrServiceNotFound,
		}
		useCase := sowremovalapplication.NewCreateSowRemovalService(repository, func() time.Time { return nowReference })

		removal, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if removal.LastState != string(sowdomain.StateAlive) {
			t.Fatalf("last state = %q", removal.LastState)
		}
		if repository.createdService != nil {
			t.Fatalf("expected no service update, got %#v", repository.createdService)
		}
	})

	t.Run("does not fail a service that is not confirmed", func(t *testing.T) {
		service := testService(serviceID, sowID, servicedomain.StateFinished, testMount(serviceID, 1, mountDayOne))
		repository := &fakeSowRemovalRepository{
			sow:     testSow(sowID, sowdomain.StateAlive),
			service: service,
		}
		useCase := sowremovalapplication.NewCreateSowRemovalService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.createdService != nil {
			t.Fatalf("expected no service update, got %#v", repository.createdService)
		}
		if service.State != servicedomain.StateFinished {
			t.Fatalf("service state changed to %v", service.State)
		}
	})

	t.Run("propagates a missing sow", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{sowErr: ports.ErrSowNotFound}
		useCase := sowremovalapplication.NewCreateSowRemovalService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("rejects a sow that is not removable", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{sow: testSow(sowID, sowdomain.StateDead)}
		useCase := sowremovalapplication.NewCreateSowRemovalService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, sowremovalapplication.ErrSowNotRemovable) {
			t.Fatalf("error = %v, want ErrSowNotRemovable", err)
		}
		if repository.created != nil {
			t.Fatalf("removal must not be persisted")
		}
	})

	t.Run("validates the removal date against the last mount", func(t *testing.T) {
		repository := repositoryWith()
		useCase := sowremovalapplication.NewCreateSowRemovalService(repository, func() time.Time { return nowReference })

		command := validCommand(sowID, actorID)
		command.RemovalDate = mountDayOne

		_, err := useCase.Execute(context.Background(), command)

		if !errors.Is(err, sowremovaldomain.ErrRemovalDateBeforeService) {
			t.Fatalf("error = %v, want ErrRemovalDateBeforeService", err)
		}
	})

	t.Run("validates the removal date against the last abortion", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			sow:      testSow(sowID, sowdomain.StateAborted),
			abortion: testAbortion(uuid.New(), sowID, removalDay),
		}
		useCase := sowremovalapplication.NewCreateSowRemovalService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, sowremovaldomain.ErrRemovalDateBeforeAbortion) {
			t.Fatalf("error = %v, want ErrRemovalDateBeforeAbortion", err)
		}
	})

	t.Run("accepts an aborted sow after its last abortion", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			sow:         testSow(sowID, sowdomain.StateAborted),
			serviceErr:  ports.ErrServiceNotFound,
			abortion:    testAbortion(uuid.New(), sowID, abortionDay),
			abortionErr: nil,
		}
		useCase := sowremovalapplication.NewCreateSowRemovalService(repository, func() time.Time { return nowReference })

		removal, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if removal.LastState != string(sowdomain.StateAborted) {
			t.Fatalf("last state = %q", removal.LastState)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := repositoryWith()
		repository.createErr = errors.New("boom")
		useCase := sowremovalapplication.NewCreateSowRemovalService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, repository.createErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}

func TestCreateSowRemovalLastAbortion(t *testing.T) {
	sowID := uuid.New()
	actorID := uuid.New()

	t.Run("ignores a missing abortion", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			sow:         testSow(sowID, sowdomain.StateAborted),
			serviceErr:  ports.ErrServiceNotFound,
			abortionErr: ports.ErrAbortionNotFound,
		}
		useCase := sowremovalapplication.NewCreateSowRemovalService(repository, func() time.Time { return nowReference })

		if _, err := useCase.Execute(context.Background(), validCommand(sowID, actorID)); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("propagates an abortion lookup error", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			sow:         testSow(sowID, sowdomain.StateAborted),
			serviceErr:  ports.ErrServiceNotFound,
			abortionErr: errors.New("boom"),
		}
		useCase := sowremovalapplication.NewCreateSowRemovalService(repository, func() time.Time { return nowReference })

		if _, err := useCase.Execute(context.Background(), validCommand(sowID, actorID)); err == nil {
			t.Fatalf("expected an error")
		}
	})
}
